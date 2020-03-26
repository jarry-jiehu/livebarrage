package comet

import (
	log "code.google.com/p/log4go"
	"encoding/json"
	"github.com/gorilla/websocket"
	"livebarrage/conf"
	"livebarrage/utils"
	"math/rand"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
)

var (
	CometServer *Server = nil
	mNum int = 0
)

const (
	DEFAULT_WORKER_POOL_SIZE = 100
	DEFAULT_WORKER_BUF_SIZE  = 100
	DEFAULT_REG_WAIT_TIMEOUT = 10
	DEFAULT_CLIENT_MAX_COUNT = 100000
	DEFAULT_READ_BUF_SIZE    = 8192 * 2
	DEFAULT_WRITE_BUF_SIZE   = 8192 * 2
	DEFAULT_KEY_EXPIRE_TTL   = int32(3600 * 24)
)

type Counter struct {
	total   int32
	initing int32
	types   []int32
}

type Server struct {
	Name string // unique name of this server
	ctrl chan bool
	lock *sync.Mutex
	wg   *sync.WaitGroup
	//funcMap        map[uint8]MsgHandler
	conf          *conf.CometConfig
	counter       Counter
	jobWorkers    map[int]*Worker
	outWorkers    map[int]*Worker
	clientmanager *ClientManager
	//msgManager     *MsgManager
	//statsManager   *StatsManager
	serviceManager *ServiceManager
	//headpool       *pool.Pool
	//datapool       *pool.Pool
}

func NewServer(config *conf.CometConfig) *Server {
	conf := *config

	return &Server{
		Name: config.InternalIp,
		ctrl: make(chan bool),
		lock: new(sync.Mutex),
		wg:   &sync.WaitGroup{},
		//funcMap:        make(map[uint8]MsgHandler),
		jobWorkers:    make(map[int]*Worker),
		outWorkers:    make(map[int]*Worker),
		conf:          &conf,
		counter:       Counter{types: []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		clientmanager: NewClientManager(),
		//msgManager:     NewMsgManager(),
		//statsManager:   NewStatsManager(),
		serviceManager: NewServiceManager(),
		//headpool:       pool.NewPool(10000, func() interface{} { return make([]byte, packet.HEADER_SIZE) }),
		//datapool:       pool.NewPool(10000, func() interface{} { return make([]byte, conf.MaxBodyLen) }),
	}
}

func InitServer(config *conf.CometConfig) *Server {
	s := NewServer(config)
	CometServer = s
	s.serviceManager.InitService()
	return s
}

func (this *Server) StartServer() {
	this.startWorkers()
	this.serviceManager.StartService()
	//注册ws路由的处理句柄
	go this.startWebSocket()

	select {
	case <-this.ctrl:
		this.stopWorkers()
		log.Error("startServer failed!")
	}
}

func (this *Server) startWorkers() {
	for n := 0; n < this.conf.JobWorkerCnt; n++ {
		w := NewWorker(n, 2000)
		this.jobWorkers[n] = w
		w.Run()
	}
	for n := 0; n < this.conf.OutWorkerCnt; n++ {
		w := NewWorker(n, 2000)
		this.outWorkers[n] = w
		w.Run()
	}
	log.Info("comet server: all workers started")
}

func (this *Server) stopWorkers() {
	for _, w := range this.jobWorkers {
		w.Stop()
	}
	for _, w := range this.outWorkers {
		w.Stop()
	}
}

func (this *Server) getJobChannel() *chan Job {
	id := rand.Intn(this.conf.JobWorkerCnt)
	return &this.jobWorkers[id].jobChannel
}

func (this *Server) getOutChannel() *chan Job {
	id := rand.Intn(this.conf.OutWorkerCnt)
	return &this.outWorkers[id].jobChannel
}

func (this *Server) StopServer() {
	// close后，所有的ctrl都返回false
	log.Info("comet server: stopping")
	close(this.ctrl)
	this.serviceManager.StopService()
	this.wg.Wait()
	log.Info("comet server: stopped")
}

func (this *Server) startWebSocket() {
	url := "0.0.0.0:8080"
	http.HandleFunc("/ws", this.serveWebsocket)
	if err := http.ListenAndServe(url, nil); err != nil {
		log.Error("CometServer ListenAndServe failed Error = ", err)
	}
}

func (this *Server) serveWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  DEFAULT_READ_BUF_SIZE,
		WriteBufferSize: DEFAULT_WRITE_BUF_SIZE,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrader err is : ", err, ", remoteaddr is : ", r.RemoteAddr)
		return
	}

	go this.handleWebsocket(ws) //per connect per goroutine
}

func (this *Server) handleWebsocket(ws *websocket.Conn) {
	log.Info("CometServer handleWebsocket() remoteAddr = %s", ws.RemoteAddr().String())
	defer func() {
		if r := recover(); r != nil {
			log.Error("Runtime error caught : ", r)
			log.Error(string(debug.Stack()))
		}
		log.Error("ws close")
		ws.Close()
	}()

	ws.SetReadLimit(int64(DEFAULT_READ_BUF_SIZE))

	client, err := HandleInit(CometServer, ws)
	if err != nil {
		return
	}
	this.clientmanager.Set(string(mNum), client)
	mNum++
	log.Info("get device manager userid: %s, client size: %d", client.userId, this.clientmanager.Size())

	for {
		if client.status == CLIENT_BROKEN || client.status == CLIENT_CLOSING {
			log.Warn("client warning exit, userid: %s", client.userId)
			this.clientmanager.Remove(client.userId)
			return
		}
		messageType, readBuf, err := ws.ReadMessage()
		if err != nil {
			log.Error("WS ReadMessage err is %s", err.Error())
			client.status = CLIENT_BROKEN
			continue
		}

		var cometMsg utils.CometMsg
		err = json.Unmarshal(readBuf, &cometMsg)
		if err != nil {
			log.Error("WS Read msgtype: %d Msg farmat error is %s", messageType, err.Error())
			client.status = CLIENT_BROKEN
			continue
		}

		if strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_COMMENTS) == 0 ||
			strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_BARRAGE) == 0 ||
			strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_LOGIN) == 0 ||
			strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_GIFT) == 0 {
			//如果是评论或者弹幕，直接推送出去
			this.serviceManager.HandleInMsg(client, &cometMsg)
		} else if strings.Compare(cometMsg.MsgType, utils.MSG_TYPE_HEARTBT) == 0 {
			//心跳
			client.SendMsgByLiveId(cometMsg.LiveId, &cometMsg)
		}
	}
}
