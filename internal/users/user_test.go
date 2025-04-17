package users

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
)

func TestGetUsers(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/users", GetUsers)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:42069/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}

	log.Println(rr.Body)
}

var Token = ""

func TestEditProfile(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/edit", EditProfile)
	router.POST("/login", authentication.LogIn)

	rrLogin := httptest.NewRecorder()

	jsonBody := []byte(`{"email": "kris@kris.com", "password": "passowrd"}`)
	readerL := bytes.NewReader(jsonBody)

	reqL, err := http.NewRequest(http.MethodPost, "http://localhost:42069/login", readerL)
	if err != nil {
		t.Fatal(err)
	}
	defer rrLogin.Result().Body.Close()

	router.ServeHTTP(rrLogin, reqL)

	if rrLogin.Code != http.StatusOK {
		t.Fatal(rrLogin.Body)
	}

	var token map[string]string
	json.NewDecoder(rrLogin.Body).Decode(&token)
	Token = token["token"]

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"name": "Kris", "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/edit", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetUserByID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/user", GetUserByID)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/user?id=1", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}

	log.Println(rr.Body)
}
