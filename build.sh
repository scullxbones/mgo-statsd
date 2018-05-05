#!/bin/sh

# Is go installed?
go_is_not_installed=`which go`
: ${go_is_not_installed:?"go binary not found. Is Golang installed?"}

# check for GOPATH and fail if not there
: ${GOPATH:?"GOPATH not defined"}

# install the deps.  should use a build tool...
dep ensure

# run go tests
go test -v ./...

# now build it
go build cmd/mgo-statsd.go
