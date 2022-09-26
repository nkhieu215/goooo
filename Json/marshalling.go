// marshalling cho phep chugn ta convert tu Go object sang Json strings
// tai lieu tham khao https://blog.vietnamlab.vn/cach-xu-li-json-trong-golang/
package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func main() {
	user := User{Id: "ID001", Name: "LanKa", Password: "123456"}
	ArrayByte, err := json.Marshal(user)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(ArrayByte))

}
