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
