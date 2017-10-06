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
go test -v $1
test_status=$?
if [ $test_status -ne 0 ]; then
    echo "command1 borked it"
    exit 1
fi
if [ -d "./agents" ]; then
  rm -rf agents
fi
if [ -d "./dropzone" ]; then
  rm -rf dropzone
fi
