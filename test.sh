#!/bin/bash
#export GOPATH=$(pwd)
echo $GOPATH
ls -la
if [ ! -d "./agents" ]; then
  mkdir agents
fi
if [ ! -d "./dropzone" ]; then
  mkdir dropzone
fi
go get code.linksmart.eu/sc/service-catalog/client
go test -v $1
test_status=$?
if [ $test_status -ne 0 ]; then
    echo "ERROR : received exitcode 1 from the test"
    exit 1
fi
if [ -d "./agents" ]; then
  rm -rf agents
fi
if [ -d "./dropzone" ]; then
  rm -rf dropzone
fi
