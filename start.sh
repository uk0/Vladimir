#!/bin/sh
BASE=$(cd `dirname $0`; pwd)

cd ${BASE} && nohup ./octopoda >> /dev/null &