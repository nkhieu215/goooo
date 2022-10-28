package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

//var account = make(chan Credentials)
var clients = make(map[*websocket.Conn]string)
var Msg = make(chan Message)

const (
	hashCost = 8
	host     = "localhost"
	user     = "nguyenkhachieu"
	password = "anhiuem2"
	dbname   = "postgres"
)

var db *sql.DB

func main() {
	// "Signin" and "Signup" are handler that we will implement
	http.HandleFunc("/signin", Signin) /// /ws
	http.HandleFunc("/signup", Signup)
	// initialize our database connection
	initDB()
	// start the server on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func initDB() {
	var err error
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	// Connect to the postgres db
	//you might have to change the connection string to add your database credentials
	db, err = sql.Open("postgres", conn)
	if err != nil {
		panic(err)
	}
	log.Println("pingok")
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Name     string `json:"name", db:"name"`
	Password string `json:"password", db:"password"`
}

type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	/* hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	// Next, insert the username, along with the hashed password into the database
	if _, err = db.Query("insert into login values ($1, $2)", creds.Name, string(hashedPassword)); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		return
	} */
	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
	if _, err = db.Query("insert into login values ($1, $2)", creds.Name, creds.Password); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("signin:", creds.Password)
}

func Signin(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	for {
		var msg Credentials
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			log.Println("account in server:", clients)
			break
		}
		log.Println("read:", msg)
		// Send the newly received message to the broadcast channel

		// Parse and decode the request body into a new `Credentials` instance

		creds := msg
		// Get the existing entry present in the database for the given username
		result := db.QueryRow("select password from login where name=$1", creds.Name)
		log.Println("2: ", creds.Name, creds.Password)
		if err != nil {
			// If there is an issue with the database, return a 500 error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// We create another instance of `Credentials` to store the credentials we get from the database
		storedCreds := &Credentials{}
		// Store the obtained password in `storedCreds`
		err = result.Scan(&storedCreds.Password)
		if err != nil {
			// If an entry with the username does not exist, send an "Unauthorized"(401) status
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// If the error is of any other type, send a 500 status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("3: ", storedCreds.Password)
		// Compare the stored hashed password, with the hashed version of the password that was received
		/* if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			// If the two passwords don't match, return a 401 status
			w.WriteHeader(http.StatusUnauthorized)
		} */
		if creds.Password == storedCreds.Password {
			err = ws.WriteJSON(creds.Name)
			if err != nil {
				log.Println(err)
			}
			clients[ws] = creds.Name
			//http.HandleFunc("/ws", Chat)
			log.Println("account: ", clients)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}
