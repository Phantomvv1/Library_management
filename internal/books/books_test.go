package books

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
)

var Token = ""

func TestAddBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book", AddBook)
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

	bookRR := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"isbn": "978-3-16-148410-0",
    "title": "Some title",
    "author": "Some author",
    "year": 1990,
    "quantity": 10000, "token": "%s"}`, Token))
	bookReader := bytes.NewReader(body)

	revReq, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book", bookReader)
	if err != nil {
		t.Fatal(err)
	}
	defer bookRR.Result().Body.Close()

	router.ServeHTTP(bookRR, revReq)

	if bookRR.Code != http.StatusOK && bookRR.Code != http.StatusForbidden {
		t.Fatal(rr.Body)
	}
}

func TestGetBooks(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/books", GetBooks)

	rr := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:42069/books", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestSearchForBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/search", SearchForBook)

	rr := httptest.NewRecorder()

	body := []byte(`{"name": "Some"}`) // idk why it's name here instead of title but if it works i don't see the problem
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/search", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestBorrowBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/borrow", BorrowBook)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"title": "Some title", "returnDate": "2025-07-01", "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/borrow", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestReturnBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/return", ReturnBook)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"title": "Some title", "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/return", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetHistory(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/history", GetHistory)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/history", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestUpdateBookID(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/update/id", UpdateBookQuantity)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"id": 8, "title": "Some title", "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/update/id", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestUpdateBookQuantity(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/quantity", UpdateBookQuantity)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"id": 8, "quantity": 0, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/quantity", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestReserveBook(t *testing.T) { // TODO: Make a way to remove the reservation
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/reserve", ReserveBook)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"title": "Some title", "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/reserve", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestRemoveBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/remove", RemoveBook)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"id": 8, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/remove", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}
