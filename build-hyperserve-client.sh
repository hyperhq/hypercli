#!/bin/bash

export DOCKER_CLIENTONLY=yes
dockerpath=$GOPATH/src/github.com/hyperhq/hypercli/
export GOPATH=$dockerpath/vendor:$GOPATH
cd $dockerpath/hyper
go build .
