#!/bin/bash

# strip debugging info from binary 
GO_LDFLAGS="-w -s"

while getopts ":d" opt; do
	case ${opt} in
	d)
		GO_LDFLAGS=""
		;;
	\?)
		echo "Usage $0 [flags]
	-h  Print Help
	-d  Enable debugging symbols in binary"
		exit
		;;
	esac
done

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
CGO_ENABLED=0 go build -ldflags="${GO_LDFLAGS}" cmd/mgo-statsd.go
