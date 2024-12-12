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

	. "github.com/Phantomvv1/Library_management/internal/librarians"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var LoggedInAs string = ""

func SHA512(text string) string {
	algorithm := sha512.New()
	algorithm.Write([]byte(text))
	result := algorithm.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func SignUpLibrarian(librarian LibrariansCreate, conn *pgx.Conn) error {
	email := librarian.Email
	password := librarian.Password

	_, err := conn.Exec(context.Background(), "create table if not exists librarians (email text, password text);")
	if err != nil {
		log.Println(err)
		return err
	}

	var check string
	err = conn.QueryRow(context.Background(), "select email from librarians where email = $1;", email).Scan(&check)
	emailExists := true
	if err != nil {
		if err == pgx.ErrNoRows {
			emailExists = false
		} else {
			log.Println(err)
		}
	}

	if emailExists {
		return errors.New("Email already exists!")
	}

	hashedPassword := SHA512(password)
	_, err = conn.Exec(context.Background(), "insert into librarians (email, password) values ($1, $2);", email, hashedPassword)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func LogIn(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABSE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information)

	var passwordCheck string
	err = conn.QueryRow(context.Background(), "select password from librarians l where l.email = $1;", information["email"]).Scan(&passwordCheck)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "There is no librarian with this email!"})
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while trying to log in"})
		}
	}

	if SHA512(information["password"]) != passwordCheck {
		log.Println("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong password"})
	}

	LoggedInAs = information["type"]
	c.JSON(http.StatusOK, nil)
}
