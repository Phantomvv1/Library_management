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
	"time"

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

func borrowBook(conn *pgx.Conn, book Book, returnDate time.Time) error {
	_, err := conn.Exec(context.Background(), "insert into borrowed_books (book_id, user_id, return_date) values ($1, $2, $3)", book.ID, CurrentPrfile.ID, returnDate)
	if err != nil {
		log.Println(err)
		return errors.New("Unable to put the information about the borrowed book in the table")
	}

	return nil
}

func createBorrowedBooksTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists borrowed_books (id serial primary key not null, book_id int, user_id int, return_date date);")
	if err != nil {
		log.Println(err)
		return errors.New("Unable to create a table for keeping the borrowed books in.")
	}

	return nil
}

func updateHistory(conn *pgx.Conn, book Book, userID int) error {
	var history []string
	err := conn.QueryRow(context.Background(), "select history from authentication a where a.id = $1 limit 1;", userID).Scan(&history)
	if err != nil {
		log.Println(err)
		return errors.New("Couldn't get the history of the user")
	}

	for _, title := range history {
		if title == book.Title {
			return nil
		}
	}

	CurrentPrfile.History = append(CurrentPrfile.History, book.Title)

	_, err = conn.Exec(context.Background(), "update authentication set history = array_append(history, $1) where id = $2;", book.Title, userID)
	if err != nil {
		log.Println(err)
		return errors.New("Error updating the history of this person")
	}

	return nil
}

func borrowReservedBooks(conn *pgx.Conn, book Book) error {
	var userID int
	err := conn.QueryRow(context.Background(), "select book_id, user_id from book_reservations b where b.book_id = $1 order by b.id asc limit 1;", book.ID).Scan(&book.ID, &userID)
	if err != nil {
		if err == pgx.ErrNoRows { //no reservations for this book
			return nil
		}

		log.Println(err)
		return errors.New("Error checking if the book is reserved")
	}

	_, err = conn.Exec(context.Background(), "update books set quantity = quantity - 1 where id = $1;", book.ID)
	if err != nil {
		log.Println(err)
		return errors.New("Error updating the database")
	}

	err = updateHistory(conn, book, userID)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = conn.Exec(context.Background(), "delete from book_reservations where book_id = $1 and user_id = $2", book.ID, userID)
	if err != nil {
		log.Println(err)
		return errors.New("Error removing the borrowing the reserved book")
	}

	return nil

}

func CreateBookReservationsTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists book_reservations (id serial primary key, book_id int, user_id int);")
	if err != nil {
		log.Println(err)
		return errors.New("Couldn't create a table for story the books reserved from customers.")
	}

	return nil
}

func CreateBookTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists books (id serial primary key, isbn text, title text, author text, year int, "+
		"quantity int);")
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

	err = CreateBookTable(conn)
	if err != nil {
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

	if err = createBorrowedBooksTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var returnDateMap map[string]string
	json.NewDecoder(c.Request.Body).Decode(&returnDateMap) //title && returnDate (author | isbn | year | id)

	var book Book
	book.Title = returnDateMap["title"]

	_, err = conn.Exec(context.Background(), "update books set quantity = quantity - 1 where title = $1 and quantity > 0;", book.Title)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book"})
		return
	}

	err = conn.QueryRow(context.Background(), "select id, quantity from books b where b.title = $1", book.Title).Scan(&book.ID, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error viewing the quantity of the book"})
		return
	}

	if book.Quantity == 0 {
		log.Println("All of the copies of this book have already been borrowed. Please chose another one.")
		c.JSON(http.StatusForbidden, gin.H{"error": "All of the copies of this book have already been borrowed. Please chose another one."})
		return
	}

	if CurrentPrfile.ID == 0 {
		log.Println("Not logged in")
		c.JSON(http.StatusForbidden, gin.H{"error": "You need to log in, in order to borrow a book"})
		return
	}

	returnDate, err := time.Parse(time.DateOnly, returnDateMap["returnDate"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the return date."})
		return
	}

	if err = borrowBook(conn, book, returnDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err = updateHistory(conn, book, CurrentPrfile.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the history of the person"})
		return
	}

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

	if err = CreateBookReservationsTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err = createBorrowedBooksTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title

	err = conn.QueryRow(context.Background(), "select id, isbn, title, author, year, quantity from books b where b.title = $1;", book.Title).Scan(
		&book.ID, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book details"})
		return
	}

	if CurrentPrfile.ID == 0 {
		log.Println("The user hasn't logged in")
		c.JSON(http.StatusForbidden, gin.H{"error": "You haven't logged in. You must be logged in, in order to return a book."})
		return
	}

	_, err = conn.Exec(context.Background(), "delete from borrowed_books where book_id = $1 and user_id = $2;", book.ID, CurrentPrfile.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error returning the book"})
		return
	}

	_, err = conn.Exec(context.Background(), "update books set quantity = quantity + 1 where id = $1;", book.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding the book to our inventory"})
		return
	}

	if err = borrowReservedBooks(conn, book); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
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

	if err = CreateBookReservationsTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var book Book
	json.NewDecoder(c.Request.Body).Decode(&book) //title & (author | isbn | year | id)

	if CurrentPrfile.ID == 0 {
		log.Println("Not logged in")
		c.JSON(http.StatusForbidden, gin.H{"error": "You need to log in, in order to reserve a book"})
		return
	}

	err = conn.QueryRow(context.Background(), "select id, isbn, title, author, year, quantity from books b where b.title = $1;", book.Title).Scan(
		&book.ID, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book details"})
		return
	}

	if book.Quantity > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "There are copies from this book available. There is no need to reserve it."})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into book_reservations (book_id, user_id) values ($1, $2);", book.ID, CurrentPrfile.ID)
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

	var rowsCount int
	err = conn.QueryRow(context.Background(), "select count(*) from book_reservations where book_id = $1", book.ID).Scan(&rowsCount)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting the reservations for this book."})
		return
	}

	if book.Quantity > rowsCount {
		for range rowsCount {
			if err = borrowReservedBooks(conn, book); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

		}
	} else {
		for range book.Quantity {
			if err = borrowReservedBooks(conn, book); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		}
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
