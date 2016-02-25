#!/bin/bash

export DOCKER_CLIENTONLY=yes
dockerpath=$GOPATH/src/github.com/docker/docker/
export GOPATH=$dockerpath/vendor:$GOPATH
cd $dockerpath/docker
go build .
