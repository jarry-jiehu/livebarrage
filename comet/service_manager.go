package comet

import (
	log "code.google.com/p/log4go"
	"livebarrage/conf"
	"livebarrage/utils"
	"sync"
)

type ServiceManager struct {
	lock     *sync.RWMutex
	consumer *KafkaConsumer
	producer *Producer
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		lock:     new(sync.RWMutex),
		producer: NewProducer(),
		consumer: NewKafkaConsumer(
			&conf.Config.Kafka,
		),
	}
}

func (this *ServiceManager) InitService() {
	this.producer.InitKafkaProducer(&conf.Config.Kafka)
	log.Info("livebarrage: kafka server inited")
	this.consumer.InitConsumer()
}

func (this *ServiceManager) StartService() {
	go this.producer.StartProducerRecord()
	go this.consumer.StartConsumerMsg()
}

func (this *ServiceManager) StopService() {
	if this.producer.ProducerDone != nil {
		this.producer.ProducerDone <- true
	}

	if this.consumer.OverCh != nil {
		this.consumer.OverCh <- true
	}
}

func (this *ServiceManager) HandleInMsg(client *Client, msg *utils.CometMsg) {
	producter := this.producer
	producter.Publish(client, msg)
}
