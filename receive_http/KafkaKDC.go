package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
)


var (
	client *kafka.Producer
	kafkaSender *KafkaSenderKDC
)

type MessageKDC struct {
	line string
	topic string
}

var offset  = make(chan string)

type KafkaSenderKDC struct {
	client *kafka.Producer
	lineChan chan *MessageKDC
}

var userName = make(chan string,100)

var cacheMap = map[string]bool{}

// 初始化kafka
func NewKafkaSenderKDC(kafkaAddr string,auth bool,user string)(afk *KafkaSenderKDC,err error){
	afk = &KafkaSenderKDC{
		lineChan:make(chan *MessageKDC,100),
	}
	var principal_user string
	if auth {
		logs.Info("Select Auth on Kerberos")
		os.Setenv("KRB5_CONFIG","/etc/krb5.conf")
		if user=="bigdata.hive"{
			principal_user ="bigdata.hive/admin"
		}else {
			principal_user = user
		}
		client, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":          kafkaAddr,
			"sasl.kerberos.service.name": "kafka",
			"sasl.kerberos.keytab":       fmt.Sprintf("/data/public/keytab/%s/%s.keytab",user,user),
			"sasl.kerberos.principal":    fmt.Sprintf("%s@we.COM",principal_user),
			"security.protocol":          "SASL_PLAINTEXT",
			"session.timeout.ms":         1000,
		})
	}else {
		logs.Warn(" Auth Not Select")
		client, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":          kafkaAddr,
			"session.timeout.ms":         1000,
		})
	}


	//

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", client)
	afk.client = client
	for i:=0;i<4;i++{
		// 根据配置文件循环开启线程去发消息到kafka
		go afk.sendToKafka()
	}
	return
}

func InitKafkaKDC()(err error){
	// 初始化通道
	for user :=range userName{

		if ""==user {
			fmt.Println("wait .... 5s")
			continue
		}
		if !cacheMap[user]{
			cacheMap[user]=true;
			kafkaSender,err = NewKafkaSenderKDC("cdh-s1.sxkj.online:9092,cdh-s2.sxkj.online:9092,cdh-s3.sxkj.online:9092,cdh-s4.sxkj.online:9092,cdh-s5.sxkj.online:9092,cdh-s6.sxkj.online:9092,cdh-s7.sxkj.online:9092",
				true,
				user,
			)
		}

		if !cacheMap[user]{
			//存在就跳出去
			fmt.Println(fmt.Sprintf("User format %s %v",user,cacheMap[user]))
			continue
		}

		if err!=nil{
			fmt.Println(fmt.Sprintf("User %s Init Error ",user),err)
			//初始化失败也跳出去
			continue
		}
	}
	return
}

func (k *KafkaSenderKDC) sendToKafka(){
	//从channel中读取日志内容放到kafka消息队列中
	for v := range k.lineChan{
		deliveryChan := make(chan kafka.Event)
		var ip ="127.0.0.1"
		err := client.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &v.topic, Partition: kafka.PartitionAny},
			Value:          []byte(v.line),
			Headers:        []kafka.Header{{Key: ip, Value: []byte("service running on host")}},
		}, deliveryChan)

		e := <-deliveryChan
		m := e.(*kafka.Message)
		if err != nil{
			logs.Error("send message to kafka failed,err:%v",err)
		}
		if m.TopicPartition.Error != nil {
			logs.Error("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			logs.Info("Delivered message to topic %s [%d] at offset %v\n",
				*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			offset<-fmt.Sprintf("%v",m.TopicPartition.Offset)
		}
		close(deliveryChan)
	}

}

func (k *KafkaSenderKDC) addMessage(line string,topic string)(err error){
	//我们通过tailf读取的日志文件内容先放到channel里面
	k.lineChan <- &MessageKDC{line:line,topic:topic}
	return
}
