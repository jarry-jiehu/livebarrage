package comet

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"livebarrage/conf"
	"livebarrage/utils"
	"os"
	"os/signal"
	"strings"
	"time"
)

type KafkaConsumer struct {
	Brokers []string
	Topics  []string
	OverCh  chan bool
	CtrlCh  []chan bool
}

func NewKafkaConsumer(kafkaConf *conf.KafkaConf) *KafkaConsumer {
	return &KafkaConsumer{
		Brokers: strings.Split(kafkaConf.Addr, ","),
		Topics:  []string{kafkaConf.Topic},
		CtrlCh:  make([]chan bool, 0),
		OverCh:  make(chan bool),
	}
}

func (this *KafkaConsumer) InitConsumer() {
	go func() {
		<-this.OverCh
		for _, ch := range this.CtrlCh {
			ch <- true
		}
		close(this.OverCh)
	}()
	return
}

func (this *KafkaConsumer) kafkaRegisterConsumerCluster(consumerId string, groupId string, consumerMsgCh chan []byte, overCh chan bool) {

	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Group.Return.Notifications = true

	consumer, err := cluster.NewConsumer(this.Brokers, groupId, this.Topics, config)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	done := make(chan bool)
	this.CtrlCh = append(this.CtrlCh, done)
	log.Info("start consumer-cluster")
	count := 0
	// 消费消息
	for {
		select {
		case msg, ok := <-consumer.Messages():
			count++
			//count just for test
			if count > 100000000 {
				count = 1
			}
			log.Info("============Consumer============receive msg total count:%d", count)
			if ok {
				log.Info("KafkaConsumer %s %s: %s/%d/%d\tvalue:%s", consumerId, groupId, msg.Topic, msg.Partition, msg.Offset, msg.Value)
				if consumerMsgCh != nil {
					consumerMsgCh <- msg.Value
				}
				consumer.MarkOffset(msg, "") // 提交offset
			}
		case err := <-consumer.Errors():
			// 打印错误信息
			log.Info("Consumer-cluster　id:%s groupId:%s Error: %s", consumerId, groupId, err.Error())
		case ntf := <-consumer.Notifications():
			// 打印rebalance信息
			log.Info("Rebalanced: %+v\t %s %s", ntf, consumerId, groupId)
		case <-signals:
			log.Info("Consumer-cluster signal to end! ", consumerId, groupId)
			if this.OverCh != nil {
				this.OverCh <- true
			}
			if overCh != nil {
				overCh <- true
			}
			return
		case <-done:
			log.Info("Consumer-cluster OverCh to end! ", consumerId, groupId)
			if overCh != nil {
				overCh <- true
			}
			close(done)
			return
		}
	}
}

func (this *KafkaConsumer) StartConsumerMsg() {
	consumerCh := make(chan []byte, 1000)
	overCh := make(chan bool)
	ipStr := utils.GetIp()
	go this.kafkaRegisterConsumerCluster(ipStr, "group6"+ipStr, consumerCh, overCh)
	for {
		select {
		case msgBytes, Ok := <-consumerCh:
			if !Ok {
				break
			}
			msg := &utils.CometMsg{}
			err := json.Unmarshal(msgBytes, msg)
			if err != nil {
				log.Error("unmarshal json msg err:", err, " msg:", string(msgBytes))
				continue
			}
			HandlePushBroadcast(CometServer, msg.LiveId, msg)
		case <-overCh:
			log.Info("comet consumer over!")
			close(overCh)
			close(consumerCh)
			return
		}
	}
}
