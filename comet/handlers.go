package comet

import (
	log "code.google.com/p/log4go"
	"errors"
	"github.com/gorilla/websocket"
	"livebarrage/utils"
	"strings"
)

/*
** client必须进行login才可以使用该websocket进行通信
** 该接口最长被调用，所以要尽量优化，不能耗时太长
 */
func HandleInit(server *Server, ws *websocket.Conn) (*Client, error) {
	messageType, readBuf, err := ws.ReadMessage()

	cometMsg, err := utils.CometMsgPara(readBuf)
	if err != nil {
		log.Error("ws login failed, err: %s", err.Error())
		return nil, err
	}

	if strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_INIT) != 0 {
		log.Error("ws msgtype is not init")
		return nil, errors.New("ws msgtype is not init")
	}

	log.Info("CometServer HandleInit() msgtype: %d, userId = %s, msg: %s", messageType, cometMsg.UserId, string(readBuf))

	client := NewClient(cometMsg.UserId, cometMsg.LiveId, ws, server.getJobChannel(), server.getOutChannel(), cometMsg.DevType)
	client.status = CLIENT_CONNECTED

	return client, nil
}

//弹幕消息广播发送
func HandlePushBroadcast(server *Server, liveid string, msg *utils.CometMsg) {
	clients := server.clientmanager.getByLiveId(liveid)

	for _, client := range clients {
		if client.status == CLIENT_BROKEN || client.status == CLIENT_CLOSING {
			continue
		}
		client.SendMsgByLiveId(liveid, msg)
	}
}
