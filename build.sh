#!/usr/bin/env bash
dir_name=`dirname $0`
cd $dir_name

set -e

CGO_ENABLED=0 go build -o bin/etcd-demo ./demo/etcd
