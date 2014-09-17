docker-hipache-updater
======================

Poor mans service router for docker.


Why?
----

I use this tool for my test server, that runs many small services. So, I'm able
to use docker and don't have to care about ports or IP addresses of my
containers.

See also: https://github.com/svenwltr/docker-dnsmasq-updater


Requirements
------------

	* docker
	* fig


Usage
-----

    fig up -d

For details see `fig.yml`.


Configuration
-------------

This program is configured by a simple JSON file. See `config.sample.json` for
an example.


Libaries
--------

 * https://github.com/fsouza/go-dockerclient
 * https://github.com/garyburd/redigo



Author
------

Sven Walter <sven@wltr.eu>
