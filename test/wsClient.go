package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"livebarrage/utils"
	"sync"
	"time"
)

var addr = flag.String("addr", "123.59.236.54:8080", "http service address")

//var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

func main() {
	flag.Parse()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 1000; i++ {
		go func() {
			url := "ws://" + *addr + "/ws"
			origin := "test://1111111/"
			ws, err := websocket.Dial(url, "", origin)
			if err != nil {
				fmt.Println(err)
			}
			go timeWriter(ws, i)
			go readws(ws, i)
			fmt.Printf("creat socket num: %d", i)
		}()
		time.Sleep(time.Millisecond * 300)
	}
	wg.Wait()
	fmt.Println("ws client close")
}

func timeWriter(conn *websocket.Conn, id int) {
	i := 0
	for {
		if i == 0 {
			cometMsg := utils.CometMsg{
				UserId:  fmt.Sprintf("userid_%d", id),
				LiveId:  fmt.Sprintf("liveid1"),
				DevType: 0,
				MsgType: "init",
				MsgInfo: "login",
			}
			sendBuf, _ := json.Marshal(cometMsg)
			websocket.Message.Send(conn, sendBuf)
			i++
		} else if i== 2{
			cometMsg := utils.CometMsg{
				UserId:  fmt.Sprintf("userid_%d", id),
				LiveId:  fmt.Sprintf("liveid1"),
				DevType: 1,
				MsgType: "barrage",
				MsgInfo: fmt.Sprintf("live barrage %d", i),
			}
			sendBuf, _ := json.Marshal(cometMsg)
			websocket.Message.Send(conn, sendBuf)
			i++
		}
		time.Sleep(time.Millisecond * 10000)
	}
}

func readws(conn *websocket.Conn, id int) {
	for {
		var msg [512]byte
		_, err := conn.Read(msg[:]) //此处阻塞，等待有数据可读
		if err != nil {
			fmt.Println("id read:", err)
			return
		}

		fmt.Printf("id : %d , received: %s, id: %d\n", id, msg)
	}
}
