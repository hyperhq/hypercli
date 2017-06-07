#!/bin/bash
# tool for autotest
# please run this scrip in host os

#############################################################################
function show_usage() {
    cat <<EOF
Usage: ./util.sh <action>
<action>:
    build      # build docker image 'hyperhq/hypercl' from Dockerfile.centos
    make       # make hyper cli in container
    enter      # enter container
    test       # test on host
EOF
}

function show_test_usage() {
    cat <<EOF
---------------------------------------------------------------------------------------------------------------
Usage:
  ./util.sh test all                                     # run all test case
  ./util.sh test all -timeout 20m                        # run all test case with specified timeout(default 10m)
  ./util.sh test -check.f <case prefix>                  # run specified test case
  ./util.sh test -check.f ^<case name>$                  # run specified prefix of test case
  ./util.sh test -check.f <case prefix> -timeout 20m     # combined use
----------------------------------------------------------------------------------------------------------------
EOF
}

#############################################################################
WORKDIR=$(cd `dirname $0`; pwd)
cd ${WORKDIR}

#############################################################################
# ensure util.conf
if [ ! -s ${WORKDIR}/util.conf ];then
    cat > ${WORKDIR}/util.conf <<EOF
export GOPATH=\$(pwd)/../vendor:$GOPATH
export HYPER_CONFIG=\$HOME/.hyperpkt1
export IMAGE_DIR=/tmp/image
export LOCAL_DOCKER_HOST="unix:///var/run/docker.sock"

#########################################
export DOCKER_REMOTE_DAEMON=1

#########################################
#packet env
#########################################
#apirouter service
export DOCKER_HOST=tcp://147.75.195.39:6443
#Hyper Credentials

export ACCESS_KEY="8Gxxxxxxxxxxxxxxxxx8JL"
export SECRET_KEY="Ek7xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxpOe"


#########################################
#zenlayer env
#########################################
#apirouter service
#DOCKER_HOST=tcp://us-west-1.hyper.sh:443

#Hyper Credentials
#ACCESS_KEY=
#SECRET_KEY=


#########################################
## AWS Credentials(option)
#########################################
export AWS_ACCESS_KEY=AKIxxxxxxxxxxxxxQRQ
export AWS_SECRET_KEY=UWuxxxxxxxxxxxxxxxxxxxxxxxxxxQCH


#########################################
##MONGODB(option)
#########################################
export MONGODB_URL=


##For test load image from basic auth url
export URL_WITH_BASIC_AUTH=http://xxxx:xxxxxx@test.hyper.sh/ubuntu.tar.gz
EOF

fi


# load util.conf
source ${WORKDIR}/util.conf

# check util.conf
if [[ "${DOCKER_HOST}" == "" ]] || [[ "${ACCESS_KEY}" == "" ]] || [[ "${SECRET_KEY}" == "" ]];then
    echo "please update 'ACCESS_KEY' and 'SECRET_KEY' in '${WORKDIR}/util.conf'"
    exit 1
fi

#############################################################################
# main
#############################################################################
case $1 in
  build)
    cd ${WORKDIR}/..
    docker build -t hyperhq/hypercli -f Dockerfile.centos .
    ;;
  make)
    echo "Start compile hyper client, please wait..."
    docker run -it --rm \
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli ./build.sh
    ;;
  enter)
    docker run -it --rm \
        -e DOCKER_HOST=${DOCKER_HOST} \
        -e ACCESS_KEY=${ACCESS_KEY} \
        -e SECRET_KEY=${SECRET_KEY} \
        -e AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
        -e AWS_SECRET_KEY=${AWS_SECRET_KEY} \
        -e URL_WITH_BASIC_AUTH=${URL_WITH_BASIC_AUTH} \
        -e MONGODB_URL=${MONGODB_URL} \
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli zsh
    ;;
  test)
    mkdir -p ${IMAGE_DIR} 
    shift
    if [ $# -ne 0 ];then
      if [ $1 == "all" ];then
        shift
        go test $@
      elif [ $1 == "-check.f" ];then
        go test $@
      else
        show_test_usage
      fi
    else
      show_test_usage
    fi
    ;;
  *) show_usage
    ;;
esac
