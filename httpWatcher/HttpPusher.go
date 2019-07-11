package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/nahid/gohttp"
	"os"
)

var (
	HttpSender *PushClient
)
var ThreadNum = 4;

type MessageHttp struct {
	line  string
	topic string
}

type PushClient struct {
	url          string
	lineHttpChan chan *MessageHttp
}

func NewHttpPusher(Address string) (afk *PushClient, err error) {
	afk = &PushClient{
		lineHttpChan: make(chan *MessageHttp, 10000),
		url:          Address,
	}
	if err != nil {
		fmt.Printf("Failed to create Connetcion: %s\n", err)
		os.Exit(1)
	}
	for i := 0; i < ThreadNum; i++ {
		// 根据配置文件循环开启线程去发消息到kafka
		go afk.Pusher()
	}
	return
}

func InitHttpPusher(onHost bool) (err error) {
	if onHost {
		HttpSender, err = NewHttpPusher("http://127.0.0.1:3000/api/msg")
	} else {
		HttpSender, err = NewHttpPusher("http://kafka.we.cn/rest/api/msg")
	}
	return
}
func (k *PushClient) Pusher() {
	//从channel中读取日志内容放到kafka消息队列中
	logs.Info("[start pusher]")
	req := gohttp.NewRequest()
	ch := make(chan *gohttp.AsyncResponse)
	for v := range k.lineHttpChan {
		var headerVals = map[string]string{}
		headerVals["Token"] = "bigdata.hive"
		headerVals["Topic"] = v.topic
		req.Body([]byte(v.line)).Headers(headerVals).AsyncPost(k.url, ch)

	}
	<-ch //让go route提前退出
}

func (k *PushClient) addMessage(line string, topic string) (err error) {
	k.lineHttpChan <- &MessageHttp{line: line, topic: topic}
	return
}
