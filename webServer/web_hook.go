package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nerney/dappy"
	"go.etcd.io/etcd/clientv3"
	"net/http"
	"strings"
	"sync"
	"time"
)

type logConfig struct {
	Type     string `json:"type"`     // 采集方式
	Topic    string `json:"topic"`    // 主题
	LogPath  string `json:"log_path"` //采集地址
	Service  string `json:"service"`
	Regexp     string `json:"regexp"`      //日志正则表达式
	SendRate int    `json:"send_rate"` // 速率
	Ip       string `json:"host"`      // 主机
	Dir      bool   `json:"is_dir"`    //是否是文件夹
}

type OffsetConfig struct {
	Offset    int64
	FilePatch string
	Key       string
	Time      string
	Topic     string
	TPS       int64
	QPS       int64
}


var (
	dialTimeout    = 100 * time.Second
	endpoints      = []string{"etcd.waterelephant.cn:23791"}
)

var broadcast = make(chan []OffsetConfig,100)

func watcher() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}

	fmt.Println("connect succ")
	defer cli.Close()
	for {

		// watch key 监听节点
		var dataArray = []OffsetConfig{}
		rch, _ := cli.Get(context.Background(), "/pulse", clientv3.WithPrefix())
		for i := 0; i < len(rch.Kvs); i++ {
			var data = OffsetConfig{ }
			if strings.Contains(string(rch.Kvs[i].Value),"/health"){
				json.Unmarshal(rch.Kvs[i].Value, &data)
				time.Sleep(time.Second)
				dataArray = append(dataArray, data)
			}
		}
		broadcast <- dataArray

	}
}

func LdapLogin(c *gin.Context) {
	//create a new client
	client := dappy.New(dappy.Options{
		BaseDN:       "CN=Users,DC=Company",
		Filter:       "sAMAccountName",
		BasePassword: "basePassword",
		BaseUser:     "baseUsername",
		URL:          "ldap.directory.com:389",
		Attrs:        []string{"cn", "mail"},
	})
	//username and password to authenticate
	username, _ := c.Params.Get("username")
	password, _ := c.Params.Get("pass")

	//attempt the authentication
	err := client.Authenticate(username, password)

	//see the results
	if err != nil {
		logs.Error(err)
	} else {
		logs.Info("user successfully authenticated!")
	}

	//get a user entry
	user, err := client.GetUserEntry(username)
	if err == nil {
		user.PrettyPrint(2)
	}
	c.Writer.WriteString(user.DN)
	c.Writer.Flush()
}

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 客户端读写消息
type wsMessage struct {
	messageType int
	data        []byte
}

// 客户端连接
type wsConnection struct {
	wsSocket *websocket.Conn // 底层websocket
	inChan   chan *wsMessage // 读队列
	outChan  chan *wsMessage // 写队列

	mutex     sync.Mutex // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

func (wsConn *wsConnection) wsReadLoop() {
	for {
		// 读一个message
		msgType, data, err := wsConn.wsSocket.ReadMessage()
		if err != nil {
			goto error
		}
		req := &wsMessage{
			msgType,
			data,
		}
		// 放入请求队列
		select {
		case wsConn.inChan <- req:
		case <-wsConn.closeChan:
			goto closed
		}
	}
error:
	wsConn.wsClose()
closed:
}

func (wsConn *wsConnection) wsWriteLoop() {
	for {
		select {
		// 取一个应答
		case msg := <-wsConn.outChan:
			// 写给websocket
			time.Sleep(1 * time.Second)
			v := <-broadcast
			data,_:=json.Marshal(v)
			if data==nil{
				break
			}
			if err := wsConn.wsSocket.WriteMessage(msg.messageType, data); err != nil {
				goto error
			}
		case <-wsConn.closeChan:
			goto closed
		}
	}
error:
	wsConn.wsClose()
closed:
}

func (wsConn *wsConnection) procLoop() {
	// 启动一个gouroutine发送心跳
	go func() {
		for {
			time.Sleep(1 * time.Second)
			v := <-broadcast
			data,_:=json.Marshal(v)
			if data==nil{
				break
			}
			if err := wsConn.wsWrite(websocket.TextMessage, data); err != nil {
				fmt.Println("heartbeat fail")
				fmt.Println(fmt.Sprintf("fuck data %s",string(data)))
				wsConn.wsClose()
				break
			}
		}
	}()

	// 这是一个同步处理模型（只是一个例子），如果希望并行处理可以每个请求一个gorutine，注意控制并发goroutine的数量!!!
	for {
		msg, err := wsConn.wsRead()
		if err != nil {
			fmt.Println(err)
			fmt.Println("read fail")
			break
		}
		fmt.Println(string(msg.data))
		err = wsConn.wsWrite(msg.messageType, msg.data)
		if err != nil {
			fmt.Println("write fail")
			break
		}
	}
}

func wsHandler(resp http.ResponseWriter, req *http.Request) {
	// 应答客户端告知升级连接为websocket
	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}
	wsConn := &wsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *wsMessage, 1000),
		outChan:   make(chan *wsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	// 处理器
	go wsConn.procLoop()
	// 读协程
	go wsConn.wsReadLoop()
	// 写协程
	go wsConn.wsWriteLoop()
}

func (wsConn *wsConnection) wsWrite(messageType int, data []byte) error {
	select {
	case wsConn.outChan <- &wsMessage{messageType, data}:
	case <-wsConn.closeChan:
		return errors.New("websocket closed")
	}
	return nil
}

func (wsConn *wsConnection) wsRead() (*wsMessage, error) {
	select {
	case msg := <-wsConn.inChan:
		return msg, nil
	case <-wsConn.closeChan:
	}
	return nil, errors.New("websocket closed")
}

func (wsConn *wsConnection) wsClose() {
	wsConn.wsSocket.Close()

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}


func main() {
	router := gin.Default()
	router.POST("/login_ldap", LdapLogin)
	router.GET("/watcher", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	router.Use(static.Serve("/", static.LocalFile("./html", true)))
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
	go watcher()
	router.Run(":3000")
}
