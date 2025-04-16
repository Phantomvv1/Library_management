package librarians

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

func TestGetLibrarians(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/librarians", GetLibrarians)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:42069/librarians", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

var Token = ""

func TestCreateEvent(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/event", CreateEvent)
	router.POST("/login", authentication.LogIn)

	rr := httptest.NewRecorder()

	jsonBody := []byte(`{"email": "kris@kris.com", "password": "passowrd"}`)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/login", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}

	var token map[string]string
	json.NewDecoder(rr.Body).Decode(&token)
	Token = token["token"]

	eventRR := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"name": "Discussion", "start": "2025-05-30T19:30:00Z", "token": "%s"}`, Token))
	eventReader := bytes.NewReader(body)

	evReq, err := http.NewRequest(http.MethodPost, "http://localhost:42069/event", eventReader)
	if err != nil {
		t.Fatal(err)
	}
	defer eventRR.Result().Body.Close()

	router.ServeHTTP(eventRR, evReq)

	if eventRR.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestInviteToEvent(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/event/invite", InviteToEvent)

	rr := httptest.NewRecorder()

	jsonBody := []byte(fmt.Sprintf(`{"email": "kris@kris.com", "token": "%s", "eventId": 9}`, Token))
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/event/invite", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusConflict { // invited successfuly or the person has already been invited
		t.Fatal(rr.Body)
	}
}

func TestGetInvited(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/event/invited", GetInvited)

	rr := httptest.NewRecorder()

	jsonBody := []byte(fmt.Sprintf(`{"id": 3, "token": "%s"}`, Token))
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/event/invited", reader)
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

func TestGetEvents(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/events", GetEvents)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:42069/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}

	// log.Println(rr.Body)
}

func TestGetUserHistory(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/user/history", GetUserHistory)

	rr := httptest.NewRecorder()

	jsonBody := []byte(fmt.Sprintf(`{"token": "%s"}`, Token))
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/user/history", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetUpcomingEvents(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/event/upcoming", GetUpcomingEvents)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:42069/event/upcoming", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}

	// log.Println(rr.Body)
}
