package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/grpclog"
)

type logConfig2 struct {
	Topic string `json:"topic"`
	LogPath string `json:"log_path"`
	Service string `json:"service"`
	SendRate int `json:"send_rate"`
}
var logConfig2Arry []logConfig2
var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
	endpoints      = []string{"localhost:2379"}
)

func main() {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	logConfig2Arry = append(logConfig2Arry, logConfig2{Topic: "test",LogPath:"/var/log/system.log",Service:"system_logs_on_unixv7",SendRate:500})
	//logConfig2Arry = append(logConfig2Arry, logConfig2{Topic: "test",LogPath:"/var/log/install.log",Service:"system_logs_on_111",SendRate:1000})
	// make sure to close the client
	b, err := json.Marshal(logConfig2Arry)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
	_, err = cli.Put(context.TODO(), "/pulse/172.17.2.52/log_config", string(b))
	_, err = cli.Put(context.TODO(), "/pulse/172.17.2.137/log_config", string(b))
	if err != nil {
		log.Fatal(err)
	}
}