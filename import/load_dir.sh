#!/bin/sh
BASE=$(cd `dirname $0`; pwd)
cd $BASE
for i in `cat import_load`; do
mkdir -p $i
rm -rf $BASE/$i/../

cp -r $GOPATH/src/$i  $BASE/$i/../
done



