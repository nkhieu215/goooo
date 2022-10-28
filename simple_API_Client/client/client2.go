package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var accts = make(chan string)

type Messages struct {
	User    string `json:"user"`
	Message string `json:"message"`
}
type users struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

var acct = make(chan users)
var addr = flag.String("addr", "localhost:8000", "http service address")

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/signin"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("client connect success")
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var msg users
			err = c.ReadJSON(&msg.Name)
			if err != nil {
				log.Printf("error: %v", err)
				break
			}
			log.Println("Hi:", msg.Name)
		}
	}()

	go func() {
		var msg users
		//fmt.Print("Enter message: ")
		msg = users{Name: "hieu", Password: "123"}
		acct <- msg
	}()

	for {
		select {
		case <-done: //khi server dong thi client cung tu dong dong theo
			return
		case msg2 := <-acct:
			log.Println("write:", msg2)
			err := c.WriteJSON(msg2)
			if err != nil {
				log.Println(err)
				return
			}
		case <-interrupt: // dung de client co the tu dong out khoi server
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}
