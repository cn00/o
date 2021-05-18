#!/bin/bash
MY_DIRNAME=$(cd $(dirname $0); pwd)
cd $MY_DIRNAME
mkdir -p data
cd ../../data/
mkdir -p tmp_data
protoc -I=. --cpp_out=tmp_data data.proto
mv tmp_data/data.pb.* ${MY_DIRNAME%/}/data/
rmdir tmp_data
