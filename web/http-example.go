package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Student struct {
	Id    int
	Name  string
	Age   int
	Class []int
}
type Students []Student

func main() {
	fmt.Println("My website")
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/about", AboutPage)
	http.HandleFunc("/api/music", MusicApiPage)
	http.HandleFunc("/api/student", StudentsApi)
	http.HandleFunc("/api/students", ListStudentsApi)
	http.ListenAndServe(":3333", nil)
}
func StudentsApi(w http.ResponseWriter, r *http.Request) {
	var student = Student{Id: 1, Name: "Diep", Age: 18, Class: []int{1, 2, 3}}
	json.NewEncoder(w).Encode(student)
}
func ListStudentsApi(w http.ResponseWriter, r *http.Request) {
	var listStudent = Students{
		Student{Id: 1, Name: "Diep", Age: 18, Class: []int{1, 2, 3}},
		Student{Id: 2, Name: "Duong", Age: 18, Class: []int{1, 2, 3}},
		Student{Id: 3, Name: "Hieu", Age: 18, Class: []int{1, 2, 3}},
		Student{Id: 4, Name: "Dong", Age: 18, Class: []int{1, 2, 3}},
	}
	json.NewEncoder(w).Encode(listStudent)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>This is home page<h1>")
}
func AboutPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>This is about page<h1>")
}
func MusicApiPage(w http.ResponseWriter, r *http.Request) {
	var data = map[string]interface{}{
		"name": "Loi xin loi cua 1 dan choi",
		"casi": "Duy Manh",
	}
	json.NewEncoder(w).Encode(data) // newEncoder( writer).Encode( 1 kieu du lieu bat ki)
}
