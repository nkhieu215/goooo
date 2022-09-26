package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Age    int    `json:"Age"`
	Social Social `json:"social"`
}
type Social struct {
	Facebook string `json:"facebook"`
	Twitter  string `json:"twitter"`
}

func main() {
	social := Social{Facebook: "https://facebook.com", Twitter: "https://twitter.com"}
	user := User{Name: "LanKa", Type: "Author", Age: 25, Social: social}
	byteArray, err := json.MarshalIndent(user, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(byteArray))
}
