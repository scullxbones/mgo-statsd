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


