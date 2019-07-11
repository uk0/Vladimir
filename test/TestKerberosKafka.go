package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
)

var Address = []string{
	"cdh-s1.sxkj.online:9092",
	"cdh-s2.sxkj.online:9092",
	"cdh-s3.sxkj.online:9092",
	"cdh-s5.sxkj.online:9092",
	"cdh-s6.sxkj.online:9092",
	"cdh-s7.sxkj.online:9092",
}

func freedConf(){

}


func AsyncProducerToKafka(topic string) {

	os.Setenv("KRB5_CONFIG","./kr.conf")
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":          "cdh-s1.sxkj.online:9092,cdh-s2.sxkj.online:9092,cdh-s3.sxkj.online:9092,cdh-s4.sxkj.online:9092,cdh-s5.sxkj.online:9092,cdh-s6.sxkj.online:9092,cdh-s7.sxkj.online:9092",
		"sasl.kerberos.service.name": "kafka",
		"sasl.kerberos.keytab":       "./bigdata.hive.keytab",
		"sasl.kerberos.principal":    "bigdata.hive/admin@we.COM",
		"security.protocol":          "SASL_PLAINTEXT",
		"session.timeout.ms":         1000,
	})
	//

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)

	// Optional delivery channel, if not specified the Producer object's
	// .Events channel is used.
	deliveryChan := make(chan kafka.Event)

	value := "Hello Go!"
	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
		Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
	}, deliveryChan)

	e := <-deliveryChan
	m := e.(*kafka.Message)
	fmt.Println(m)
	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	close(deliveryChan)
}


const (
	kRB5CONF =
		`[libdefaults]
default_realm = we.COM
dns_lookup_kdc = false
dns_lookup_realm = false
ticket_lifetime = 86400
renew_lifetime = 604800
forwardable = true
default_tgs_enctypes = rc4-hmac
default_tkt_enctypes = rc4-hmac
permitted_enctypes = rc4-hmac
udp_preference_limit = 1
kdc_timeout = 3000
[realms]
SXKJ.COM = {
kdc = cdh-m1.sxkj.online
admin_server = cdh-m1.sxkj.online
}
[domain_realm]
		`
)

func main() {

	//bst,_:=ioutil.ReadFile("/Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/test/keys/bigdata.hive.keytab")
	//
	//kt, _ := keytab.Parse(bst)
	//c, _ := config.NewConfigFromString(kRB5CONF)
	//c.LibDefaults.NoAddresses = true
	//cl := client.NewClientWithKeytab("bigdata.hive/admin", "SXKJ.COM", kt)
	//cl.WithConfig(c)



	AsyncProducerToKafka("sentry-test")

	//err:=cl.Login()
	//if err!=nil{
	//	fmt.Println(err)
	//}
}
