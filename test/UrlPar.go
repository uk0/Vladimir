package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
)

func main() {

	data :=toJSON(UrlFormat("sampshare\u003d%7B%22i%22%3A%221554961425964-5109391-02769ea08d3244-28777732%22%2C%22p%22%3A%22pages%2Fvip%2FannounceList%2FannounceDetail%2FannounceDetail%22%2C%22d%22%3A1%7D\u0026contentId\u003d20\u0026contentNum\u003d20\u0026origin\u003dannounce\u0026numberN190411164552\u003d\u0026back\u003d1\u0026inviteCode\u003dq3n2sx\u0026fromType\u003dannouncedetail"))
	fmt.Println(data)
}
func UrlFormat(url2 string) map[string][]string {
	m, _ := url.ParseQuery(url2)
	return m
}


func toJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(string(js), ",", ", ",-1)
}