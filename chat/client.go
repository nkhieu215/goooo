package main

// trinh bay bang websocket

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
type Message struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("client connect success")
	defer c.Close()
	done := make(chan struct{})
	go func() {
		//msg2 := <-broadcast
		fmt.Println(broadcast)
	}()
	go func() {
		defer close(done)
		for {
			var msg Message
			// Read in a new message as JSON and map it to a Message object
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients, c)
				break
			}
			// Send the newly received message to the broadcast channel
			broadcast <- msg
		}
	}()
	//a := make(chan string)
	var msg Message
	//fmt.Print("Enter message: ")
	msg = Message{Id: 2, Message: "Nguyen Khac Hieu"}
	fmt.Scan(&msg)
	broadcast <- msg

	//fmt.Scan(&broadcast)
	//msg2 := msg

	for {
		select {
		case <-done:
			return
		case msg := <-broadcast:
			err := c.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
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
