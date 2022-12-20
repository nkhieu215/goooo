package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//------------GLOBAL VARIABLES--------------
var (
	router    *mux.Router
	secretkey string = "SecretKeyjwt!!!"
)

//var messages = make(map[string]string)
//var listToken = make(map[string]string)
var members = make(map[string]map[string]bool)
var check string
var checkMembers string

//--------------tao bien chay-------------------
var sender_id int
var receiver_id int
var room_id int
var room_name string

//----------------Struct---------- khoi tao bien
type Users struct { // khởi tạo thông tin cơ bản của user
	gorm.Model
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type zoom struct {
	Class string `json:"class"`
}

type Token struct {
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

type Authentication struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Error struct {
	IsError bool   `json:"iserror"`
	Message string `json:"message"`
}
type Request struct { // Nội dung của requests
	Time    time.Time `json:"time"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	Room    string    `json:"room"`
	Message string    `json:"message"`
}

type receive struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

// khởi tạo channel và map để lưu trữ danh sách user và rooms
var clients = make(map[*websocket.Conn]string)
var rooms = make(map[string]map[string]*websocket.Conn)
var Msg = make(chan Request)
var MsgRoom = make(chan Request)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//returns database connection - kết nối với database postgresql
func GetDatabase() *sql.DB {
	err := godotenv.Load("./db.env")
	if err != nil {
		log.Println("could not load .env file", err)
	}
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DBNAME")
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	// mo ket noi voi db
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatalln(err, "invalid database url")
	}
	err = db.Ping()
	if err != nil {
		log.Println("ko ket noi duoc db: ", err)
	}
	//fmt.Println("Database connection successful")
	return db
}

//close database connection - đóng kết nối
func CloseDatabase(connection *sql.DB) {
	err := connection.Close()
	if err != nil {
		log.Fatal("Can't close database", err)
	}
	connection.Close()
}

// set error message in Error struct
func SetErrors(err Error, message string) Error {
	err.IsError = true
	err.Message = message
	return err
}

//take password as input and generate new hash password from it
func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// so sanh password voi password da dc ma hoa
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generate Jwt token - tạo token cho user mới login
func GenerateJWT(email string) (string, error) {
	var mySigningKey = []byte(secretkey)
	token := jwt.New(jwt.SigningMethodHS256)

	// khởi tạo thông tin cơ bản cho token
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Printf("Something wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

// --------------- MIDDLEWARE FUNCTION-------------

//check whether user is authorized or not - check sự hợp lệ của token
func IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection := GetDatabase()
		defer CloseDatabase(connection)
		if r.Header["Token"] == nil {
			var err Error
			err = SetErrors(err, "No token found")
			json.NewEncoder(w).Encode(err)
		}

		var mySigningKey = []byte(secretkey)

		token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing token.")
			}
			return mySigningKey, nil
		})

		if err != nil {
			var err Error
			err = SetErrors(err, "your Token has been expired")
			json.NewEncoder(w).Encode(err)
		}
		result := connection.QueryRow("select email from user_info where token = $1", r.Header["Token"][0])

		var email string
		result.Scan(&email)

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["email"] == email {
				r.Header.Set("email", email)
				r.Header.Set("Token", r.Header["Token"][0])

			}
			handler.ServeHTTP(w, r)
			return
		}
		var reserr Error
		reserr = SetErrors(reserr, "Not authorized.")
		json.NewEncoder(w).Encode(err)
	}
}

//-------------	ROUTES	---------------------
//create a mux router
func CreateRouter() {
	router = mux.NewRouter()
}

//initialize all routes
func InitializeRoutes() {
	router.HandleFunc("/signup", SignUp).Methods("POST")
	router.HandleFunc("/signin", SignIn).Methods("POST")
	router.Handle("/enterzoom", IsAuthorized(enterZoom)).Methods("POST")
	router.HandleFunc("/createzoom", IsAuthorized(createZoom)).Methods("POST")
	//nhan tin
	router.HandleFunc("/recieveMss", IsAuthorized(RecieveMessage))
	router.HandleFunc("/logout", IsAuthorized(logout)).Methods("GET")
	router.HandleFunc("/", Index).Methods("GET")

}

// start the server
func ServerStart() {
	fmt.Println("Server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Access-Control-Allow-Origin", "Content-Type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router))
	if err != nil {
		log.Fatal(err)
	}
}

//-----------------ROUTES HANDLER-------------------------
func SignUp(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)

	// nhận thông tin đăng kí user
	var user Users
	err := json.NewDecoder(r.Body).Decode(&user)
	//log.Println(user)
	if err != nil {
		var err Error
		err = SetErrors(err, "error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	//kiểm tra sự tồn tại của user trong db
	var dbuser Users
	er := connection.QueryRow("select user_id,name,password,email from user_info where name=$1;", user.Name)
	er.Scan(&dbuser.Id, &dbuser.Name, &dbuser.Password, &dbuser.Email)
	log.Println("signup ")
	if dbuser.Email != "" {
		var err Error
		err = SetErrors(err, "Email already in use")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	user.Password, err = GeneratehashPassword(user.Password)
	if err != nil {
		log.Fatalln("Error in password hashing.")
	}

	if _, err := connection.Exec("insert into user_info (name,password,email,created_on) values($1,$2,$3,$4)", user.Name, user.Password, user.Email, time.Now()); err != nil {
		log.Fatal(err, "can not insert into user_info")
		json.NewEncoder(w).Encode(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

//đăng nhập
func SignIn(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)

	var authDetails Authentication

	// tiếp nhận thông tin đăng nhập
	err := json.NewDecoder(r.Body).Decode(&authDetails)
	//log.Println("authDetails: ", authDetails)
	if err != nil {
		var err Error
		err = SetErrors(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	// kiểm tra trong db
	var authUser Users
	er := connection.QueryRow("select name,password,email from user_info where name=$1;", authDetails.Name)
	er.Scan(&authUser.Name, &authUser.Password, &authUser.Email)
	log.Println("auth: ")
	if authUser.Email == "" {
		var err Error
		err = SetErrors(err, "Username or password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	check := CheckPasswordHash(authDetails.Password, authUser.Password)
	if !check {
		var err Error
		err = SetErrors(err, "Username or password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	validToken, err := GenerateJWT(authUser.Email)
	if err != nil {
		var err Error
		err = SetErrors(err, "faile to generate token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var token Token
	token.Email = authUser.Email
	token.TokenString = validToken

	// lưu token trong db
	if _, err := connection.Exec("update user_info set token=$1 ,last_login=$2 where email = $3", token.TokenString, time.Now(), token.Email); err != nil {
		var err Error
		err = SetErrors(err, "can not insert into user_info")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)

}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home public index page"))
}

func RecieveMessage(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	connection := GetDatabase()
	defer CloseDatabase(connection)

	var msg Request
	err = ws.ReadJSON(&msg)
	if err != nil {
		log.Printf("error: %v", err)
		for room, _ := range rooms {
			delete(rooms[room], msg.From)
		}
		delete(clients, ws)
		log.Println("account:", clients)
		log.Println("account in room:", rooms)

	}
	clients[ws] = r.Header.Get("email")

	//lay id cua user
	er := connection.QueryRow("select user_id from user_info where email=$1 ", r.Header.Get("email"))
	var user_id int
	er.Scan(&user_id)

	//------------------------------------------------------------ Tuong tac giua cac User------------------------------------------------------------------------------
	// luu dia chi user trong map
	for room, _ := range rooms {

		if room == msg.Room {

			rooms[room][r.Header.Get("email")] = ws
			log.Println("success!!!!!!!!!!!")

		} else if room != msg.Room {
			for user, _ := range members[room] {
				if user == r.Header.Get("email") {
					rooms[room][r.Header.Get("email")] = ws
					log.Println("success!!!")
				}
			}
		}
	}
	log.Println("checkRoom: ", rooms)

	// doc tat ca tin nhan cu gui cho user
	results, err := connection.Query("select time,sender_id, receiver_id,room_id,message from messages where receiver_id = $1", user_id)
	if err != nil {
		log.Println("ko doc duoc tin nhan cu", err)
	}
	defer results.Close()
	for results.Next() {
		var message Request
		err = results.Scan(&message.Time, &sender_id, &receiver_id, &room_id, &message.Message)
		room := connection.QueryRow("select room_name from room_info where room_id = $1", room_id)
		room.Scan(&message.Room)
		sender := connection.QueryRow("select email from user_info where user_id = $1", sender_id)
		sender.Scan(&message.From)
		receiver := connection.QueryRow("select user_id from user_info where email = $1", receiver_id)
		receiver.Scan(&message.To)
		if err != nil {
			// handle this error
			panic(err)
		}
		//log.Println("msg from: ", stored)

		err = ws.WriteJSON(message)
		if err != nil {
			log.Println("ko gui dc tin nhan: ", err)
		}
	}
	if _, err := connection.Exec("delete from messages where receiver_id = $1", user_id); err != nil {
		log.Fatal("can not delete message in messages :", err)
	}
	//------------------------------------------------------------ Ket thuc tuong tac giua cac User------------------------------------------------------------------------------

	//---------------------------------------------------------- Tuong tac trong room --------------------------------------------------------------------------------
	// thong bao user join room
	result := connection.QueryRow("select room_name from room_info where room_name = $1", msg.Room)
	var room_name string
	result.Scan(&room_name)
	for room, _ := range rooms {
		if room == room_name {
			for user, wss := range rooms[room] {
				if user != r.Header.Get("email") && members[room][r.Header.Get("email")] == false {
					var ms Request
					ms.Time = time.Now()
					ms.From = "system"
					ms.Room = room
					ms.Message = r.Header.Get("email") + " has entered in " + room_name + "\n"
					err := wss.WriteJSON(ms)
					if err != nil {
						log.Printf("error: %v", err)
						ws.Close()
					}
				}
				members[room][user] = true // xac nhan user da join vao room
				log.Println("thanh cong !!")
			}
		}
	}
	//---------------------------------------------------------- Ket thuc uong tac trong room --------------------------------------------------------------------------------

	go func() {
		for {
			msg := <-Msg
			connection := GetDatabase()
			defer CloseDatabase(connection)
			sender := connection.QueryRow("select user_id from user_info where email = $1", msg.From)
			sender.Scan(&sender_id)
			result := connection.QueryRow("select room_id,room_name from room_info where room_name = $1", msg.Room)
			result.Scan(&room_id, &room_name)
			log.Println("room_name: ", room_name)
			if room_name == "public" { // luu tin nhan ca nhan
				log.Println(clients)
				receiver := connection.QueryRow("select user_id from user_info where email = $1", msg.To)
				receiver.Scan(&receiver_id)
				if _, err := connection.Exec("insert into messages(time,sender_id,receiver_id,room_id,message) values($1,$2,$3,$4,$5)", msg.Time, sender_id, receiver_id, 0, msg.Message); err != nil {
					log.Println("can not update messages: ", err)
				}
			} else if room_name != "public" {
				// luu tin nhan room cho tung email
				for user, _ := range members[room_name] {
					receiver := connection.QueryRow("select user_id from user_info where email = $1", user)
					receiver.Scan(&receiver_id)
					if _, err := connection.Exec("insert into messages  (time, sender_id, receiver_id, room_id,message) values($1,$2,$3,$4,$5)", msg.Time, sender_id, receiver_id, room_id, msg.Message); err != nil {
						log.Println("fail to insert message: ", err)
					}
				}
			}

			log.Println("class: ", rooms)
			//gui tin nhan ca nhan
			for client, email := range clients {
				if msg.To == email {
					err := client.WriteJSON(msg)
					if err != nil {
						log.Printf("error: %v", err)
						client.Close()
						delete(clients, client)
					}
					if _, err := connection.Exec("delete from messages where receiver_id = $1 and room_id=0", receiver_id); err != nil {
						log.Fatal("can not delete message in messages :", err)
					}
				}
			}

			// Grab the next message from the Msg channel
			for room, _ := range rooms {
				if room == msg.Room {
					// gui mss toi tat ca users
					for _, wss := range rooms[room] {
						err := wss.WriteJSON(msg)
						if err != nil {
							log.Printf("error: %v", err)
						}
					}
					// xoa mss cua nhung users dang ol
					for email, _ := range rooms[room] {
						for check, _ := range members[msg.Room] {
							if email == check {
								receiver := connection.QueryRow("select user_id from user_info where email = $1", check)
								receiver.Scan(&receiver_id)
								if _, err := connection.Exec("delete from messages where receiver_id = $1 and room_id=$2", receiver_id, room_id); err != nil {
									log.Fatal("can not update message in messages :", err)
								}
							}
						}
					}
				}
				log.Println("Thanh Cong")
			}
			room_name = ""
		}
	}()

	// Read in a new message as JSON and map it to a Message object
	for {
		var msg Request
		msg.From = r.Header.Get("email")
		msg.Time = time.Now()
		msg.Room = "public"
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			for room, _ := range rooms {
				delete(rooms[room], msg.From)
			}
			delete(clients, ws)
			log.Println("account:", clients)
			log.Println("account in room:", rooms)
			break
		}
		Msg <- msg
	}
}

// logout
func logout(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	if _, err := connection.Exec("update user_info set token='' where email = $1", r.Header.Get("email")); err != nil {
		var err Error
		err = SetErrors(err, "can not delete token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r.Header.Get("email"))

}

// tạo room
func createZoom(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	var user zoom
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		var err Error
		err = SetErrors(err, "error in reading request-payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	er, err := connection.Query("select room_name from room_info where room_name=$1 ", user.Class)
	er.Scan(&check)
	//log.Println("check: ", check)
	defer er.Close()
	if check == user.Class {
		w.Write([]byte(" \n room already existed .......  "))
		check = ""
	} else {
		rooms[user.Class] = make(map[string]*websocket.Conn)
		members[user.Class] = make(map[string]bool)
		er := connection.QueryRow("select user_id from user_info where email=$1 ", r.Header.Get("email"))
		var user_id int
		er.Scan(&user_id)
		mss := "\t\t* * *" + user.Class + " has been created by " + r.Header.Get("email") + "* * *"
		if _, err := connection.Exec("insert into room_info (room_name,time_created,creater) values($1,$2,$3)", user.Class, time.Now(), user_id); err != nil {
			log.Println("can not insert into room_info: ", err)
		}

		er = connection.QueryRow("select room_id from room_info where room_name=$1 ", user.Class)
		var room_id int
		er.Scan(&room_id)

		if _, err := connection.Exec("insert into messages (time,sender_id,room_id,message) values($1,0,$2,$3)", time.Now(), room_id, mss); err != nil {
			log.Println("can not insert into messages: ", err)
		}

		if _, err := connection.Exec("insert into members (user_id,room_id,time_apply) values($1,$2,$3)", user_id, room_id, time.Now()); err != nil {
			var err Error
			err = SetErrors(err, "can not insert into messages")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(err)
		}

		//ghi file
		file, err := os.OpenFile("room.txt", os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		defer file.Close()
		fmt.Println(user.Class)
		if _, err := file.WriteString(user.Class + "\n"); err != nil {
			panic(err)
		}
		log.Println("ok!!!")
		check = ""
		w.Write([]byte(r.Header.Get("email")))
		w.Write([]byte(" has created zoom "))
		w.Write([]byte(user.Class))
	}
}

// join room
func enterZoom(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	var user zoom
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		var err Error
		err = SetErrors(err, "error in reading request-payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	fmt.Println("entering:   ", user.Class)
	members[user.Class][r.Header.Get("email")] = false
	//tin nhan he thong
	er := connection.QueryRow("select user_id from user_info where email=$1 ", r.Header.Get("email"))
	var user_id int
	er.Scan(&user_id)
	mss := r.Header.Get("email") + " has entered in " + user.Class + "\n"
	result := connection.QueryRow("select message from messages where message = $1", mss)
	result.Scan(&checkMembers)
	if checkMembers == "" {
		er := connection.QueryRow("select room_id from room_info where room_name=$1 ", user.Class)
		var room_id int
		er.Scan(&room_id)
		if _, err := connection.Exec("insert into messages (message,sender_id,room_id,time) values($1,0,$2,$3)", mss, room_id, time.Now()); err != nil {
			log.Fatal("can not update in zoom_info :", err)
		}

		er = connection.QueryRow("select user_id from members where user_id=$1 and room_id=$2 ", user_id, room_id)
		var id int
		er.Scan(&id)
		if id == 0 {
			if _, err := connection.Exec("insert into members (room_id,user_id,time_apply) values($1,$2,$3)", room_id, user_id, time.Now()); err != nil {
				log.Fatal("can not update in members :", err)
			}
		}
		w.Write([]byte(user.Class))
		router.HandleFunc("/recieveMss", IsAuthorized(RecieveMessage))
		checkMembers = ""
	} else {
		w.Write([]byte(user.Class))
		router.HandleFunc("/recieveMss", IsAuthorized(RecieveMessage))
		checkMembers = ""
	}
}

func removeRoom(s []*websocket.Conn, str *websocket.Conn) {
	for i, value := range s {
		if value == str {
			s[i] = s[len(s)-1]
		}
	}
}

func restoreData() {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	results, err := connection.Query("select room_name, room_id from room_info ")
	if err != nil {
		log.Println(err)
	}
	defer results.Close()
	for results.Next() {
		var room string
		var room_id int
		err = results.Scan(&room, &room_id)
		if err != nil {
			log.Println("can not scan room: ", err)
		}
		members[room] = make(map[string]bool)
		rooms[room] = make(map[string]*websocket.Conn)
		result, err := connection.Query(`select user_id from members where room_id =$1 `, room_id)
		if err != nil {
			log.Println(err)
		}
		// khoi phuc danh sach member
		defer result.Close()
		for result.Next() {
			var user_id int
			err = result.Scan(&user_id)
			if err != nil {
				log.Println("can not scan user_id: ", err)
			}
			user := connection.QueryRow(`select email from user_info where user_id=$1`, user_id)
			var email string
			err = user.Scan(&email)
			members[room][email] = true
		}
	}
	log.Println("members: ", members)
}

func InitializeRoom() { //khoi tao room ban dau
	file, err := os.Open("room.txt")
	if err != nil {
		fmt.Println("Error: ", err)
	}
	defer file.Close()
	connection := GetDatabase()
	defer CloseDatabase(connection)

	result := connection.QueryRow("select name from user_info where name = 'system'")
	result.Scan(&checkMembers)
	if checkMembers == "" {
		if _, err := connection.Exec("insert into user_info (user_id,name,created_on) values(0,'system',$1)", time.Now()); err != nil {
			log.Fatal("can not insert  in user_info :", err)
		}
		if _, err := connection.Exec("insert into room_info values($1,$2,$3,$4)", 0, "public", time.Now(), 0); err != nil {
			log.Fatal("can not insert  in user_info :", err)
		}
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		str := connection.QueryRow("select room_name from room_info where room_name=$1", scanner.Text())
		var room string
		str.Scan(&room)
		if room != scanner.Text() {
			members[scanner.Text()] = make(map[string]bool)
			if _, err := connection.Exec("insert into room_info (creater,room_name,time_created) values(0,$1,$2)", scanner.Text(), time.Now()); err != nil {
				log.Fatal("can not insert  in user_info :", err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
func main() {
	/* 	connection := GetDatabase()
	   	result := connection.QueryRow("select name from user_info where name = 'system'")
	   	result.Scan(&checkMembers)
	   	if checkMembers == "" {
	   		if _, err := connection.Exec("insert into user_info (user_id,name,created_on) values(0,'system',$1)", time.Now()); err != nil {
	   			log.Fatal("can not insert  in user_info :", err)
	   		}
	   		if _, err := connection.Exec("insert into room_info values($1,$2,$3,$4)", 0, "public", time.Now(), 0); err != nil {
	   			log.Fatal("can not insert  in user_info :", err)
	   		}
	   		CreateRouter()
	   		InitializeRoutes()
	   		InitializeRoom()
	   		ServerStart()
	   	} else { */
	GetDatabase()
	restoreData()
	//checkMembers = ""
	CreateRouter()
	InitializeRoutes()
	InitializeRoom()
	ServerStart()
	//}
}
