package main

//https://viblo.asia/p/build-a-realtime-chat-server-with-go-and-websockets-naQZR1zXKvx
import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var count = 0

//var arr []*websocket.Conn
var clients = make(map[*websocket.Conn]int) // connected clients
var broadcast = make(chan Message)          // broadcast channel
type Message struct {
	Id      int
	Message string
}

func reader(conn *websocket.Conn) {
	clients[conn] = count
	fmt.Println(clients)
	err := conn.WriteMessage(1, []byte("Hi Client!"+strconv.Itoa(clients[conn])))
	if err != nil {
		log.Println(err)
	}
	for {
		// read in a message
		var p Message
		_, p.Message, err := conn.ReadMessage(p.Message)
		if err != nil {
			log.Println(err)
			count--
			return
		}

		//var name int = count
		// print out that message for clarity
		//log.Println("Client" + strconv.Itoa(clients[conn]) + " " + string(p))
		for cli := range clients {
			if err := cli.WriteMessage(websocket.TextMessage, []byte(p.Message)); err != nil {
				log.Println(err)
				cli.Close()
				delete(clients, cli)
				return
			}
		}
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()
	/* 	log.Println("Client Connected")
	   	arr = append(arr, ws)
	   	count++
	   	fmt.Println(arr)
	   	var name int = count
	   	err = ws.WriteMessage(1, []byte("Hi Client!"+strconv.Itoa(name)))
	   	if err != nil {
	   		log.Println(err)
	   	} */
	count++
	reader(ws)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
