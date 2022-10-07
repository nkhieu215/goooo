package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type book struct {
	ID       string `json:"id`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

var books = []book{
	{ID: "1", Title: "In Search of Lost Time", Author: "Marcel Proust", Quantity: 2},
	{ID: "2", Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Quantity: 5},
	{ID: "3", Title: "War and Peace", Author: "Leo Tolstoy", Quantity: 6},
}

func getBooks(c *gin.Context) {
	c.JSON(http.StatusOK, books) //hien du lieu duoi dang json
}

func createBook(c *gin.Context) {
	var newBook book
	if err := c.BindJSON(&newBook); err != nil {
		return
	} // lenh ket noi(tro) truc tiep vao newbook de sua cac truong ben trong

	books = append(books, newBook)
	c.IndentedJSON(http.StatusCreated, newBook) // tra laij trang thai ket qua tao newBook

}

func bookById(c *gin.Context) { // param: tham so duong dan
	id := c.Param("id")
	book, err := getBookById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"message": "Book not found",
		})
		return
	}
	c.IndentedJSON(http.StatusOK, book)
}
func getBookById(id string) (*book, error) {
	for i, b := range books {
		if b.ID == id {
			return &books[i], nil
		}
	}
	return nil, errors.New("book not found")
}

func checkoutBook(c *gin.Context) {
	// querry: tham so truy van
	id, ok := c.GetQuery("id")
	if ok == false {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "missing id query parameter.",
		})
		return
	}
	book, err := getBookById(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{
			"message": " Book not found",
		})
		return
	}
	if book.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "book not available.",
		})
	}
	book.Quantity -= 1
	c.IndentedJSON(http.StatusOK, book)
}

func returnBook(c *gin.Context) {
	id, ok := c.GetQuery("id")
	if ok == false {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "missing id query parameter.",
		})
		return
	}
	book, err := getBookById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": " Book not found",
		})
		return
	}
	book.Quantity += 1
	c.JSON(http.StatusOK, book)
}

func main() {
	router := gin.Default()
	router.GET("/book", getBooks)
	router.POST("/book", createBook)
	router.GET("/book/:id", bookById)
	router.PATCH("/checkout", checkoutBook)
	router.PATCH("/return", returnBook)
	router.Run("localhost:8080")
}
