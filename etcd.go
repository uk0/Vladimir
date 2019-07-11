package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var Client *clientv3.Client
var logConfChan chan string

// 初始化etcd
func initEtcd(addr []string, keyfmt string, timeout time.Duration) (err error) {

	var keys []string
	for _, ip := range ipArrays {
		keys = append(keys, fmt.Sprintf(keyfmt, ip))
	}

	logConfChan = make(chan string, 20)
	logs.Info("etcd watch key:%v timeout:%v", keys, timeout)

	Client, err = clientv3.New(clientv3.Config{
		Endpoints:   addr,
		DialTimeout: timeout,
	})
	if err != nil {
		logs.Error("connect failed,err:%v", err)
		return
	}
	logs.Info("init etcd success")
	waitGroup.Add(3) //TODO etcd and watchDir and save offset
	for _, key := range keys {

		CheckCluster(key)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 从etcd中获取要收集日志的信息
		resp, err := Client.Get(ctx, key)
		cancel()
		if err != nil {
			logs.Warn("get key %s failed,err:%v", key, err)
			continue
		}

		for _, ev := range resp.Kvs {
			logs.Debug("%q : %q\n", ev.Key, ev.Value)
			logs.Info("start watch Keys %s \n", ev.Key)
			logConfChan <- string(ev.Value)
		}
	}
	logs.Info("start watch Successfully %s \n",keys)
	go WatchEtcd(keys) //订阅某一个Etcd 的Key
	return
}
/***
监控日志配置信息发生变动及时修改
**/
func WatchEtcd(keys []string) {
	// 这里用于检测当需要收集的日志信息更改时及时更新
	var watchChans []clientv3.WatchChan
	for _, key := range keys {
		rch := Client.Watch(context.Background(), key)
		watchChans = append(watchChans, rch)
	}
	for {
		for _, watchC := range watchChans {
			select {
			case wresp := <-watchC:
				for _, ev := range wresp.Events {

					logs.Info(string(ev.Kv.Value))

					logs.Info("[watchChans] %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
					//fmt.Println(fmt.Sprintf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value))
					logConfChan <- string(ev.Kv.Value)
				}
			default:

			}
		}
		time.Sleep(time.Second)
	}
	waitGroup.Done()
}
/**
自动保存每个受监控文件的偏移量
**/
func LoadOffSet(Ekey string) OffsetConfig {
	localIp, _ := getLocalIP();
	resp, _ := Client.Get(context.TODO(), fmt.Sprintf("/pulse/%s/watch/%s/health", localIp[0], Ekey))

	var dat OffsetConfig
	if len(resp.Kvs) <= 0 {
		logs.Error("Get Key Result Len <= 0")
		return dat
	}
	json.Unmarshal(resp.Kvs[0].Value, &dat)
	return dat
}
/**
提供给监控文件夹服务
发现新生成文件夹自动上报到etcd
**/
func AddConfig(config logConfig) {
	localIp, _ := getLocalIP();
	var logConfig2Array []logConfig
	// 反向赋值
	// 因为etcd put操作是全部替换所以要找到以前的配置重新增加到本次的作业内。作业不会收到影响
	rch, _ := Client.Get(context.Background(), fmt.Sprintf("/pulse/%s/log_config",localIp[0]), clientv3.WithPrefix())
	var data = []logConfig{}
	for i := 0; i < len(rch.Kvs); i++ {
		fmt.Println(string(rch.Kvs[i].Value))
		json.Unmarshal(rch.Kvs[i].Value, &data)
		for _, idata := range data {
			if idata.LogPath==""||idata.Type==""||idata.Topic=="" || (idata.LogPath==config.LogPath){
				continue
			}
			logConfig2Array = append(logConfig2Array, idata)
		}
	}

	logConfig2Array = append(logConfig2Array, config)

	fmt.Println(len(logConfig2Array))
	b, err := json.Marshal(logConfig2Array)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = Client.Put(context.TODO(), fmt.Sprintf("/pulse/%s/log_config", localIp[0]), string(b))
	if err != nil {
		logs.Error(err)
	}
}

//构建通道 并且实时保存进度字节数 包括换行
func AutoSaveOffset(offset <-chan OffsetConfig) {
	for {
		var offSetConf = <-offset
		b, err := json.Marshal(offSetConf)
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = Client.Put(context.TODO(), offSetConf.Key, string(b))
		if err != nil {
			logs.Error(err)
		}
		//time.Sleep(time.Second * 1) // 简单睡一秒
	}
	waitGroup.Done()
}
// 检查当前节点是否已经上报已经上线

func CheckCluster(key string) bool {
	resp, _ := Client.Get(context.Background(), key)
	logs.Info(" keys len  = %d",len(resp.Kvs))
	if len(resp.Kvs)==1 {
		logs.Info("已经存在这个Key %s \n",key)
		return true
	}
	err, _ := Client.Put(context.Background(), key, string("[{}]"))
	if err != nil {
		logs.Info("创建失败")
		logs.Error(err)
	}
	logs.Info("Create Successfully ")
	//defer  Client.Close()
	return true
}
//将整个chan 进行返回 里面包含初次启动的配置信息
func GetLogConf() chan string {
	return logConfChan
}
