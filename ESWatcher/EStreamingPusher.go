package main


import (
	"context"
	"fmt"
	"github.com/olivere/elastic"
	"time"
)

var host = []string{
	"http://47.96.170.1:9200/",
}

var client *elastic.Client

//初始化
func init() {
	var err error
	client, err = elastic.NewClient(elastic.SetBasicAuth("elastic","aaaa"),elastic.SetURL(host...))
	if err != nil {
		fmt.Printf("create client failed, err: %v", err)
	}
}

//ping 连接测试
func PingNode() {
	start := time.Now()

	info, code, err := client.Ping(host[0]).Do(context.Background())
	if err != nil {
		fmt.Printf("ping es failed, err: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("cost time: %v\n", duration)
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
}

//校验 index 是否存在
func IndexExists(index ...string) bool {
	exists, err := client.IndexExists(index...).Do(context.Background())
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return exists
}

//获取指定 Id 的文档
func GetDoc(index, id string) []byte {
	temp := client.Get().Index(index).Id(id)
	get, err := temp.Do(context.Background())
	if err != nil {
		panic(err)
	}
	if get.Found {
		fmt.Printf("Got document %s in version %d from index %s, type %s\n", get.Id, get.Version, get.Index, get.Type)
	}
	source, err := get.Source.MarshalJSON()
	if err != nil {
		fmt.Printf("byte convert string failed, err: %v", err)
	}
	return source
}

//term 查询
func TermQuery(index, type_, fieldName, fieldValue string) *elastic.SearchResult {
	query := elastic.NewTermQuery(fieldName, fieldValue)
	//_ = elastic.NewQueryStringQuery(fieldValue) //关键字查询

	searchResult, err := client.Search().
		Index(index).Type(type_).
		Query(query).
		From(0).Size(10).
		Pretty(true).
		Do(context.Background())

	if err != nil {
		panic(err)
	}
	fmt.Printf("query cost %d millisecond.\n", searchResult.TookInMillis)

	return searchResult
}

func Search(index, type_ string) *elastic.SearchResult {
	boolQuery := elastic.NewBoolQuery()
	searchResult, err := client.Search(index).
		Type(type_).Query(boolQuery).Pretty(true).Do(context.Background())
	if err != nil {
		panic(err)
	}

	return searchResult
}

func main() {
	PingNode()
}

