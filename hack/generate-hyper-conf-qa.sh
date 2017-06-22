#!/bin/bash

# set default value of DOCKER_HOST and BRANCH
if [[ "$DOCKER_HOST" == "" ]];then
  DOCKER_HOST="tcp://us-west-1.hyper.sh:443"
fi
if [[ "$BRANCH" == "" ]];then
  BRANCH="master"
fi


if [[ "$@" != "./build.sh" ]];then
    #ensure config for hyper cli
    mkdir -p ~/.hyper
    cat > ~/.hyper/config.json <<EOF
{
    "clouds": {
        "${DOCKER_HOST}": {
            "accesskey": "${ACCESS_KEY}",
            "secretkey": "${SECRET_KEY}"
        }
    }
}
EOF

env

echo "========== config git proxy =========="
if [ "${http_proxy}" != "" ];then
  git config --global http.proxy ${http_proxy}
fi
if [ "${https_proxy}" != "" ];then
  git config --global https.proxy ${https_proxy}
fi
git config --list | grep proxy

echo "========== Clone hypercli repo =========="
mkdir -p /go/src/github.com/{hyperhq,docker}
cd /go/src/github.com/hyperhq
git clone https://github.com/hyperhq/hypercli.git

echo "========== Build hypercli =========="
cd /go/src/github.com/hyperhq/hypercli
git checkout $BRANCH
if [[ $? -ne 0 ]];then
  echo "Branch $BRANCH not exist!"
  exit 1
fi
./build.sh
ln -s /go/src/github.com/hyperhq/hypercli /go/src/github.com/docker/docker
ln -s /go/src/github.com/hyperhq/hypercli/hyper/hyper /usr/bin/hyper
echo alias hypercli=\"hyper -H \${DOCKER_HOST}\" >> /root/.bashrc
source /root/.bashrc

echo "##############################################################################################"
echo "##                               Welcome to integration test env                            ##"
echo "##############################################################################################"
#show config for hyper cli
echo "Current hyper config: ~/.hyper/config.json"
echo "----------------------------------------------------------------------------------------------"
cat ~/.hyper/config.json \
  | sed 's/"secretkey":.*/"secretkey": "******************************"/g' \
  | sed 's/"auth":.*/"auth": "******************************"/g'
echo "----------------------------------------------------------------------------------------------"

fi

#execute command
if [[ $# -ne 0 ]];then
    cd /go/src/github.com/hyperhq/hypercli/integration-cli && eval $@
    if [[ "$@" == "./build.sh" ]];then
    #show make result
        if [[ $? -eq 0 ]];then
            echo "OK:)"
        else
            echo "Failed:("
        fi
    fi
fi
