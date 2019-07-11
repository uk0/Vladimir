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

type KafkaSenderKDC struct {
	client *kafka.Producer
	lineChan chan *MessageKDC
}

// 初始化kafka
func NewKafkaSenderKDC(kafkaAddr string,auth bool)(afk *KafkaSenderKDC,err error){
	afk = &KafkaSenderKDC{
		lineChan:make(chan *MessageKDC,10000),
	}
	if auth {
		os.Setenv("KRB5_CONFIG","./kr.conf")
		client, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":          kafkaAddr,
			"sasl.kerberos.service.name": "kafka",
			"sasl.kerberos.keytab":       "./bigdata.hive.keytab",
			"sasl.kerberos.principal":    "bigdata.hive/admin@we.COM",
			"security.protocol":          "SASL_PLAINTEXT",
			"session.timeout.ms":         1000,
		})
	}else{
		client, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":          kafkaAddr,
			"session.timeout.ms":         1000,
		})
		//
	}

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", client)
	afk.client = client
	for i:=0;i<appConfig.KafkaThreadNum;i++{
		// 根据配置文件循环开启线程去发消息到kafka
		go afk.sendToKafka()
	}
	return
}

func InitKafkaKDC()(err error){
	kafkaSender,err = NewKafkaSenderKDC(appConfig.kafkaAddr,true)
	return
}

func (k *KafkaSenderKDC) sendToKafka(){
	//从channel中读取日志内容放到kafka消息队列中
	for v := range k.lineChan{
		deliveryChan := make(chan kafka.Event)
		ip,_:=getLocalIP()
		err := client.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &v.topic, Partition: kafka.PartitionAny},
			Value:          []byte(v.line),
			Headers:        []kafka.Header{{Key: ip[0], Value: []byte("service running on host")}},
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
		}
		close(deliveryChan)
	}

}

func (k *KafkaSenderKDC) addMessage(line string,topic string)(err error){
	//我们通过tailf读取的日志文件内容先放到channel里面
	logs.Info("Kafka Add Message")
	k.lineChan <- &MessageKDC{line:line,topic:topic}
	return
}
