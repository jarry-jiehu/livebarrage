package comet

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"livebarrage/conf"
	"livebarrage/utils"
	"strings"
	"time"
)

//进行producter，生产inmsg到kafka中
type Producer struct {
	Addr          []string
	KafkaProducer sarama.AsyncProducer
	TopicStr      string
	ProducerDone  chan bool
}

func NewProducer() *Producer {
	return &Producer{}
}

func (this *Producer) Publish(client *Client, msg *utils.CometMsg) bool {
	job := &DataPublishJob{
		producer: this,
		client:   client,
		cometMsg: msg,
	}
	// TODO: 可能会阻塞
	if len(*(client.JobChannel)) >= 500 {
		log.Warn("conn %p, channel %p: out channel full 1: len %d",
			client.ws, client.OutChannel, len(*(client.OutChannel)))
	}
	*(client.JobChannel) <- job
	//log.Info("Publish end")

	return true
}

type DataPublishJob struct {
	producer *Producer
	client   *Client
	cometMsg *utils.CometMsg
}

func (this *DataPublishJob) Do(msgtype int) bool {
	//to do 发布消息
	log.Info("DataPublishJob liveid: %s, mystype: %s, msginfo: %s", this.cometMsg.LiveId, this.cometMsg.MsgType, this.cometMsg.MsgInfo)
	content, err := json.Marshal(this.cometMsg)
	if err != nil {
		log.Info("DataPublishJob json marshal err:%+v", err)
	} else {
		this.producer.PublishMsgProducer(string(content))
	}
	return true
}

func (this *Producer) InitKafkaProducer(kafkaConf *conf.KafkaConf) {
	var err error
	producerConfig := sarama.NewConfig()
	//等待服务器所有副本都保存成功后的响应
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	//随机向partition发送消息
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	producerConfig.Version = sarama.V0_10_0_1

	this.TopicStr = kafkaConf.Topic
	this.Addr = strings.Split(kafkaConf.Addr, ",")
	this.ProducerDone = make(chan bool)
	fmt.Println("start make producer")
	//使用配置,新建一个异步生产者
	this.KafkaProducer, err = sarama.NewAsyncProducer(this.Addr, producerConfig)
	if err != nil {
		fmt.Println(err)
		log.Error("new async producer err:", err)
	}
}

func (this *Producer) StartProducerRecord() {
	defer this.KafkaProducer.AsyncClose()
	for {
		select {
		case suc, ok := <-this.KafkaProducer.Successes():
			if ok {
				log.Info("KafkaProducer sendOk offset: %d,timestamp:%s,partitions:%d,topic:%s", suc.Offset, suc.Timestamp.Format("2006-01-02 15:04:05"), suc.Partition, suc.Topic)
			}
		case fail, ok := <-this.KafkaProducer.Errors():
			if ok {
				//fmt.Println("producer err: ", fail.Err)
				log.Info("KafkaProducer send err: ", fail.Err)
			}
		case <-this.ProducerDone:
			log.Info("KafkaProducer Finish!")
			close(this.ProducerDone)
			return
		}
	}
}

func (this *Producer) PublishMsgProducer(content string) {
	msg := &sarama.ProducerMessage{
		Topic:     this.TopicStr,
		Value:     sarama.ByteEncoder(content),
		Timestamp: time.Now(),
	}

	//log.Info("kafka producer publish msg: %s", content)
	//发送消息
	this.KafkaProducer.Input() <- msg
}
