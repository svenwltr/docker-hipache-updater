package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/garyburd/redigo/redis"
)

var (
	appConfig           AppConfig
	hostConfig          HostConfig
	dockerClient        *docker.Client
	dockerEventListener chan *docker.APIEvents
	redisConnection     redis.Conn
)

type AppConfig struct {
	DockerEndpoint string
	HostConfigPath string
	RedisAddress   string
}

type HostConfigItem struct {
	Domain    string
	Container string
	Port      int
}
type HostConfig []HostConfigItem

type Container struct {
	Name string
	IP   string
}

func main() {
	LoadAppConfig()
	LoadHostConfig()
	InitDockerClient()
	InitDockerEventListener()
	InitRedisConnection()

	UpdateHipache()
	WatchEvents()

}

func LoadAppConfig() {
	appConfig = AppConfig{}

	flag.StringVar(&appConfig.DockerEndpoint, "docker", "unix:///var/run/docker.sock", "Path to Docker socket.")
	flag.StringVar(&appConfig.HostConfigPath, "config", "config.json", "Path to host configuration.")
	flag.StringVar(&appConfig.RedisAddress, "redis", ":6379", "Redis address.")

	flag.Parse()

}

func LoadHostConfig() {
	var err error

	data, err := ioutil.ReadFile(appConfig.HostConfigPath)
	if err != nil {
		log.Fatal("Unable to open config file: ", err)
	}

	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		log.Fatal("Unable to read config file: ", err)
	}

}

func InitDockerClient() {
	var err error

	dockerClient, err = docker.NewClient(appConfig.DockerEndpoint)

	if err != nil {
		log.Fatal("Unable to initialize Docker client. ", err)
	}

}

func InitDockerEventListener() {
	var err error

	dockerEventListener = make(chan *docker.APIEvents, 10)
	err = dockerClient.AddEventListener(dockerEventListener)
	if err != nil {
		log.Fatal("Unable to initialize Docker event listener: ", err)
	}

}

func InitRedisConnection() {
	var err error
	redisConnection, err = redis.Dial("tcp", appConfig.RedisAddress)

	if err != nil {
		log.Fatal("Unable to connect to Redis: ", err)
	}

}

func WatchEvents() {
	for _ = range dockerEventListener {
		UpdateHipache()
	}

}

func UpdateHipache() {
	containers := GetRunningContainers()

	// aggregate all addresses
	data := make(map[string][]string)
	for _, item := range hostConfig {
		for _, container := range containers {
			if item.Container == container.Name {
				data[item.Domain] = append(data[item.Domain], fmt.Sprintf("http://%s:%d/", container.IP, item.Port))
			}
		}
	}

	redisConnection.Do("RENAME", "activeDomains", "oldDomains")

	// write hosts to db
	for domain, addresses := range data {
		key := "frontend:" + domain
		redisConnection.Do("MULTI")
		redisConnection.Do("DEL", key)
		redisConnection.Do("RPUSH", key, domain)
		for _, address := range addresses {
			redisConnection.Do("RPUSH", key, address)
		}
		redisConnection.Do("EXEC")

		// remember key for later clean up
		redisConnection.Do("SADD", "activeDomains", domain)

	}

	// clean up old hosts
	rmDomains, _ := redis.Strings(redisConnection.Do("SDIFF", "oldDomains", "activeDomains"))
	for _, domain := range rmDomains {
		redisConnection.Do("DEL", "frontend:"+domain)
	}

}

func GetRunningContainers() []Container {
	opts := docker.ListContainersOptions{}
	plainContainers, err := dockerClient.ListContainers(opts)
	if err != nil {
		log.Fatal("Unable to list containers: ", err)
	}

	var containers []Container

	for _, container := range plainContainers {
		inspect, _ := dockerClient.InspectContainer(container.ID)
		containers = append(containers, Container{
			inspect.Name[1:],
			inspect.NetworkSettings.IPAddress,
		})

	}

	return containers

}
