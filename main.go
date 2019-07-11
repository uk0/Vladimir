package main

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"os"
	"time"
)

func getLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "trace":
		return logs.LevelTrace
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "error":
		return logs.LevelError
	default:
		return logs.LevelDebug
	}
}

func initLog() (err error) {
	//初始化日志库
	config := make(map[string]interface{})
	config["filename"] = appConfig.LogPath
	config["level"] = getLevel(appConfig.LogLevel)
	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println(" json.Marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func main() {
	//export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
	//export PKG_CONFIG_PATH=/usr/lib/pkgconfig/
	os.Setenv("PKG_CONFIG_PATH", "/usr/local/lib/pkgconfig/")
	err := initConfig()
	if err != nil {
		panic(fmt.Sprintf("init config failed,err:%v\n", err))
	}
	err = initLog()
	if err != nil {
		return
	}
	logs.Debug("init success")
	ipArrays, err = getLocalIP()
	logs.Info(ipArrays)
	if err != nil {
		logs.Error("get local ip failed, err:%v", err)
		return
	}
	//TODO 内网 外网区别内网kafka 外网http
	logs.Debug("get local ip succ, ips:%v", ipArrays)
	// init kafka sender
	switch appConfig.startType {
	case "kafka":
		err = InitKafkaKDC()
		logs.Info("Init Kafka Connection")
		if err != nil {
			logs.Error("init kafka failed, err:%v", err)
			return
		}
	case "http":
		// init http sender
		logs.Info("Init Http Connection")
		err = InitHttpPusher()
		if err != nil {
			logs.Error("init http failed, err:%v", err)
			return
		}
	}

	err = initEtcd(appConfig.etcdAddr, appConfig.etcdWatchKeyFmt,
		time.Duration(appConfig.etcdTimeout)*time.Millisecond)
	if err != nil {
		logs.Error("init etcd failed, err:%v", err)
		return
	}
	RunServer()
}
