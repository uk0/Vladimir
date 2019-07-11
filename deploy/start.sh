#!/usr/bin/env bash


etcd --quota-backend-bytes=$((16*1024*1024)) --auto-compaction-retention=1
