package reviews

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Review struct {
	ID      int     `json:"id"`
	Stars   float32 `json:"stars"`
	Comment string  `json:"comment"`
	BookID  int     `json:"bookID"`
}

type Vote struct {
	ID       int    `json:"id"`
	Vote     string `json:"vote"`
	ReviewID int    `json:"reviewID"`
	UserID   int    `json:"userID"`
}

type ThVote struct {
	result   int
	err      error
	voteType string
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

func (v Vote) validVote() bool {
	switch v.Vote {
	case "up":
		return true
	case "down":
		return true
	default:
		return false
	}
}

func CreateReviewsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists reviews (id serial primary key, user_id int references authentication(id), book_id int references books(id)"+
		" , stars numeric, comment text)")
	if err != nil {
		return err
	}

	return nil
}

func CreateVotesTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists votes (id serial primary key, vote text, review_id int references reviews(id), user_id int references authentication(id))")
	if err != nil {
		return err
	}

	return nil
}

func LeaveReview(c *gin.Context) {
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

	title := ""
	err = conn.QueryRow(context.Background(), "select title from books where id = $1", review.BookID).Scan(&title)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the title of the book from the database"})
		return
	}

	hasBorrowedThatBook := false
	err = conn.QueryRow(context.Background(), fmt.Sprintf(`select hasborrowed from
				(
				select a.id, a.history, ($1 = ANY (a.history)) as hasBorrowed
				from books b join authentication a on b.id = a.id
				)
				where id = $2`), title, id).Scan(&hasBorrowedThatBook)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "You haven't borrowed this book, so you can't leave a review"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to check if the user has borrowed this book"})
		return
	}

	if !hasBorrowedThatBook {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error you are unable to leave reviews on books that you haven't borrowed"})
		return
	}

	check := 0
	ok = false
	err = conn.QueryRow(context.Background(), "select id from reviews where user_id = $1 and book_id = $2", id, review.BookID).Scan(&check)
	if err != nil {
		if err == pgx.ErrNoRows {
			ok = true
		}

		if !ok {
			log.Println(err)
			c.JSON(http.StatusForbidden, gin.H{"error": "Error a user can leave a review only once!"})
			return
		}
	}

	if check != 0 {
		log.Println("Error a user can leave a review only once!")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error a user can leave a review only once!"})
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

func DeleteReview(c *gin.Context) {
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

	if err = CreateReviewsTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to create a table for reviews"})
		return
	}

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

func EditReview(c *gin.Context) {
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && bookID && (comment || stars)

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Token was not provided correctl")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error token was not provided correctl"})
		return
	}

	id, _, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	bookIDFl, ok := information["bookID"].(float64)
	if !ok {
		log.Println("Error incorrectly provided the id of the book")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided the id of the book"})
		return
	}

	review := Review{}
	review.BookID = int(bookIDFl)

	problem := false
	starsCheck, ok := information["stars"].(float64)
	if !ok {
		problem = true
	} else {
		review.Stars = float32(starsCheck)
		if !review.validNumberOfStarts() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error invalid number of stars"})
			return
		}
	}

	newComment := true
	review.Comment, ok = information["comment"].(string)
	if !ok && problem {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no new information given in order to edit the old one"})
		return
	} else if !ok {
		newComment = false
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	var stars float32
	comment := ""
	err = conn.QueryRow(context.Background(), "select comment, stars from reviews where user_id = $1 and book_id = $2", id, review.BookID).Scan(&comment, &stars)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"error": "Error you haven't left a review on this book, so you can't edit it"})
			return
		}

		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while checking if you have left a review"})
		return
	}

	if problem {
		review.Stars = stars
	} else if !newComment {
		review.Comment = comment
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the new comment or stars"})
		return
	}

	_, err = conn.Exec(context.Background(), "update reviews set comment = $1, stars = $2 where user_id = $3 and book_id = $4", review.Comment, review.Stars, id, review.BookID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable update the review"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetReviewsForBook(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // bookID || title

	problem := false
	bookID := 0
	bookIDFl, ok := information["bookID"].(float64)
	if !ok {
		problem = true
	} else {
		bookID = int(bookIDFl)
	}

	title, ok := information["title"].(string)
	if !ok && problem {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error not enough information provided in order to determine which book you want the reviews to"})
		return
	}

	var rows pgx.Rows
	if !problem {
		rows, err = conn.Query(context.Background(), "select id, stars, comment from reviews r where r.book_id = $1", bookID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the reviews from the database"})
			return
		}
	} else {
		err = conn.QueryRow(context.Background(), "select id from books b where b.title = $1", title).Scan(&bookID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information about the book from the database"})
			return
		}

		rows, err = conn.Query(context.Background(), "select id, stars, comment from reviews r where r.book_id = $1", bookID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the reviews from the database"})
			return
		}
	}

	var reviews []Review
	for rows.Next() {
		review := Review{}
		err = rows.Scan(&review.ID, &review.Stars, &review.Comment)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data"})
			return
		}

		review.BookID = bookID
		reviews = append(reviews, review)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to work with the information correctly"})
		return

	}

	if reviews == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No reviews have been made on this book yet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func GetReviewsOfUser(c *gin.Context) {
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && userID

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrectly provided token"})
		return
	}

	_, accountType, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error only librarians can view the comments of a certain user"})
		return
	}

	userIDFl, ok := information["userID"].(float64)
	if !ok {
		log.Println("Incorrectly provided information about the id of the user")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided information about the id of the user"})
		return
	}
	userID := int(userIDFl)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, stars, comment, book_id from reviews where user_id = $1", userID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information from the database"})
		return
	}

	var reviews []Review
	for rows.Next() {
		review := Review{}
		err = rows.Scan(&review.ID, &review.Stars, &review.Comment, &review.BookID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data"})
			return
		}

		reviews = append(reviews, review)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func GetBookRating(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	var information map[string]int
	json.NewDecoder(c.Request.Body).Decode(&information) //bookID

	bookID, ok := information["bookID"]
	if !ok {
		log.Println("Error the id of the book is not specified")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error the id of the book is not specified"})
		return
	}

	rows, err := conn.Query(context.Background(), "select stars from reviews where book_id = $1", bookID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information about the reviews on this book"})
		return
	}

	count := 0
	var sum float32
	for rows.Next() {
		var stars float32
		err = rows.Scan(&stars)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the rating of this book"})
			return
		}

		sum += stars
		count++
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the reviews"})
		return
	}

	rating := sum / float32(count)
	c.JSON(http.StatusOK, gin.H{"rating": rating})
}

func GetHighestRatedReviews(c *gin.Context) {
	var information map[string]int
	json.NewDecoder(c.Request.Body).Decode(&information)

	bookID, ok := information["bookID"]
	if !ok {
		log.Println("Incorrectly provided id of the book")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided id of the book"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, stars, comment from reviews r where r.stars >= 4 and r.book_id = $1", bookID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information from the database"})
		return
	}

	var reviews []Review
	for rows.Next() {
		review := Review{}
		err = rows.Scan(&review.ID, &review.Stars, &review.Comment)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the information"})
			return
		}

		review.BookID = bookID
		reviews = append(reviews, review)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the information from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func GetLowestRatedReviews(c *gin.Context) {
	var information map[string]int
	json.NewDecoder(c.Request.Body).Decode(&information)

	bookID, ok := information["bookID"]
	if !ok {
		log.Println("Incorrectly provided id of the book")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided id of the book"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, stars, comment from reviews r where r.stars <= 2 and r.book_id = $1", bookID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get information from the database"})
		return
	}

	var reviews []Review
	for rows.Next() {
		review := Review{}
		err = rows.Scan(&review.ID, &review.Stars, &review.Comment)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the information"})
			return
		}

		review.BookID = bookID
		reviews = append(reviews, review)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the information from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func VoteForReview(c *gin.Context) {
	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // token && vote && reviewID

	token, ok := information["token"].(string)
	if !ok {
		log.Println("Incorrectly provided token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided token"})
		return
	}

	id, _, err := ValidateJWT(token)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	vote := Vote{}
	vote.UserID = id

	vote.Vote, ok = information["vote"].(string)
	if !ok {
		log.Println("Incorrectly provided vote")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided vote"})
		return
	}

	if !vote.validVote() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error invalid vote"})
		return
	}

	reviewID, ok := information["reviewID"].(float64)
	if !ok {
		log.Println("Incorrectly provided the id of the review which you are voting about")
		c.JSON(http.StatusBadRequest, gin.H{"errror": "Error incorrectly provided the id which you are voting about"})
		return
	}
	vote.ReviewID = int(reviewID)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	if err = CreateVotesTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to create a table for the votes"})
		return
	}

	idCheck := 0
	err = conn.QueryRow(context.Background(), "select id from votes where review_id = $1 and user_id = $2", vote.ReviewID, id).Scan(&idCheck)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while trying to check if the user has already voted for this review"})
			return
		}
	}

	if idCheck != 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error you can't vote multiple times for the same review"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into votes (vote, user_id, review_id) values ($1, $2, $3)", vote.Vote, vote.UserID, vote.ReviewID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to put the information about your vote in the database"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func getVotes(conn *pgx.Conn, result chan<- ThVote, reviewID int, mu *sync.Mutex, voteType string) {
	res := ThVote{}
	res.voteType = voteType
	count := 0

	mu.Lock()
	err := conn.QueryRow(context.Background(), "select count(*) from votes v where v.vote = $1 and v.review_id = $2", res.voteType, reviewID).Scan(&count)
	mu.Unlock()

	if err != nil {
		res.result = 0
		res.err = err

		result <- res
		return
	}

	res.result = count
	res.err = nil
	result <- res
}

func GetVotesForReview(c *gin.Context) {
	result := make(chan ThVote)

	var info map[string]int
	json.NewDecoder(c.Request.Body).Decode(&info)

	reviewID, ok := info["reviewID"]
	if !ok {
		log.Println("Incorrectly provided the id of the review")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided the id of the review"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	mu := &sync.Mutex{}
	go getVotes(conn, result, reviewID, mu, "up")
	go getVotes(conn, result, reviewID, mu, "down")

	information := make(map[string]int)
	for range 2 {
		select {
		case r := <-result:
			if r.err != nil {
				log.Println(r.err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the votes for this review"})
				return
			}

			information[r.voteType] = r.result
		}
	}

	c.JSON(http.StatusOK, information)
}

type thReview struct {
	count int
	err   error
}

func getReviews(conn *pgx.Conn, result chan<- thReview, stars float32, mu *sync.Mutex, bookID int) {
	res := thReview{}

	mu.Lock()
	err := conn.QueryRow(context.Background(), "select count(*) from reviews r where r.stars = $1 and r.book_id = $2", stars, bookID).Scan(&res.count)
	mu.Unlock()

	if err != nil {
		res.err = err
		result <- res
		return
	}

	res.err = nil
	result <- res
}

func RatingDetails(c *gin.Context) {
	result := make(chan thReview)
	mu := &sync.Mutex{}

	var information map[string]int
	json.NewDecoder(c.Request.Body).Decode(&information)

	bookID, ok := information["bookID"]
	if !ok {
		log.Println("Incorrectly provided bookID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error incorrectly provided bookID"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	for i := 0.0; i <= 5.0; i += 0.5 {
		go getReviews(conn, result, float32(i), mu, bookID)
	}

	info := make(map[string]int)
	for i := 0.0; i <= 5.0; i += 0.5 {
		select {
		case res := <-result:
			if res.err != nil {
				log.Println(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error collecting the reviews"})
				return
			}

			info[fmt.Sprintf("%v", i)] = res.count
		}
	}

	c.JSON(http.StatusOK, info)
}
