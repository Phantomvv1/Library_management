package books

import (
	"context"
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
