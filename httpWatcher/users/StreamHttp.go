package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/nahid/gohttp"
	"github.com/rs/xid"
	"io"
	"strings"
	"time"
)


func toJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		logs.Error(err)
	}
	return strings.Replace(string(js), ",", ", ", -1)
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

		inf["properties_uuid"] = GetUuid()

		//line = strings.Replace(line, "\f", "", -1)
		sendData := toJSON(inf)
		arrayInf = append(arrayInf, sendData)

		HttpSender.addMessage(sendData, "buried-point-user")
	}
	return arrayInf
}

func Nowdate()string  {
	time := time.Now()
	return time.String()
}
// 从第一条时间 limit 10000 条 将Time 打印出来

func GetDataStream()  {
	req := gohttp.NewRequest()
	url := fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	var formdata = map[string]string{}
	formdata["q"] = fmt.Sprintf("SELECT * FROM users")
	logs.Info(formdata["q"])
	formdata["format"] = "json"
	var headerVals = map[string]string{}
	var response,_ =req.FormData(formdata).Headers(headerVals).Post(url)
	if response.GetResp().StatusCode != 200 || response.GetResp().ContentLength == 0 {
		logs.Info("wait .....")
	} else {
		resp_str, _ := response.GetBodyAsString()
		StringToJsonArray(resp_str)
	}
}


func GetUuid() string {
	guid := xid.New()
	return fmt.Sprintf("000000000000%s", guid.String())
}

func main() {
	InitHttpPusher(true)
	GetDataStream()
}