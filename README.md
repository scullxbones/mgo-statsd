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

Can be built from this repo using:

```
$ docker build -t mgo-statsd .
```

Or can be pulled from docker hub via:
```
$ docker pull scullxbones/mgo-statsd
```

and run (assuming an already running MongoDB) using:
```
$ docker run -dit --name mgo-statsd mgo-statsd [optional parameters]
```


