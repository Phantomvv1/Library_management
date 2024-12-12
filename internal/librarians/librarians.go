package librarians

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Librarians struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetLibrarians(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
	}
	defer conn.Close(context.Background())

	var userList []Librarians
	rows, err := conn.Query(context.Background(), "select id, email, name from librarians;")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't read from the database"})
	}

	for rows.Next() {
		var user Librarians
		err = rows.Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process librarians"})
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch librarians"})
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}
