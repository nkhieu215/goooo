package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

type Tokens struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

type users struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

type signin struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

var acct = make(chan users)
var token Tokens
var rec string

func request() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		var str string
		var msg users
		for {
			//fmt.Print("Write your message: ")
			fmt.Scan(&str)
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				fmt.Println(err)
			}
			acct <- msg
		}
	}()
	tickerValue := time.NewTicker(time.Second)
	for {
		select {
		case <-tickerValue.C:
			url := "http://localhost:8080/admin"
			method := "GET"
			payload := strings.NewReader(``)

			client := &http.Client{}
			req, err := http.NewRequest(method, url, payload)

			if err != nil {
				fmt.Println(err)
				return
			}
			req.Header.Add("Token", token.TokenString)

			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			if string(body) != "" {
				fmt.Println(string(body))
			}
		case msg2 := <-acct:
			url := "http://localhost:8080/user"
			method := "POST"
			ArrayByte, err := json.Marshal(msg2)
			if err != nil {
				fmt.Println(err)
			}
			payload := strings.NewReader(string(ArrayByte))
			client := &http.Client{}
			req, err := http.NewRequest(method, url, payload)

			if err != nil {
				fmt.Println(err)
				return
			}
			req.Header.Add("Token", token.TokenString)
			req.Header.Add("Content-Type", "application/json")

			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(body))
		case <-interrupt: // dung de client co the tu dong out khoi server
			log.Println("interrupt")
			close(done)
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}

}

func main() {
	clientSignin()
	if token.TokenString != "" {
		request()
	}
}

func clientSignin() {
	var str string
	fmt.Print("Enter your name and password: ")
	fmt.Scan(&str)

	url := "http://localhost:8080/signin"
	method := "POST"

	payload := strings.NewReader(str)

	client := &http.Client{}

	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("Signin success: ", string(body))
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Welcome ", token.Email)
}
