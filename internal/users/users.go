package users

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

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
	json.NewDecoder(c.Request.Body).Decode(&information) // (name || email) && token

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
	useName := true
	newName, ok := information["name"]
	if !ok {
		useName = false
	} else {
		name = newName
	}

	newEmail, ok := information["email"]
	if !ok && !useName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no new information provided"})
	} else if newEmail != "" {
		email = newEmail
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

func GetUserByID(c *gin.Context) {
	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information)

	_, accountType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error only librarians can get a certain user"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	params := c.Request.URL.Query()
	idString := params.Get("id")
	if idString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error id of the person you are searching for is missing"})
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to parse the id of the user"})
		return
	}

	var user Profile
	user.ID = id
	err = conn.QueryRow(context.Background(), "select name, email, history, type from authentication a where a.id = $1", id).Scan(&user.Name, &user.Email, &user.History, &user.Type)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error there is user with this id"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the information abut the user from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
