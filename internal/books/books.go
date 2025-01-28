package books

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Quantity int    `json:"quantity"`
}

func CreateBookTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists books (id serial primary key, isbn text, title text, author text, year int, "+
		"reserved_from_id int, quantity int, foreign key (reserved_from_id) references authentication(id));")
	if err != nil {
		log.Println(err)
		return errors.New("Couldn't create a table")
	}

	return nil
}
func getBooks(conn *pgx.Conn) ([]Book, error) {
	var bookList []Book
	rows, err := conn.Query(context.Background(), "select id, isbn, title, author, year, quantity from books;")
	if err != nil {
		log.Println(err)
		return nil, errors.New("Failed to fetch books")
	}

	for rows.Next() {
		var book Book
		err = rows.Scan(&book.ID, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Quantity)
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

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	bookList, err := getBooks(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if bookList == nil {
		log.Println("There are no books created")
		c.JSON(http.StatusNotFound, gin.H{"error": "There are no books created"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"books": bookList})
}

func AddBook(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		log.Println("Users can't add books")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error users can't add books"})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //isbn, title, author, year, quantity

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	_, err = conn.Exec(context.Background(), "create table if not exists books (id serial primary key, isbn text, title text, author text, year int, "+
		"reserved_from_id int, quantity int, foreign key (reserved_from_id) references authentication(id));")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't create a table"})
		return
	}

	fmt.Println(book)
	_, err = conn.Exec(context.Background(), "insert into books (isbn, title, author, year, quantity) values ($1, $2, $3, $4, $5);",
		book.ISBN, book.Title, book.Author, book.Year, book.Quantity)
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

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //name

	bookList, err := getBooks(conn)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var matchedNames []string
	for _, book := range bookList {
		foundMatch, err := regexp.MatchString(".*"+information["name"]+".*", book.Title)
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
		log.Println("Couldn't find books with the name provided")
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

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	err = conn.QueryRow(context.Background(), "update books b set b.quantity = b.quantity - 1 where b.title = $1;", book.Title).Scan(
		&book.ID, &book.Title, &book.Author, &book.Year, &book.ISBN, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book"})
		return
	}

	if book.Quantity == 0 {
		log.Println("All of the copies of this book have already been borrowed. Please chose another one.")
		c.JSON(http.StatusForbidden, gin.H{"error": "Trying to borrow an already borrowed book"})
		return
	}

	var history string
	err = conn.QueryRow(context.Background(), "select history from authentication a where a.id = $1;", CurrentPrfile.ID).Scan(&history)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the database"})
		return
	}
	editHistory := []byte(history)
	if string(editHistory[0]) == " " {
		history = book.Title
	} else {
		history = history + ", " + book.Title
	}

	_, err = conn.Exec(context.Background(), "update authentication a set history = $1 where a.id = $2;", history, CurrentPrfile.ID)

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

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	err = conn.QueryRow(context.Background(), "update books b set b.quantity = b.quantity + 1 where b.title = $1;", book.Title).Scan(
		&book.ID, &book.Title, &book.Author, &book.Year, &book.ISBN, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error returning the book"})
	}

	var reserved_id int
	err = conn.QueryRow(context.Background(), "select reserved_from_id from books b where b.title = $1;", book.Title).Scan(&reserved_id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if teh book is reserved"})
		return
	}

	if reserved_id == 0 {
		_, err = conn.Exec(context.Background(), "update books b set b.quantity = b.quantity + 1 where id = $1;", book.ID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the database"})
			return
		}
	} else {
		_, err = conn.Exec(context.Background(), "update authentication a set a.history = a.history || ', ' || $1 where a.id = $2;", book.Title, reserved_id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the history of this person"})
			return
		}
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

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	_, err = conn.Exec(context.Background(), "update books set reserved_from_id = $1 from books b where b.title = $2;", CurrentPrfile.ID, book.Title)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reserving the book"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func UpdateBookQuantity(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		log.Println("Users can't update the quantity of books")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error users can't update the quantity of books"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) // id && quantity
	_, err = conn.Exec(context.Background(), "update books set quantity = $1 where id = $2;", book.Quantity, book.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the quantity of books"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func RemoveBook(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		log.Println("Users can't remove books")
		c.JSON(http.StatusForbidden, gin.H{"error": "Error users can't remove books"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) // id && (title || description)
	_, err = conn.Exec(context.Background(), "delete from books b where b.id = $1", book.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing the book"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
