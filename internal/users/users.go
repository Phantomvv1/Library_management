package users

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Users struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    int    `json:"id"`
}

func GetUsers(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
	}
	defer conn.Close(context.Background())

	var userList []Users
	rows, err := conn.Query(context.Background(), "select id, email, name from users;")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't read from the database"})
	}

	for rows.Next() {
		var user Users
		err = rows.Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process users"})
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.Println("Error occurred during row iteration:", rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}
