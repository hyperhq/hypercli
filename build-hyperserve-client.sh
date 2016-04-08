#!/bin/bash

CLI_ROOT=$(dirname "{BASH_SOURCE}")
export DOCKER_CLIENTONLY=yes
cd ${CLI_ROOT}
export GOPATH=$(pwd)/vendor:$GOPATH
cd $(pwd)/hyper
go build .
