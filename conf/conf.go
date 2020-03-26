package conf

import (
	"encoding/json"
	"os"
	"time"
)

type CometConfig struct {
	Number           uint32 `jons:"num"`
	Port             string `json:"port"`
	PublicAddress    string `json:"public_address"`
	InternalIp       string `json:"internal_ip"`
	ExternalIp       string `json:"external_ip"`
	AcceptTimeout    uint32 `json:"accept_timeout"`
	AcceptInterval   uint32 `json:"accept_interval"`
	ReadTimeout      uint32 `json:"read_timeout"`
	WriteTimeout     uint32 `json:"write_timeout"`
	InitTimeout      uint32 `json:"init_timeout"`
	MaxIniting       uint32 `json:"max_initing"`
	MaxBodyLen       uint32 `json:"max_bodylen"`
	MaxClients       uint32 `json:"max_clients"`
	WorkerCnt        int    `json:"worker_count"`
	JobWorkerCnt     int    `json:"job_worker_count"`
	OutWorkerCnt     int    `json:"out_worker_count"`
	HbTimeout        uint32 `json:"heartbeat_timeout"`
	HbInterval       uint32 `json:"heartbeat_interval"`
	ReconnTime       uint32 `json:"reconn_time"`
	MobileHbTimeout  uint32 `json:"mobile_heartbeat_timeout"`
	MobileHbInterval uint32 `json:"mobile_heartbeat_interval"`
	MobileReconnTime uint32 `json:"mobile_reconn_time"`
	RouterHbTimeout  uint32 `json:"router_heartbeat_timeout"`
	RouterHbInterval uint32 `json:"router_heartbeat_interval"`
	RouterReconnTime uint32 `json:"router_reconn_time"`
	Region           string `json:"region"`
	Group            string `json:"group"`
	Idc              string `json:"idc"`
}

type ZKConfig struct {
	Enable  bool          `json:"enable"`
	Addr    string        `json:"addr"`
	Timeout time.Duration `json:"timeout"`
	Path    string        `json:"path"`
}

type KafkaConf struct {
	Addr string `json:"addr"`
	Topic string `json:"topic"`
}

type ConfigStruct struct {
	Comet     CometConfig   `json:"comet"`
	Kafka     KafkaConf     `json:"kafka"`
}

var (
	Config ConfigStruct
)

func LoadConfig(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&Config)
	if err != nil {
		return err
	}
	return nil
}
