package books

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Book struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   uint   `json:"year"`
}

func GetBooks(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
	}
	defer conn.Close(context.Background())

	var bookList []Book
	rows, err := conn.Query(context.Background(), "select name, isbn, title, author, year from books;")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
	}

	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ID, &book.Name, &book.ISBN, &book.Title, &book.Author, &book.Year)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process books"})
		}

		bookList = append(bookList, book)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch books"})
	}

	c.JSON(http.StatusOK, gin.H{"books": bookList})
}

func AddBook(c *gin.Context) {
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book)

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "create table if not exists books (id primary key serial, name text, isbn text, title text, author text, year text);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create a table"})
	}

	_, err = conn.Exec(context.Background(), "insert into books (name, isbn, title, author, year) values ($1, $2, $3, $4, $5);",
		book.Name, book.ISBN, book.Title, book.Author, book.Year)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't insert into the table"})
	}

	c.JSON(http.StatusOK, nil)
}
