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

type LibrariansCreate struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func GetLibrarians(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer conn.Close(context.Background())

	var userList []Librarians
	rows, err := conn.Query(context.Background(), "select id, email, name from authentication where type = 'librarian';")
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNoContent, gin.H{"error": "There are no librarians"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't read from the database"})
			return
		}
	}

	for rows.Next() {
		var user Librarians
		err = rows.Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process librarians"})
			return
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch librarians"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}
