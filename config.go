package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/uk0/Octopoda/Tools"
	"strings"
)

type AppConfig struct {
	LogPath         string
	LogLevel        string
	kafkaAddr       string
	KafkaThreadNum  int
	etcdAddr        []string
	etcdWatchKeyFmt string
	etcdHealthKey   string
	etcdTimeout     int
	startType      string
}

var appConfig = &AppConfig{}

func initConfig() (err error) {
	config := Tools.GetConfig()
	if err != nil {
		return
	}

	logPath := config.LogPath
	if err != nil || len(logPath) == 0 {
		return
	}
	logLevel := config.LogLevel
	if err != nil || len(logLevel) == 0 {
		return
	}
	kafkaAddr := config.KafkaAddress
	if err != nil || len(kafkaAddr) == 0 {
		return
	}
	logFiles := strings.Split(config.LogFiles, ",")
	if err != nil || len(logFiles) == 0 {
		return
	}

	arr := strings.Split(config.EtcdAddress, ",")
	for _, v := range arr {
		str := strings.TrimSpace(v)
		if len(str) == 0 {
			continue
		}
		appConfig.etcdAddr = append(appConfig.etcdAddr, str)
	}

	etcdHKey := config.EtcdHealthKey
	etcdKey := config.EtcdWatchKey
	if err != nil || len(etcdKey) == 0 {
		logs.Warn("get etcd watch key failed, err:%v", err)
		return
	}

	appConfig.etcdTimeout = config.EtcdTimeout
	appConfig.KafkaThreadNum = config.KafkaThread

	appConfig.etcdWatchKeyFmt = etcdKey
	appConfig.etcdHealthKey = etcdHKey
	appConfig.kafkaAddr = kafkaAddr
	appConfig.LogLevel = logLevel
	appConfig.LogPath = logPath
	appConfig.startType = config.StartType

	fmt.Printf("load config succ, data:%v\n", appConfig)
	return
}
