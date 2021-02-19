package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/nahid/gohttp"
	"github.com/rs/xid"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

var MiniBatch = 1024;

var outTime int64 = 100;

var start = 0;

var (
	xHttpWatcher *WatchClient
)
var WatchThreadNum = 1;

var ticker = time.NewTicker(time.Duration(outTime) * time.Second) // --- A
type MessageWatcher struct {
	start  int
	offset int
}

type WatchClient struct {
	url          string
	lineHttpChan chan *MessageWatcher
}

var chHttp = make(chan *gohttp.AsyncResponse, 100)

func NewHttpWatcher(address string) (afk *WatchClient, err error) {
	afk = &WatchClient{
		lineHttpChan: make(chan *MessageWatcher, 10),
	}
	if err != nil {
		fmt.Printf("Failed to create Connetcion: %s\n", err)
		os.Exit(1)
	}
	afk.url = address;
	for i := 0; i < WatchThreadNum; i++ {
		fmt.Println(address)
		// 根据配置文件循环开启线程去发消息到kafka
		go PusherKafka()
		go afk.Watcher()
	}

	return
}

func InitHttpWatcher() (err error) {
	var url = fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	HttpWatcher, err = NewHttpWatcher(url)
	return
}

func (k *WatchClient) Watcher() {

	//从channel中读取日志内容放到kafka消息队列中
	logs.Info("[start watcher]")
	req := gohttp.NewRequest()
	for v := range k.lineHttpChan {
		var formdata = map[string]string{}
		formdata["q"] = fmt.Sprintf("SELECT * FROM events ORDER BY time limit  %d,%d", v.start, v.offset)
		logs.Info(formdata["q"])

		formdata["format"] = "json"
		var headerVals = map[string]string{}

		req.FormData(formdata).Headers(headerVals).AsyncPost(k.url, chHttp)
	}
}

func (k *WatchClient) AddTask(s int, e int) (err error) {
	k.lineHttpChan <- &MessageWatcher{start: s, offset: e}
	return
}

func toJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		logs.Error(err)
	}
	return strings.Replace(string(js), ",", ", ", -1)
}

func PusherKafka() {
	for i := range chHttp {
		fmt.Println(i.Resp.GetBodyAsString())

		if i.Resp.GetResp().StatusCode != 200 || i.Resp.GetResp().ContentLength == 0 {
			logs.Info("wait .....")
			continue
		} else {
			var result, _ = i.Resp.GetBodyAsString()

			rd := bufio.NewReader(strings.NewReader(result))
			// 有数据进行++ 没有数据进行watch on
			start++;
			for {
				line, err := rd.ReadString('\n')
				if err != nil || err == io.EOF {
					break
				}
				var inf map[string]interface{}

				_ = json.Unmarshal([]byte(line), &inf)
				fmt.Println("---------------------------------------------")
				fmt.Println(inf)
				if inf["$url_query"] != nil {
					inf["$url_query"] = toJSON(UrlFormat(inf["$url_query"].(string)))
				}
				inf["properties_uuid"] = GetUuid()
				fmt.Println("---------------------------------------------")
				//line = strings.Replace(line, "\f", "", -1)
				sendData := toJSON(inf)
				logs.Info("[SEND DATA %s]", sendData)

				HttpSender.addMessage(sendData, "buried-point-mall")
			}
		}
	}
}

func UrlFormat(url2 string) map[string][]string {
	m, _ := url.ParseQuery(url2)
	return m
}

func GetUuid() string {
	//h := md5.New()
	//h.Write(data)
	//s := hex.EncodeToString(h.Sum(nil))
	//fmt.Println(s)
	guid := xid.New()
	return fmt.Sprintf("000000000000%s", guid.String())
}

func main2() {
	InitHttpWatcher()
	InitHttpPusher(true)
	loopWorker()
}

func loopWorker() {
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			HttpWatcher.AddTask((start*MiniBatch) + 1, MiniBatch)
		}
	}
}
