package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/nahid/gohttp"
	"io"
	"strings"
)

var TableChan = make(chan string,100)

var start_count = 0

func main3() {
	InitHttpPusher(true)
	go GetList()
	for date := range TableChan {
		SendData(date)
	}
}

func SendData(date string) {
	var url = fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	req := gohttp.NewRequest()
	var formdata = map[string]string{}
	formdata["q"] = fmt.Sprintf("SELECT * FROM events where date = '%s'", date)
	formdata["format"] = "json"
	var headerVals = map[string]string{}
	resp, _ := req.FormData(formdata).Headers(headerVals).Post(url)
	var result, _ = resp.GetBodyAsString()
	rd := bufio.NewReader(strings.NewReader(result))
	// 有数据进行++ 没有数据进行watch on

	for {
		line, err := rd.ReadString('\n')
		start_count++;
		if err != nil || err == io.EOF {
			break
		}
		var inf map[string]interface{}
		fmt.Println(fmt.Sprintf("[ Send Line %d On TaskJob %s ]", start_count, date))
		_ = json.Unmarshal([]byte(line), &inf)

		if inf["$url_query"] != nil {
			inf["$url_query"] = toJSON(UrlFormat(inf["$url_query"].(string)))
		}
		inf["properties_uuid"] = GetUuid()
		//line = strings.Replace(line, "\f", "", -1)
		sendData := toJSON(inf)
		HttpSender.addMessage(sendData, "buried-point-mall")
	}
}

func GetList() {

	var url = fmt.Sprintf("https://shuixiangyoupin.cloud.sensorsdata.cn/api/sql/query?token=%s&project=%s", "token", "production")
	req := gohttp.NewRequest()
	//SELECT date   FROM events  group by date order by date
	var formdata = map[string]string{}
	formdata["q"] = fmt.Sprintf("SELECT date FROM events where date >= '2019-05-08' and  date <= '2019-05-09' group by date order by date")
	formdata["format"] = "csv"
	var headerVals = map[string]string{}
	resp, _ := req.FormData(formdata).Headers(headerVals).Post(url)
	var result, _ = resp.GetBodyAsString()
	rd := bufio.NewReader(strings.NewReader(result))
	for {
		line, err := rd.ReadString('\n')
		if err != nil || err == io.EOF {
			break
		}
		fmt.Println(fmt.Sprintf("[ Date = %s , Start Job  ]", line))
		TableChan <- line
	}
}
