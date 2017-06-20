#!/bin/sh

# Is go installed?
go_is_not_installed=`which go`
: ${go_is_not_installed:?"go binary not found. Is Golang installed?"}

# check for GOPATH and fail if not there
: ${GOPATH:?"GOPATH not defined"}

# install the deps.  should use a build tool...
go get github.com/cactus/go-statsd-client/statsd
go get github.com/vharitonsky/iniflags
go get gopkg.in/mgo.v2
go get github.com/kr/pretty

# now build it
go build
