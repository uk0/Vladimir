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
	"strings"
	"time"
)

var MiniBatch = 1024;


func toJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		logs.Error(err)
	}
	return strings.Replace(string(js), ",", ", ", -1)
}
//取到今天第一条时间

func GetNowData() string{
	req := gohttp.NewRequest()
	url := fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	var formdata = map[string]string{}
	formdata["q"] = fmt.Sprintf("SELECT time FROM events where date='%s' ORDER BY time limit 1", Nowdate())
	logs.Info(formdata["q"])

	formdata["format"] = "json"
	var headerVals = map[string]string{}
	var response,_ =req.FormData(formdata).Headers(headerVals).Post(url)
	var data = map[string]string{}
	resp_str,_ :=response.GetBodyAsString()
	json.Unmarshal([]byte(resp_str),&data)
	return data["time"]
}


func StringToJsonArray(result string) []string {
	count:=0
	var arrayInf []string
	rd := bufio.NewReader(strings.NewReader(result))
	// 有数据进行++ 没有数据进行watch on
	for {
		count++;
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		var inf map[string]interface{}

		_ = json.Unmarshal([]byte(line), &inf)
		if inf["$url_query"] != nil {
			inf["$url_query"] = toJSON(UrlFormat(inf["$url_query"].(string)))
		}
		inf["properties_uuid"] = GetUuid()
		//line = strings.Replace(line, "\f", "", -1)
		sendData := toJSON(inf)
		arrayInf = append(arrayInf, sendData)

		HttpSender.addMessage(sendData, "buried-point-mall")
	}
	return arrayInf
}

func Nowdate()string  {
	t := "2006-01-02"
	time := time.Now()
	return time.Format(t)
}
// 从第一条时间 limit 10000 条 将Time 打印出来

func GetDataStream(time string)  {
	req := gohttp.NewRequest()
	url := fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	var formdata = map[string]string{}
	formdata["q"] = fmt.Sprintf("SELECT * FROM events where time >= '%s' ORDER BY time  limit  0,%d",time,MiniBatch)
	logs.Info(formdata["q"])
	formdata["format"] = "json"
	var headerVals = map[string]string{}
	var response,_ =req.FormData(formdata).Headers(headerVals).Post(url)
	if response.GetResp().StatusCode != 200 || response.GetResp().ContentLength == 0 {
		logs.Info("wait .....")
		GetDataStream(time) //如果出现问题开始递归
	} else {
		resp_str,_ :=response.GetBodyAsString()
		var data = StringToJsonArray(resp_str)
		fmt.Println(len(data))
		var formatData = map[string]string{}
		json.Unmarshal([]byte(data[len(data)-1]),&formatData)
		fmt.Println(formatData["time"])
		GetDataStream(formatData["time"])
	}
}

//  取当前表最后一个时间查看 是否等于

func UrlFormat(url2 string) map[string][]string {
	m, _ := url.ParseQuery(url2)
	return m
}

func GetUuid() string {
	guid := xid.New()
	return fmt.Sprintf("000000000000%s", guid.String())
}

func main() {
	InitHttpPusher(true)
	GetDataStream(GetNowData())
}