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
EOF
}

#############################################################################
WORKDIR=$(cd `dirname $0`; pwd)
cd ${WORKDIR}

#############################################################################
# ensure util.conf
if [ ! -s ${WORKDIR}/util.conf ];then
    cat > ${WORKDIR}/util.conf <<EOF
#########################################
#packet env
#########################################
#apirouter service
DOCKER_HOST=tcp://147.75.195.39:6443

#Hyper Credentials
ACCESS_KEY=
SECRET_KEY=


#########################################
#zenlayer env
#########################################
#apirouter service
#DOCKER_HOST=tcp://us-west-1.hyper.sh:443

#Hyper Credentials
#ACCESS_KEY=
#SECRET_KEY=


#########################################
##AWS Credentials
#########################################
AWS_ACCESS_KEY=
AWS_SECRET_KEY=


##For test load image from basic auth url
URL_WITH_BASIC_AUTH=http://username:password@test.xx.xx/ubuntu.tar.gz
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
    docker run -it --rm \
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli ./build-hyperserve-client.sh
    ;;
  enter)
    docker run -it --rm \
        -e DOCKER_HOST=${DOCKER_HOST} \
        -e ACCESS_KEY=${ACCESS_KEY} \
        -e SECRET_KEY=${SECRET_KEY} \
        -e AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
        -e AWS_SECRET_KEY=${AWS_SECRET_KEY} \
        -e URL_WITH_BASIC_AUTH=${URL_WITH_BASIC_AUTH} \
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli zsh
    ;;
  *) show_usage
    ;;
esac
