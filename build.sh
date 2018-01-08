#!/bin/bash
VERSION="1.0.0-SNAPSHOT"

ARCH=( amd64 arm )
PLATFORM=linux
for i in  "${ARCH[@]}"
do
  GOOS=$PLATFORM GOARCH=$i go build -o ./build/linux-$i/thingpublisher -ldflags "-X main.Version=$VERSION" $1
  cp -r ./conf ./build/linux-$i/
  sed -i "s/iot.eclipse.org/localhost/g" ./build/linux-$i/conf/thing-publisher.json
  cp LICENSE ./build/linux-$i/
  cp README.md ./build/linux-$i/
  mkdir ./build/linux-$i/dropzone
  mkdir ./build/linux-$i/agents
  cd ./build/linux-$i/ && tar -zcvf "ThingPublisher-linux-$i-$VERSION.tar.gz" ./
  sha256sum -b ThingPublisher-linux-$i-$VERSION.tar.gz > ThingPublisher-linux-$i-$VERSION.tar.gz.sha256
  SHA_OUT=$?
  if [ $SHA_OUT -eq 0 ];then
    echo "$PLATFORM/$i artifact build and check sum created, GOOD "
  else
    echo "Failed to create check sum for $PLATFORM/$i. ERROR !"
    exit 1
fi
  cd ../../
done
