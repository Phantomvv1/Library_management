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
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type Profile struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Type    string   `json:"type"`
	History []string `json:"history"`
}

func GenerateJWT(id int, accountType string, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":         id,
		"type":       accountType,
		"email":      email,
		"expiration": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := os.Getenv("JWT_KEY")
	return token.SignedString([]byte(jwtKey))
}

func ValidateJWT(tokenString string) (int, string, error) {
	claims := &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnsupported
		}

		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil || !token.Valid {
		return 0, "", err
	}

	expiration, ok := (*claims)["expiration"].(float64)
	if !ok {
		return 0, "", errors.New("Error parsing the expiration date of the token")
	}

	if int64(expiration) < time.Now().Unix() {
		return 0, "", errors.New("Error token has expired")
	}

	id, ok := (*claims)["id"].(float64)
	if !ok {
		return 0, "", errors.New("Incorrect type of id")
	}

	accountType, ok := (*claims)["type"].(string)
	if !ok {
		return 0, "", errors.New("Incorrect type of account")
	}

	return int(id), accountType, nil
}

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

	validEmail, err := regexp.MatchString(".*@.*", information["email"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Error validating the email"})
		return
	}

	if !validEmail {
		log.Println("Invalid email")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error invalid email"})
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

	jwtToken, err := GenerateJWT(id, typeOfAccount, email)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while generating your token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}

func GetCurrentProfile(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	var tokenString map[string]string
	json.NewDecoder(c.Request.Body).Decode(&tokenString)

	var id int
	var accountType string
	id, accountType, err = ValidateJWT(tokenString["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error validating the token"})
		return
	}

	var name, email string
	var history []string
	err = conn.QueryRow(context.Background(), "select name, email, history from authentication where id = $1", id).Scan(&name, &email, &history)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting information from the database"})
		return
	}

	UserProfile := Profile{
		ID:      id,
		Name:    name,
		Email:   email,
		Type:    accountType,
		History: history,
	}

	c.JSON(http.StatusOK, gin.H{"profile information": UserProfile})
}
