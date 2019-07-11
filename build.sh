#!/usr/bin/env bash
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig/
BASE=$(cd `dirname $0`; pwd)
# auto build
cd $BASE
go build -race -tags static  -o  ./release/octopoda ./
cp conf.toml ./release/conf.toml
cp kill.sh ./release/kill.sh
cp start.sh ./release/start.sh
