package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	//"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//------------GLOBAL VARIABLES--------------
var (
	router    *mux.Router
	secretkey string = "SecretKeyjwt!!!"
)

/* var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} */

//----------------Struct---------- khoi tao bien
type Users struct {
	gorm.Model
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type Token struct {
	Role        string `json:"role"`
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
type Request struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

//--------------DATABASE FUNCTIONS---------------
const (
	host     = "localhost"
	user     = "nguyenkhachieu"
	password = "anhiuem2"
	dbname   = "postgres"
)

//returns database connection
func GetDatabase() *sql.DB {
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	// mo ket noi voi db
	db, err := sql.Open("postgres", conn)
	if err != nil {
		log.Fatalln(err, "invalid database url")
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Database connected")
	}
	fmt.Println("Database connection successful")
	return db
}

//create user table in userdb
/* func InitialMigration() {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	connection.AutoMigrate(Users{})
} */

//close database connection
func CloseDatabase(connection *sql.DB) {
	err := connection.Close()
	if err != nil {
		log.Fatal("Can't close database", err)
	}
	connection.Close()
}

//--------------HELPER FUNCTIONS----------------

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

// generate Jwt token
func GenerateJWT(email, role string) (string, error) {
	var mySigningKey = []byte(secretkey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Errorf("Something wrong: %s", err.Error())
		return "", err
	}
	return tokenString, nil
}

// --------------- MIDDLEWARE FUNCTION-------------

//check whether user is authorized or not
func IsAuthorized(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == "admin" {
				r.Header.Set("role", "admin")
				handler.ServeHTTP(w, r)
				return
			} else if claims["role"] == "user" {
				r.Header.Set("role", "user")
				handler.ServeHTTP(w, r)
				return
			}
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
	router.HandleFunc("/admin", IsAuthorized(AdminIndex)).Methods("GET")
	router.HandleFunc("/user", IsAuthorized(UserIndex)).Methods("GET")
	//router.HandleFunc("/ws", IsAuthorized(Requests)).Methods("POST")
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

	var user Users
	err := json.NewDecoder(r.Body).Decode(&user)
	log.Println(user)
	if err != nil {
		var err Error
		err = SetErrors(err, "error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var dbuser Users
	er := connection.QueryRow("select * from login1 where name=$1;", user.Name)
	er.Scan(&dbuser.Id, &dbuser.Name, &dbuser.Password, &dbuser.Email, &dbuser.Role)
	log.Println("signup: ", dbuser)
	//check email is already registered or not
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
	log.Println(user)
	//insert user details in database
	if _, err := connection.Query("insert into login1 values($1,$2,$3,$4,$5)", user.Id, user.Name, user.Password, user.Email, user.Role); err != nil {
		var err Error
		err = SetErrors(err, "can not insert into table")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	connection := GetDatabase()
	defer CloseDatabase(connection)

	var authDetails Authentication

	err := json.NewDecoder(r.Body).Decode(&authDetails)
	log.Println("authDetails: ", authDetails)
	if err != nil {
		var err Error
		err = SetErrors(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var authUser Users
	//connection.Where("email =    ?", connection.Table("login1")).Where("email=?", authDetails.Email).Find(&authUser)
	er := connection.QueryRow("select * from login1 where name=$1;", authDetails.Name)
	er.Scan(&authUser.Id, &authUser.Name, &authUser.Password, &authUser.Email, &authUser.Role)
	log.Println("auth: ", authUser)
	if authUser.Email == "" {
		var err Error
		err = SetErrors(err, "Username or password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	check := CheckPasswordHash(authDetails.Password, authUser.Password)
	log.Println("auth: ", authUser.Password)
	log.Println("authDetails: ", authDetails.Password)
	if !check {
		var err Error
		err = SetErrors(err, "Username or password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	validToken, err := GenerateJWT(authUser.Email, authUser.Role)
	if err != nil {
		var err Error
		err = SetErrors(err, "faile to generate token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var token Token
	token.Email = authUser.Email
	token.Role = authUser.Role
	token.TokenString = validToken
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home public index page"))
}

func AdminIndex(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("role") != "admin" {
		w.Write([]byte("Not authorized."))
		return
	}
	w.Write([]byte("Welcome, Admin."))
}

func UserIndex(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Role") != "user" {
		w.Write([]byte("Not Authorized."))
		return
	}
	w.Write([]byte("Welcome, User."))
}

/* func Requests(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	var request Request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println("errrrr: ", err)
	}
	log.Println(request)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
	err := ws.
} */

func main() {
	GetDatabase()
	CreateRouter()
	InitializeRoutes()
	ServerStart()
}
