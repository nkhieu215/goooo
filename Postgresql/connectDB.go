package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

//khai bao thong tin ket noi voi db
const (
	host     = "localhost"
	user     = "nguyenkhachieu"
	password = "anhiuem2"
	dbname   = "postgres"
)

func main() {
	// tao chuoi ket noi voi db, sslmode =disable de disable SSL
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	// mo ket noi voi db
	db, err := sql.Open("postgres", conn)
	if err != nil {
		fmt.Printf("fail to open DB: %v \n", err)
		return
	}
	defer db.Close()
	// goi ham ping() de kiem tra ket noi da thanh cong hay chua
	err = db.Ping()
	if err != nil {
		fmt.Printf("fail to connect DB: %v", err)
		return
	}
	fmt.Println("Ping ok")
	_sql := "select * from player;"
	row, err := db.Query(_sql)
	if err != nil {
		fmt.Printf("fail to query: %v \n", err)
		return
	}
	var col1 int
	var col2 string
	var col3 int
	var col4 string
	var col5 string
	for row.Next() {
		row.Scan(&col1, &col2, &col3, &col4, &col5)
		fmt.Printf("[value col1: %v,\tvalue col2: %v ,\tvalue col3: %v ,\tvalue col4: %v ,\tvalue col5: %v]\n", col1, col2, col3, col4, col5)
	}
	_sql1 := "delete from player where id = '0'"
	row1, err := db.Query(_sql1)
	if err != nil {
		fmt.Println("fail to query: %v \n", err)
	}
	for row1.Next() {
		row1.Scan(&col1, &col2, &col3, &col4, &col5)
		fmt.Printf("[value col1: %v,\tvalue col2: %v ,\tvalue col3: %v ,\tvalue col4: %v ,\tvalue col5: %v]\n", col1, col2, col3, col4, col5)
	}
	fmt.Println("End !!!")
}
