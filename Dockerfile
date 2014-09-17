FROM centos
MAINTAINER Sven Walter <sven.walter@wltr.eu>

RUN yum -y install epel-release
RUN yum -y update

RUN mkdir /app
ADD docker-hipache-updater /app/docker-hipache-updater

VOLUME /app/docker.sock
VOLUME /app/updater.json

CMD ["/app/docker-hipache-updater", \
	"--config", "/app/updater.json", \
	"--docker", "unix:///app/docker.sock", \
	"--redis", "redis:6379" \
]
