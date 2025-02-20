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
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
	information := make(map[string]string)
	json.NewDecoder(c.Request.Body).Decode(&information) //name, email

	id, _, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	var name, email, accountType string
	err = conn.QueryRow(context.Background(), "select name, email, type from authentication a where a.id = $1", id).Scan(&name, &email, &accountType)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting information from the database"})
		return
	}

	createNewToken := false
	if name != information["name"] && information["name"] != "" {
		name = information["name"]
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error invalid new name"})
		return
	}

	if email != information["email"] && information["email"] != "" {
		email = information["email"]
		createNewToken = true
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error invalid new email"})
		return
	}

	_, err = conn.Exec(context.Background(), "update authentication set name = $1, email = $2 where id = $3", name, email, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the information in the databse"})
		return
	}

	if createNewToken {
		token, err := GenerateJWT(id, accountType, email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}

	c.JSON(http.StatusOK, nil)
}
