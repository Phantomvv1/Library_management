package main

import (
	. "github.com/Phantomvv1/Library_management/internal/users"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/users", GetUsers)

	r.Run(":42069")
}
