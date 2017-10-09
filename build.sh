#!/bin/bash
VERSION="1.0.0-SNAPSHOT"
GOOS=linux GOARCH=amd64 go build -o ./build/linux-amd64/thingpublisher -ldflags "-X main.Version=$VERSION" $1
cp -r ./conf ./build/linux-amd64/
sed -i "s/test.mosquitto.org/localhost/g" ./build/linux-amd64/conf/thing-publisher.json
cp LICENSE ./build/linux-amd64/
mkdir ./build/linux-amd64/dropzone
mkdir ./build/linux-amd64/agents
cd ./build/linux-amd64/ && tar -zcvf "ThingPublisher-linux-amd64-$VERSION.tar.gz" ./
cd ../../

GOOS=linux GOARCH=arm go build -o ./build/linux-arm/thingpublisher -ldflags "-X main.Version=$VERSION" $1
cp -r ./conf ./build/linux-arm/
sed -i "s/test.mosquitto.org/localhost/g" ./build/linux-arm/conf/thing-publisher.json
cp LICENSE ./build/linux-arm/
mkdir ./build/linux-arm/dropzone
mkdir ./build/linux-arm/agents
cd ./build/linux-arm/ && tar -zcvf "ThingPublisher-linux-arm-$VERSION.tar.gz" ./


