package books

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Book struct {
	ID       int    `json:"id"`
	ISBN     string `json:"isbn"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Year     uint   `json:"year"`
	Borrowed bool   `json:"borrowed"`
}

func getBooks(conn *pgx.Conn) ([]Book, error) {
	var bookList []Book
	rows, err := conn.Query(context.Background(), "select isbn, title, author, year, borrowed from books;")
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to fetch books")
	}

	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ID, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Borrowed)
		if err != nil {
			log.Println(err)
			return nil, errors.New("Failed to process books")
		}

		bookList = append(bookList, book)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		return nil, errors.New("Failed to fetch books")
	}

	return bookList, nil
}

func GetBooks(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	bookList, err := getBooks(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"books": bookList})
}

func AddBook(c *gin.Context) {
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //name, isbn, title, author, year

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "create table if not exists books (id primary key serial, isbn text, title text, author text, year text, "+
		"borrowed boolean, reserved_from_id foreign key int);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create a table"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into books (name, isbn, title, author, year, borrowed, reserved_from_id) values ($1, $2, $3, $4, false, 0);",
		book.ISBN, book.Title, book.Author, book.Year)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't insert into the table"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func SearchForBook(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
	}
	defer conn.Close(context.Background())

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //name

	bookList, err := getBooks(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var matchedNames []string
	for _, book := range bookList {
		foundMatch, err := regexp.MatchString("%"+information["name"]+"%", book.Title)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if foundMatch {
			matchedNames = append(matchedNames, book.Title)
		}

	}

	if matchedNames == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find books with the name provided"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"books": matchedNames})
}

func BorrowBook(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	err = conn.QueryRow(context.Background(), "select id, title, author, year, isbn, borrowed from books b where b.title = $1", book.Title).Scan(
		&book.ID, &book.Title, &book.Author, &book.Year, &book.ISBN, &book.Borrowed)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book"})
		return
	}

	if book.Borrowed {
		log.Println("This book has already been borrowed. Please chose another one.")
		c.JSON(http.StatusForbidden, gin.H{"error": "Trying to borrow an already borrowed book"})
		return
	}

	_, err = conn.Exec(context.Background(), "update books set borrowed = true, where id = $1", book.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the database"})
		return
	}

	var history string
	err = conn.QueryRow(context.Background(), "select history from authentication a where a.id = $1", CurrentPrfile.ID).Scan(&history)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the database"})
		return
	}
	history = history + " " + book.Title

	_, err = conn.Exec(context.Background(), "update authentication a set history = $1 where a.id = $2", history, CurrentPrfile.ID)

	c.JSON(http.StatusOK, nil)
}

func ReturnBook(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	err = conn.QueryRow(context.Background(), "select id, title, author, year, isbn, borrowed from books b where b.title = $1", book.Title).Scan(
		&book.ID, &book.Title, &book.Author, &book.Year, &book.ISBN, &book.Borrowed)

	if !book.Borrowed {
		log.Println("This book is not borrowed. Please chose another one.")
		c.JSON(http.StatusForbidden, gin.H{"error": "Trying to return a not borrowed book"})
		return
	}

	var reserved_id int
	err = conn.QueryRow(context.Background(), "select reserved_from_id from books b where b.title = $1", book.Title).Scan(&reserved_id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if teh book is reserved"})
		return
	}

	if reserved_id == 0 {
		_, err = conn.Exec(context.Background(), "update books set borrowed = false where id = $1", book.ID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the database"})
			return
		}
	} else {
		_, err = conn.Exec(context.Background(), "update authentication a set a.history = concat(a.history, ' ', $1) where a.id = $2", book.Title, reserved_id)
	}

	c.JSON(http.StatusOK, nil)
}

func GetHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"history": CurrentPrfile.History})
}

func ReserveBook(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	err = conn.QueryRow(context.Background(), "update books set reserved_from_id = $1 from books b where b.title = $2", CurrentPrfile.ID, book.Title).Scan(
		&book.ID, &book.Title, &book.Author, &book.Year, &book.ISBN, &book.Borrowed)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reserving the book"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
