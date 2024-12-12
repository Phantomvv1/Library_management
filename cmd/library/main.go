package main

import (
	. "github.com/Phantomvv1/Library_management/internal/authentication"
	. "github.com/Phantomvv1/Library_management/internal/books"
	. "github.com/Phantomvv1/Library_management/internal/librarians"
	. "github.com/Phantomvv1/Library_management/internal/users"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/users", GetUsers)
	r.GET("/books", GetBooks)
	r.GET("/librarians", GetLibrarians)
	r.POST("/librarians", CreateLibrarian)
	r.POST("/book", AddBook)
	r.POST("/log", LogIn)

	r.Run(":42069")
}
