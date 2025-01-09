package users

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    int    `json:"id"`
}

func GetUsers(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "create table if not exists authentication (id serial primary key not null, name text, email text, password text, type text, history text);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a table for authentication"})
		return
	}

	var userList []User
	rows, err := conn.Query(context.Background(), "select id, email, name from authentication where type = 'user';")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't read from the database"})
		return
	}

	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process users"})
			return
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	if userList == nil {
		log.Println("No users created yet!")
		c.JSON(http.StatusNotFound, gin.H{"error": "No users have been created yet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
}

func EditProfile(c *gin.Context) {
	var edit User
	json.NewDecoder(c.Request.Body).Decode(&edit) //name, email

	if edit.Name != CurrentPrfile.Name {
		CurrentPrfile.Name = edit.Name
	}
	if edit.Email != "" {
		CurrentPrfile.Email = edit.Email
	}
}
