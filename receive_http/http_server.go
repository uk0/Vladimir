package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"github.com/nerney/dappy"
	"io/ioutil"
	"net/http"
	"time"
)


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

func initLog() (err error) {
	//初始化日志库
	config := make(map[string]interface{})
	config["filename"] = "./http_server.log"
	config["level"] = logs.LevelInfo
	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println(" json.Marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, err := ctx.GetRawData()
		header := ctx.Request.Header
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("data: %v\n", string(data))
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data)) // 关键点
		ctx.Request.Header = header // 关键点
		ctx.Next()
	}
}

func main() {
	initLog()
	go InitKafkaKDC()
	router := gin.Default()
	s := &http.Server{
		Addr:           ":3000",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//router.POST("/login_ldap", LdapLogin)
	api := router.Group("/api")
	{
		api.POST("/msg", Middleware(), func(c *gin.Context) {
			cCp := c.Copy()
			go func() {
				dat1a, err := cCp.GetRawData()
				if err != nil {
					fmt.Println(err.Error())
				}
				var Token = cCp.Request.Header.Get("Token")
				var Topic = cCp.Request.Header.Get("Topic")


				if "" == Token {
					logs.Info("token is Null")
				}
				userName <- Token

				kafkaSender.addMessage(string(dat1a), Topic)
				resp := map[string]string{"status": "successfully", "offset": <-offset, "Topic": Topic}
				fmt.Println(resp)
			}()
		})
	}
	userName <- "bigdata.hive"

	s.ListenAndServe()
}
