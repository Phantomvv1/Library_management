package reviews

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

func TestLeaveReview(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/review", LeaveReview)
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

	reviewRR := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"stars": 4, "comment": "This is a very good book", "bookID": 1, "token": "%s"}`, Token))
	eventReader := bytes.NewReader(body)

	revReq, err := http.NewRequest(http.MethodPost, "http://localhost:42069/review", eventReader)
	if err != nil {
		t.Fatal(err)
	}
	defer reviewRR.Result().Body.Close()

	router.ServeHTTP(reviewRR, revReq)

	if reviewRR.Code != http.StatusOK && reviewRR.Code != http.StatusForbidden {
		t.Fatal(rr.Body)
	}
}

func TestEditReview(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.PUT("/review", EditReview)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"bookID": 1, "stars": 5, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPut, "http://localhost:42069/review", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetReviewsForBook(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/review", GetReviewsForBook)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/review", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetReviewsOfUser(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/user/reviews", GetReviewsOfUser)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"userID": 1, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/user/reviews", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetBookRating(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/rating", GetBookRating)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/rating", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetHighestRatedReviews(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/review/high", GetHighestRatedReviews)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/review/high", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetLowestRatedReviews(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/review/low", GetLowestRatedReviews)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/review/low", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestVoteForReview(t *testing.T) { // fix this one
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/review/vote", VoteForReview)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"vote": "up", "reviewID": 1, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/review/vote", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusForbidden {
		t.Fatal(rr.Body)
	}
}

func TestGetVotesForReview(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/review/votes", GetVotesForReview)

	rr := httptest.NewRecorder()

	body := []byte(`{"reviewID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/review/votes", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestGetVotesForReviewSQL(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/review/votes/sql", GetVotesForReviewSQL)

	rr := httptest.NewRecorder()

	body := []byte(`{"reviewID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/review/votes/sql", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestRatingDetails(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/rating/details", RatingDetails)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/rating/details", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestRatingDetailsSQL(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/book/rating/details/sql", RatingDetailsSQL)

	rr := httptest.NewRecorder()

	body := []byte(`{"bookID": 1}`)
	reader := bytes.NewReader(body)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:42069/book/rating/details/sql", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}

func TestDeleteReview(t *testing.T) { // fix this one too
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.DELETE("/review", DeleteReview)

	rr := httptest.NewRecorder()

	body := []byte(fmt.Sprintf(`{"bookID": 1, "token": "%s"}`, Token))
	reader := bytes.NewReader(body)

	revReq, err := http.NewRequest(http.MethodDelete, "http://localhost:42069/review", reader)
	if err != nil {
		t.Fatal(err)
	}
	defer rr.Result().Body.Close()

	router.ServeHTTP(rr, revReq)

	if rr.Code != http.StatusOK {
		t.Fatal(rr.Body)
	}
}
