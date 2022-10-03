package main

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

//https://webdevelop.pro/blog/guide-creating-websocket-client-golang-using-mutex-and-channel/
type WebsocketClient struct {
	wsconn    *websocket.Conn
	configStr string
}

func (conn *WebsocketClient) Connect() *websocket.Conn {
	if conn.wsconn != nil {
		return conn.wsconn
		fmt.Println(conn.wsconn)
	}

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		ws, _, err := websocket.DefaultDialer.Dial(conn.configStr, nil)
		if err != nil {
			continue
		}
		conn.wsconn = ws
		go conn.listen()
		return conn.wsconn
	}
}
func (conn *WebsocketClient) listen() *websocket.Conn {}
