package authentication

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Type    string   `json:"type"`
	History []string `json:"history"`
}

var CurrentPrfile Profile

func SHA512(text string) string {
	algorithm := sha512.New()
	algorithm.Write([]byte(text))
	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func CreateAuthTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists authentication (id serial primary key, name text, email text, password text, type text, history text[]);")
	if err != nil {
		log.Println(err)
		return errors.New("Error creating a table for authentication")
	}

	return nil
}

func SignUp(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Databse connection failed"})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //name, email, password, type

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var check string
	err = conn.QueryRow(context.Background(), "select email from authentication where email = $1;", information["email"]).Scan(&check)
	emailExists := true
	if err != nil {
		if err == pgx.ErrNoRows {
			emailExists = false
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the password from the table"})
			return
		}
	}

	if emailExists {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "There is already a person with this email"})
		return
	}

	hashedPassword := SHA512(information["password"])
	_, err = conn.Exec(context.Background(), "insert into authentication (name, email, password, type, history) values ($1, $2, $3, $4, array[]::text[]);",
		information["name"], information["email"], hashedPassword, information["type"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting the information into the database."})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func LogIn(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //email, password

	var passwordCheck, name, email, typeOfAccount string
	var history []string
	var id int
	err = conn.QueryRow(context.Background(), "select password, name, type, email, history, id from authentication a where a.email = $1;", information["email"]).Scan(
		&passwordCheck, &name, &typeOfAccount, &email, &history, &id)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "There isn't anybody registered with this email!"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while trying to log in"})
			return
		}
	}

	if SHA512(information["password"]) != passwordCheck {
		log.Println("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
		return
	}

	CurrentPrfile.ID = id
	CurrentPrfile.Name = name
	CurrentPrfile.Email = email
	CurrentPrfile.Type = typeOfAccount
	CurrentPrfile.History = history
	c.JSON(http.StatusOK, nil)
}

func GetCurrentProfile(c *gin.Context) {
	if reflect.DeepEqual(CurrentPrfile, Profile{}) {
		log.Println("You haven't logged in yet. There is no profile information.")
		c.JSON(http.StatusForbidden, gin.H{"error": "You haven't logged in yet. There is no profile information."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile information": CurrentPrfile})
}
