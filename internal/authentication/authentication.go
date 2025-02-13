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
	"regexp"
	"strconv"
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

var CurrentProfile Profile

var jwtKey string

func GenerateJWT(id int, accountType string, email string) (string, error) {
	claims := jwt.MapClaims{
		"id":         id,
		"type":       accountType,
		"email":      email,
		"expiration": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if jwtKey == "" {
		jwtKey = os.Getenv("JWT_KEY")
	}
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (int, string, error) {
	claims := &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.ErrUnsupported
		}

		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		log.Println(err)
		return 0, "", err
	}

	newClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("Error parsing the token")
	}

	var tokenExpiration int64
	if expiration, ok := newClaims["expiration"].(string); ok {
		tokenExpiration, err = strconv.ParseInt(expiration, 10, 64)
		if err != nil {
			return 0, "", err
		}
	} else {
		return 0, "", errors.New("Error parsing the expiration date of the token")
	}

	if tokenExpiration < time.Now().Unix() {
		return 0, "", errors.New("Error token has expired")
	}

	id, ok := newClaims["id"].(int)
	if !ok {
		return 0, "", errors.New("Incorrect type of id")
	}

	accountType, ok := newClaims["type"].(string)
	if !ok {
		return 0, "", errors.New("Incorrect type of account")
	}

	return id, accountType, nil
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

func GetCurrentProfile(c *gin.Context) { //Needs fixing in order to work with JWT
	if reflect.DeepEqual(CurrentProfile, Profile{}) {
		log.Println("You haven't logged in yet. There is no profile information.")
		c.JSON(http.StatusForbidden, gin.H{"error": "You haven't logged in yet. There is no profile information."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile information": CurrentProfile})
}
