package utils

import (
	log "code.google.com/p/log4go"
	"encoding/json"
)

const (
	MSG_TYPE_INIT     string = "init"     // client has connected, init client
	MSG_TYPE_BARRAGE  string = "barrage"  // client has finish INIT, it's push message
	MSG_TYPE_COMMENTS string = "comments" // client has finish INIT, it's push message
	MSG_TYPE_HEARTBT  string = "heartbt"  // client has finish INIT, it's reply heartbt
	MSG_TYPE_GIFT     string = "gift"     //client has finish INIT, it's push message
	MSG_TYPE_LOGIN    string = "login"    //client has finish INIT， it login server
)

type CometMsg struct {
	UserId  string `json:"userId"`
	Level   int    `json:"level"`
	LiveId  string `json:"liveId"`
	DevType uint8  `json:"devType"`
	MsgType string `json:"msgType"`
	MsgInfo string `json:"msgInfo"`
}

type BarrageMsg struct {
	MsgType string `json:"msgType"`
	LiveId  string `json:"liveId"`
	UserId  string `json:"userId"`
	MsgInfo string `json:"msgInfo"`
}

func CometMsgPara(readBuf []byte) (*CometMsg, error) {
	cometMsg := &CometMsg{}
	err := json.Unmarshal(readBuf, cometMsg)
	if err != nil {
		log.Error("WS Read Msg farmat error is ", err.Error())
		return cometMsg, err
	}

	return cometMsg, nil
}
