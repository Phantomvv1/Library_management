package books

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Book struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ISBN     string `json:"isbn"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Year     uint   `json:"year"`
	Borrowed bool   `json:"borrowed"`
}

func getBooks(conn *pgx.Conn) ([]Book, error) {
	var bookList []Book
	rows, err := conn.Query(context.Background(), "select name, isbn, title, author, year, borrowed from books;")
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to fetch books")
	}

	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ID, &book.Name, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Borrowed)
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

	_, err = conn.Exec(context.Background(), "create table if not exists books (id primary key serial, name text, isbn text, title text, author text, year text, borrowed boolean);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create a table"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into books (name, isbn, title, author, year, borrowed) values ($1, $2, $3, $4, $5, false);",
		book.Name, book.ISBN, book.Title, book.Author, book.Year)
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
		foundMatch, err := regexp.MatchString("%"+information["name"]+"%", book.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if foundMatch {
			matchedNames = append(matchedNames, book.Name)
		}

	}

	if matchedNames == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Couldn't find books with the name provided"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"books": matchedNames})
}
