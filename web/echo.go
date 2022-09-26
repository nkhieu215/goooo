package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

// hàm routing echo, gồm hai params
// r *http.Request : dùng để đọc yêu cầu từ client
// wr http.ResponseWriter : dùng để ghi phản hồi về client
func echo(wr http.ResponseWriter, r *http.Request) {
	// đọc thông điệp mà client gửi tới trong r.Body
	msg, err := ioutil.ReadAll(r.Body)
	// phản hồi về client lỗi nếu có
	if err != nil {
		wr.Write([]byte("echo error"))
		return
	}
	// phản hồi về client chính thông điệp mà client gửi
	writeLen, err := wr.Write(msg)
	// nếu lỗi xảy ra, hoặc kích thước thông điệp phản hồi khác
	// kích thước thông điệp nhận được
	if err != nil || writeLen != len(msg) {
		log.Println(err, "write len:", writeLen)
	}
}
func main() {
	// mapping url ứng với hàm routing echo
	http.HandleFunc("/", echo)
	// địa chỉ http://127.0.0.1:8080/
	err := http.ListenAndServe(":8080", nil)
	// log ra lỗi nếu bị trùng port
	if err != nil {
		log.Fatal(err)
	}
}
