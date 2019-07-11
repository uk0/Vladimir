package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/astaxie/beego/logs"
	"google.golang.org/grpc/grpclog"

	//"github.com/astaxie/beego/logs"
	"github.com/codegangsta/cli"
	//"github.com/grpc/grpc-go/grpclog"
	"github.com/uk0/Octopoda/Tools"
	"go.etcd.io/etcd/clientv3"
	//"os"
	"strings"
	"time"
)

const (
	CHECK = true
)

var logConfig2Arry []logConfig
var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
	endpoints      = []string{"etcd.we.cn:23791"}
	//endpoints = []string{"127.0.0.1:2379"}
)

type OffsetConfig struct {
	IP        string
	Offset    float64 `json:"Offset"`
	FilePatch string  `json:"FilePatch"`
	Key       string  `json:"Key"`
	Time      string  `json:"Time"`
	Topic     string  `json:"Topic"`
	TPS       int64   `json:"TPS"`
	QPS       int64   `json:"QPS"`
}

type logConfig struct {
	IP       string
	Topic    string `json:"topic"`
	Regexp   string `json:"regexp"`
	Type     string `json:"type"`
	LogPath  string `json:"log_path"`
	Service  string `json:"service"`
	SendRate int    `json:"send_rate"`
	Dir      bool   `json:"is_dir"`
}

type HostList struct {
	Host string `json:"host"`
}

func NewCleanCommand() cli.Command {
	return cli.Command{
		Name:  "clear",
		Usage: "delete root path / data",
		Action: func(c *cli.Context) error {
			DeleteRoot()
			return nil
		},
	}
}

func NewConfListCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "pulse on etcd config",
		Action: func(c *cli.Context) error {

			table := Tools.New([]string{"topic", "regexp", "log_path", "service", "send_rate", "IP", "type", "is_dir"})
			var data = QueryEtcdLogConfig()

			var inInterface []map[string]interface{}
			inrec, _ := json.Marshal(data)
			json.Unmarshal(inrec, &inInterface)
			// iterate through inrecs
			for _, val := range inInterface {
				table.AddRow(val)
			}
			table.Print()
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "online", Usage: "online client offset"},
		},
	}
}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

func NewAddTaskCommand() cli.Command {
	return cli.Command{
		Name:  "task",
		Usage: "addtask on services",
		Action: func(c *cli.Context) error {
			AddTaskEtcd(logConfig{
				LogPath:  c.String("f"),
				Topic:    c.String("t"),
				Service:  c.String("s"),
				Regexp:   c.String("re"),
				Type:     c.String("p"),
				SendRate: 100,
			}, c.String("host"))

			return nil
		}, Flags: []cli.Flag{
			cli.StringFlag{Name: "re", Usage: "regexp"},          // 正则表达式
			cli.StringFlag{Name: "host", Usage: "config host"},   // host
			cli.StringFlag{Name: "p", Usage: "config send type"}, // type
			cli.StringFlag{Name: "t", Usage: "config topic"},     // topic
			cli.StringFlag{Name: "f", Usage: "config log_path"},  // log_path
			cli.StringFlag{Name: "s", Usage: "config  service"},  // service
		},
	}
}

func NewHostsCommand() cli.Command {
	return cli.Command{
		Name:  "host",
		Usage: "host list",
		Action: func(c *cli.Context) error {

			cli, err := clientv3.New(clientv3.Config{
				Endpoints:   endpoints,
				DialTimeout: 5 * time.Second,
			})
			if err != nil {
				fmt.Println("connect failed, err:", err)
				return nil
			}
			defer cli.Close()
			// watch key 监听节点

			table := Tools.New([]string{"host"})

			var datas = []HostList{}
			rch, _ := cli.Get(context.Background(), "/pulse/", clientv3.WithPrefix())

			for i := 0; i < len(rch.Kvs); i++ {
				var tempStr = strings.Split(string(rch.Kvs[i].Key), "/")
				datas = append(datas, HostList{Host: tempStr[2]})
			}
			var inInterface []map[string]interface{}
			inrec, _ := json.Marshal(datas)
			json.Unmarshal(inrec, &inInterface)
			// iterate through inrecs
			for _, val := range inInterface {
				table.AddRow(val)
			}
			table.Print()

			return nil
		},
	}
}

func NewServiceListCommand() cli.Command {
	return cli.Command{
		Name:  "services",
		Usage: "pulse on services",
		Action: func(c *cli.Context) error {

			if c.String("watch") == "on" {
				for {
					time.Sleep(time.Duration(1) * time.Second)
					var data = QueryEtcdOnlineClient()
					table := Tools.New([]string{"Offset", "FilePatch", "Time", "Topic", "TPS", "QPS", "IP"})
					var inInterface []map[string]interface{}

					inrec, _ := json.Marshal(data)
					d := json.NewDecoder(bytes.NewReader(inrec))
					d.UseNumber()
					err := d.Decode(&inInterface)
					if err != nil {
						// 错误处理
					}
					//json.Unmarshal(inrec, &inInterface)
					// iterate through inrecs
					for _, val := range inInterface {
						table.AddRow(val)
					}
					table.Print()
				}
			}
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{Name: "watch", Usage: "online client offset"},
		},
	}
}

func QueryEtcdOnlineClient() []OffsetConfig {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return nil
	}
	defer cli.Close()
	// watch key 监听节点
	var dataArray = []OffsetConfig{}
	rch, _ := cli.Get(context.Background(), "/pulse", clientv3.WithPrefix())
	for i := 0; i < len(rch.Kvs); i++ {
		if strings.Contains(string(rch.Kvs[i].Key), "/health") {
			var data = OffsetConfig{}
			json.Unmarshal(rch.Kvs[i].Value, &data)
			data.IP = strings.Split(string(data.Key), "/")[2]
			//data.Offset = Decimal(data.Offset / 1000 / 1000)
			dataArray = append(dataArray, data)
		}

	}
	return dataArray
}

func DeleteRoot() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}
	defer cli.Close()
	// watch key 监听节点
	delresp, _ := cli.Delete(context.Background(), "/", clientv3.WithPrefix())
	fmt.Println(delresp)
	return
}

func QueryEtcdLogConfig() []logConfig {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return nil
	}
	defer cli.Close()
	// watch key 监听节点
	var dataArray = []logConfig{}
	rch, _ := cli.Get(context.Background(), "/pulse", clientv3.WithPrefix())
	for i := 0; i < len(rch.Kvs); i++ {
		if strings.Contains(string(rch.Kvs[i].Key), "/log_config") {
			var data = []logConfig{}
			json.Unmarshal(rch.Kvs[i].Value, &data)
			for _, datas := range data {
				datas.IP = strings.Split(string(rch.Kvs[i].Key), "/")[2]
				dataArray = append(dataArray, datas)
			}
		}

	}
	return dataArray
}

func AddTaskEtcd(saveData logConfig, taskJobIp string) {
	clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		logs.Error(err)
	}
	saveData.Dir = strings.HasSuffix(saveData.LogPath, "/")
	// 反向赋值
	// 因为etcd put操作是全部替换所以要找到以前的配置重新增加到本次的作业内。作业不会收到影响
	rch, _ := cli.Get(context.Background(), fmt.Sprintf("/pulse/%s/log_config", taskJobIp), clientv3.WithPrefix())
	var data = []logConfig{}
	for i := 0; i < len(rch.Kvs); i++ {
		fmt.Println(string(rch.Kvs[i].Value))
		json.Unmarshal(rch.Kvs[i].Value, &data)
		for _, idata := range data {
			if (idata.LogPath == "" || idata.Type == "" || idata.Topic == "") || (idata.LogPath == saveData.LogPath) {
				continue
			}
			logConfig2Arry = append(logConfig2Arry, idata)
		}
	}

	logConfig2Arry = append(logConfig2Arry, saveData)

	fmt.Println(len(logConfig2Arry))
	b, err := json.Marshal(logConfig2Arry)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = cli.Put(context.TODO(), fmt.Sprintf("/pulse/%s/log_config", taskJobIp), string(b))
	if err != nil {
		logs.Error(err)
	}
	cli.Close()
}
