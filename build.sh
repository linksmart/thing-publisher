#!/bin/bash
VERSION="1.0.0-SNAPSHOT"

ARCH=( amd64 arm )
for i in  "${ARCH[@]}"
do
  GOOS=linux GOARCH=$i go build -o ./build/linux-$i/thingpublisher -ldflags "-X main.Version=$VERSION" $1
  cp -r ./conf ./build/linux-$i/
  sed -i "s/iot.eclipse.org/localhost/g" ./build/linux-$i/conf/thing-publisher.json
  cp LICENSE ./build/linux-$i/
  mkdir ./build/linux-$i/dropzone
  mkdir ./build/linux-$i/agents
  cd ./build/linux-$i/ && tar -zcvf "ThingPublisher-linux-$i-$VERSION.tar.gz" ./
  sha256sum -b ThingPublisher-linux-$i-$VERSION.tar.gz > ThingPublisher-linux-$i-$VERSION.tar.gz.sha256
  cd ../../
done
