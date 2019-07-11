package main

import (
	"bufio"
	"fmt"
	"github.com/Shopify/sarama"
	"gopkg.in/jcmturner/gokrb5.v5/client"
	"gopkg.in/jcmturner/gokrb5.v5/config"
	"gopkg.in/jcmturner/gokrb5.v5/keytab"
	"io/ioutil"
	"os"
	"strings"
)

const (
	kRB5CONF1 = `[libdefaults]
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
		`
)

func main() {

	bst, _ := ioutil.ReadFile("test/bigdata.hive.keytab")

	kt, _ := keytab.Parse(bst)
	c, _ := config.NewConfigFromString(kRB5CONF1)
	c.LibDefaults.NoAddresses = true
	cl := client.NewClientWithKeytab("bigdata.hive/admin", "SXKJ.COM", kt)
	cl.WithConfig(c)
	cl.Login()

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	producer, err := sarama.NewSyncProducer([]string{"cdh-s1.sxkj.online:9092"}, config)

	if err != nil {
		panic(err)
	}
	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic:     "sentry-test",
		Partition: int32(-1),
		Key:       sarama.StringEncoder("key"),
	}

	var value string
	for {
		// 生产消息
		inputReader := bufio.NewReader(os.Stdin)
		value, err = inputReader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		value = strings.Replace(value, "\n", "", -1)
		msg.Value = sarama.ByteEncoder(value)
		paritition, offset, err := producer.SendMessage(msg)

		if err != nil {
			fmt.Println("Send Message Fail")
		}

		fmt.Printf("Partion = %d, offset = %d\n", paritition, offset)
	}
}
