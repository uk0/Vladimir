#!/bin/sh

kill -9 `ps -ef | grep octopoda | awk '{print$2}'`