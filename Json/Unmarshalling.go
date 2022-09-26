// nguoc cua marshalling convert du lieu
package main

import (
	"encoding/json"
	"fmt"
)

type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func main() {
	// truong hop dl co cau truc
	jsonString := `{"title":"Learning JSON in Golang","author":"Lanka"}` // chuoi string kieu json
	var book Book
	err := json.Unmarshal([]byte(jsonString), &book)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v \n", book)
	// truong hop du lieu ko co cau truc
	/* 	var book1 map[string]interface{} //kieu tra ve cua bien dc giai nen
	   	err := json.Unmarshal([]byte(jsonString), &book1)
	   	if err != nil {
	   		panic(err)
	   	}
	   	fmt.Printf("%+v\n", book1) */
	// cach khac
	/* var book2 interface{}
	err := json.Unmarshal([]byte(jsonString), &book2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", book2) */
}
