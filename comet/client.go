package comet

import (
	log "code.google.com/p/log4go"
	"github.com/gorilla/websocket"
	"livebarrage/utils"
	"sync"
)

const (
	CLIENT_CONNECTED uint8 = 0 // client has connected, it's ready to send message exclude PUSH
	CLIENT_READY     uint8 = 1 // client has finish INIT, it's ready to push message
	CLIENT_BROKEN    uint8 = 2 // client has broken
	CLIENT_KICK1     uint8 = 3 // client has been kick away by new connection
	CLIENT_KICK2     uint8 = 4 // client has been kick away by command
	CLIENT_CLOSING   uint8 = 5 // client is under closing, this is the last status
)

type Client struct {
	userId     string
	liveId     string
	devtype    uint8
	version    uint16
	ws         *websocket.Conn
	status     uint8
	JobChannel *chan Job
	OutChannel *chan Job
	lock       *sync.RWMutex
}

func NewClient(userId string, liveId string, ws *websocket.Conn, jobChannel *chan Job, outChannel *chan Job, devtype uint8) *Client {
	return &Client{
		userId:     userId,
		liveId:     liveId,
		devtype:    devtype,
		ws:         ws,
		JobChannel: jobChannel,
		OutChannel: outChannel,
		lock:       new(sync.RWMutex),
	}
}

func (client *Client) addJob(job Job) {
	if len(*(client.JobChannel)) >= 300 {
		log.Warn("conn %p, channel %p: job channel full: len %d",
			client.ws, client.JobChannel, len(*(client.JobChannel)))
	}
	*(client.JobChannel) <- job
}

func (client *Client) addOutJob(job Job) {
	if len(*(client.OutChannel)) >= 500 {
		log.Warn("conn %p, channel %p: job channel full: len %d",
			client.ws, client.JobChannel, len(*(client.JobChannel)))
	}
	*(client.JobChannel) <- job
}

func (client *Client) SendMsgByLiveId(liveid string, rmsg *utils.CometMsg) bool {
	//log.Info("SendMsgByLiveId start")
	job := &DataSendJob{
		client: client,
		rmsg:   rmsg,
	}
	// TODO: 可能会阻塞
	if len(*(client.OutChannel)) >= 500 {
		log.Warn("conn %p, channel %p: out channel full 1: len %d",
			client.ws, client.OutChannel, len(*(client.OutChannel)))
		client.ws.Close()
		return false
	}
	*(client.OutChannel) <- job
	//log.Info("SendMsgByLiveId end")
	return true
}

type DataSendJob struct {
	client *Client
	rmsg   *utils.CometMsg
}

func (this *DataSendJob) Do(worker int) bool {
	//log.Info("DataSendJob Do start")
	client := this.client
	rmsg := this.rmsg

	if client.status != CLIENT_READY && client.status != CLIENT_CONNECTED {
		return true
	}

	if err := client.ws.WriteJSON(rmsg); err != nil {
		client.status = CLIENT_BROKEN
		log.Info("datasend job error, liveid: %s", rmsg.LiveId)
	}else {
		log.Info("DataSendJob send json success, liveid: %s", rmsg.LiveId)
	}

	return true
}
