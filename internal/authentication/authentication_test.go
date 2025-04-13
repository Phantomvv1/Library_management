package authentication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSignUp(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/signup", SignUp)

	rr := httptest.NewRecorder()

	jsonBody := []byte(`{"name": "Some_name", "email": "random_email@gmail.com", "password": "password", "type": "librarian"}`)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/signup", reader)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr.Body)
		os.Exit(1)
	}
}

var Token = ""

func TestLogIn(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/login", LogIn)

	rr := httptest.NewRecorder()

	jsonBody := []byte(`{"email": "random_email@gmail.com", "password": "password"}`)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/login", reader)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr.Body)
		os.Exit(1)
	}

	var token map[string]string
	json.NewDecoder(rr.Body).Decode(&token)
	Token = token["token"]
}

func TestGetCurrentProfile(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/profile", GetCurrentProfile)

	rr := httptest.NewRecorder()

	body := fmt.Sprintf(`{"token": "%s"}`, Token)
	jsonBody := []byte(body)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/profile", reader)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr.Body)
		os.Exit(1)
	}
}

func TestDeleteAccount(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.DELETE("/user", DeleteAccount)

	rr := httptest.NewRecorder()

	body := fmt.Sprintf(`{"email": "random_email@gmail.com", "token": "%s"}`, Token)
	jsonBody := []byte(body)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:42069/user", reader)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		log.Println(rr.Body)
		os.Exit(1)
	}
}
