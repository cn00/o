#!/bin/bash
MY_DIRNAME=$(dirname $0)
cd $MY_DIRNAME
# protoc -I=. --cpp_out=. data/data.proto
# g++ -std=c++11 main.cpp data/data.pb.cc -o test ← 失敗
# clang++ -g -Wall -std=c++11 -lprotobuf main.cc addressbook.pb.cc　←サンプル
# clang++ -g -Wall -std=c++11 -lprotobuf main.cpp data/data.pb.cc
g++ -std=c++11 -lprotobuf main.cpp data/data.pb.cc -o test
# ./test change
