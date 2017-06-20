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
  cp util.conf.template util.conf
fi


# load util.conf
source ${WORKDIR}/util.conf

# check util.conf
if [[ "${ACCESS_KEY}" == "" ]] || [[ "${SECRET_KEY}" == "" ]];then
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
        -e DOCKER_HOST=${HYPER_HOST} \
        -e ACCESS_KEY=${ACCESS_KEY} \
        -e SECRET_KEY=${SECRET_KEY} \
        -e AWS_ACCESS_KEY=${AWS_ACCESS_KEY} \
        -e AWS_SECRET_KEY=${AWS_SECRET_KEY} \
        -e URL_WITH_BASIC_AUTH=${URL_WITH_BASIC_AUTH} \
        -e MONGODB_URL=${MONGODB_URL} \
        -v $(pwd)/../:/go/src/github.com/hyperhq/hypercli \
        hyperhq/hypercli bash
    ;;
  test)
    export DOCKER_HOST=${HYPER_HOST}
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
