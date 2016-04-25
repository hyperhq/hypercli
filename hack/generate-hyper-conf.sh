#!/bin/bash

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

#show config for hyper cli
echo "Current hyper config: ~/.hyper/config.json"
echo "----------------------------------------------"
cat ~/.hyper/config.json
echo "----------------------------------------------"

#show example
cat <<EOF

Run in container(example):
  ./build-hyperserve-client.sh              # build hyper cli
  hyper -H \${DOCKER_HOST} info | grep "ID:" # get tennat id
  hyper -H \${DOCKER_HOST} pull busybox      # pull image
  hyper -H \${DOCKER_HOST} images            # list images
  cd integration-cli && go test             # start autotest


EOF

#execute command
[ $# -ne 0 ] && eval $@
