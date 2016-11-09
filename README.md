mgo-statsd
==========

Small go process which polls mongodb for server status shipping as metrics to statsd


## Compiling

Make sure `golang` is installed and `GOPATH` is defined in your environment.

Then run `./build.sh`.

## Usage

The simplest form is just to run it this way and it will attempt to connect via
unauthorized fashion to a mongodb instance on localhost.

```
./mgo-statsd  -statsd_host="statsd.hostname"
```

## Docker container

Launch a container using the image on Docker Hub built from this source repo:
```
$ docker run -dit --name mgo-statsd scullxbones/mgo-statsd [optional parameters]
```

To build a local image from this repo using:
```
$ docker build -t mgo-statsd .
```

### Docker-based development stack using Docker Compose

If you have both Docker and [Docker Compose](https://docs.docker.com/compose/) installed, you can launch a complete development stack with a single command by using the provided ```docker-compose.yml``` file.

The stack defines the following containers:
* A MongoDB 3.x service, with logging muted.
* A StatsD service configured with console output backend, for debugging.
* A mgo-statsd service linked to above services, built from source. See manifest for used command-line options.

Start the stack by running:
```
$ docker-compose up
```
Stop it by ```CTRL+C```'ing it. See Docker Compose docs for help operating the stack.
