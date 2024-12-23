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
	r.GET("/history", GetHistory)
	r.POST("/book", AddBook)
	r.POST("/signup", SignUp)
	r.POST("/login", LogIn)
	r.POST("/searchbook", SearchForBook)
	r.POST("/edit", EditProfile)
	r.POST("/book/borrow", BorrowBook)
	r.POST("/book/return", ReturnBook)
	r.POST("/book/reserve", ReserveBook)
	r.POST("/event", CreateEvent)
	r.POST("event/invite", InviteToEvent)

	r.Run(":42069")
}
