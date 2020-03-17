#!/bin/bash

if [ "${1}" == "" ]; then
  echo "Please enter docker image name"
  exit 0
fi

rm -rf ./bin
mkdir -p ./bin
docker build -t service-account:dev .
docker run --rm -d --name service-account service-account:dev sleep 10
docker cp service-account:/go/src/00pf00/service-account/bin/ServiceAccount ./bin

cp ./build/Dockerfile ./bin
cd ./bin
docker build -t "${1}" .