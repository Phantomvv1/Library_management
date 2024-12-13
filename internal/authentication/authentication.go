package authentication

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Profile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Type  string `json:"type"`
}

var CurrentPrfile Profile

func SHA512(text string) string {
	algorithm := sha512.New()
	algorithm.Write([]byte(text))
	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func SignUp(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABSE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Databse connection failed"})
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //name, email, password, type

	_, err = conn.Exec(context.Background(), "create table if not exists authentication (id primary key serial not null, name text, email text, password text, type text);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a table for authentication"})
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
		}
	}

	if emailExists {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "There is already a person with this email"})
	}

	hashedPassword := SHA512(information["password"])
	_, err = conn.Exec(context.Background(), "insert into authentication (name, email, password, type) values ($1, $2, $3);",
		information["name"], information["email"], hashedPassword, information["type"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting the information into the database."})
	}

	c.JSON(http.StatusOK, nil)
}

func LogIn(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABSE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //email, password

	var passwordCheck, name, email, typeOfAccount string
	err = conn.QueryRow(context.Background(), "select password, name, type, email from authentication a where a.email = $1;", information["email"]).Scan(&passwordCheck, &name, &typeOfAccount, &email)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "There isn't anybody registered with this email!"})
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while trying to log in"})
		}
	}

	if SHA512(information["password"]) != passwordCheck {
		log.Println("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
	}

	CurrentPrfile.Name = name
	CurrentPrfile.Email = email
	CurrentPrfile.Type = typeOfAccount
	c.JSON(http.StatusOK, nil)
}
