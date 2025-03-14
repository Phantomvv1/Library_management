package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Review struct {
	Stars   float32 `json:"stars"`
	Comment string  `json:"comment"`
	BookID  int     `json:"bookID"`
}

func (r Review) validNumberOfStarts() bool {
	switch r.Stars {
	case 0:
		return true
	case 0.5:
		return true
	case 1:
		return true
	case 1.5:
		return true
	case 2:
		return true
	case 2.5:
		return true
	case 3:
		return true
	case 3.5:
		return true
	case 4:
		return true
	case 4.5:
		return true
	case 5:
		return true

	default:
		return false
	}
}

func CreateReviewsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists reviews (id serial primary key, user_id int, book_id, stars numeric, comment text)")
	if err != nil {
		return err
	}

	return nil
}

func LeaveReview(c *gin.Context) { //to be tested
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // stars && comment && token && bookID

	token, ok := information["token"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	id, _, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	review := Review{}
	stars, ok := information["stars"].(float64)
	if !ok {
		log.Println("Stars are not the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error stars are not the correct type"})
		return
	}

	review.Stars = float32(stars)
	if !review.validNumberOfStarts() {
		log.Println("Incorrect number of stars")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrect number of stars"})
		return
	}

	review.Comment, ok = information["comment"].(string)
	if !ok {
		log.Println("The comment is not correctly provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error the comment is not correctly provided"})
		return
	}

	bookID, ok := information["bookID"].(float64)
	if !ok {
		log.Println("The id of the book you are leaving a review to is not correctly provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error the id of the book you are leaving a review to is not correctly provided"})
		return
	}
	review.BookID = int(bookID)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	if err = CreateReviewsTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to create a table for reviews"})
		return
	}

	hasBorrowedThatBook := false
	err = conn.QueryRow(context.Background(), fmt.Sprint(`select hasborrowed from
				(
				select b.id, a.history, (b.title = ANY (a.history)) as hasBorrowed
				from books b join authentication a on b.id = a.id
				)
				where id = $1`), review.BookID).Scan(&hasBorrowedThatBook)

	if !hasBorrowedThatBook {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error you are unable to leave reviews on books that you haven't borrowed"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into reviews (user_id, book_id , stars, comment) values ($1, $2, $3, $4)", id, review.BookID, review.Stars, review.Comment)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to put your review information in the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func DeleteReview(c *gin.Context) { // to be tested
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && book_id

	token, ok := information["token"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	id, _, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	bookIDString, ok := information["bookID"].(float64)
	if !ok {
		log.Println("Incorrectly provided the id to which you have left a review")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided the id to which you have left a review"})
		return
	}
	bookID := int(bookIDString)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	check := 0
	err = conn.QueryRow(context.Background(), "delete from reviews where user_id = $1 and book_id = $2 returning id", id, bookID).Scan(&check)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error there is no review left by this user on this book"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information from the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
