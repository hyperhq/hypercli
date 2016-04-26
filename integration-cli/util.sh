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
#apirouter service
DOCKER_HOST=tcp://147.75.195.39:6443

#credentials
ACCESS_KEY=
SECRET_KEY=
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
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli zsh
    ;;
  *) show_usage
    ;;
esac
