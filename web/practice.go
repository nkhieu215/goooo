package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "", 0) // bien ghi nhat ki
func showInfoHandler(wr http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	wr.Write([]byte("Nguyen Long \n 18 tuoi \n Ha Noi"))
	timeElapsed := time.Since(timeStart)
	fmt.Println("Info: ", timeElapsed)
}
func showEmailHandler(wr http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	wr.Write([]byte("LongNguyen@gmail.com"))
	timeElapsed := time.Since(timeStart)
	fmt.Println("Email: ", timeElapsed)
}
func showFriendsHandler(wr http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	wr.Write([]byte("Long's friends is Nam and Lien"))
	timeElapsed := time.Since(timeStart)
	fmt.Println("Friends: ", timeElapsed)
}
func helloHandler(wr http.ResponseWriter, r *http.Request) {
	wr.Write([]byte("Hello"))
} // đây là một hàm thực hiện việc đóng gói hàm truyền vào
// để ghi nhận thời gian thực thi service
// hàm này được xem như là một Middleware
func timeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
		// ghi nhận thời gian trước khi chạy
		timeStart := time.Now()
		// next là hàm business logic được truyền vào
		next.ServeHTTP(wr, r)
		// tinh toan thoi gian thuc thi
		timeElapsed := time.Since(timeStart)
		logger.Println(timeElapsed)
	})
}

// Trong đó, http.Handler:
// type Handler interface {
//    ServeHTTP(ResponseWriter, *Request)
// }
func main() {
	http.Handle("/", timeMiddleware(http.HandlerFunc(helloHandler)))
	http.HandleFunc("/info", showInfoHandler)
	http.HandleFunc("/email", showEmailHandler)
	http.HandleFunc("/friend", showFriendsHandler)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
	}
}
