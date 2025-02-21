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
	"strconv"
	"time"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	. "github.com/Phantomvv1/Library_management/internal/users"
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

func borrowBook(conn *pgx.Conn, userID int, book Book, returnDate time.Time) error {
	_, err := conn.Exec(context.Background(), "insert into borrowed_books (book_id, user_id, return_date) values ($1, $2, $3)", book.ID, userID, returnDate)
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
		return errors.New("Couldn't create a table for books")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bookList, err := getBooks(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bookList == nil {
		log.Println("There are no books created")
		c.JSON(http.StatusNotFound, gin.H{"error": "There are no books created"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"books": bookList})
}

func AddBook(c *gin.Context) { //to be tested
	var information map[string]string
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&information) //isbn, title, author, year, quantity

	_, accountType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error users can't add books"})
		return
	}

	book.ISBN = information["isbn"]
	book.Title = information["title"]
	book.Author = information["author"]
	quantity, err := strconv.Atoi(information["quantity"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the quantity of the book"})
		return
	}
	book.Quantity = quantity

	year, err := strconv.Atoi(information["quantity"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the quantity of the book"})
		return
	}
	book.Year = uint(year) //To make year a timestamp (time.Time)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //name

	bookList, err := getBooks(conn)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func BorrowBook(c *gin.Context) { //to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = createBorrowedBooksTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //title && returnDate (author | isbn | year | id)

	id, _, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var book Book
	book.Title = information["title"]

	_, err = conn.Exec(context.Background(), "update books set quantity = quantity - 1 where title = $1 and quantity > 0;", book.Title)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book"})
		return
	}

	err = conn.QueryRow(context.Background(), "select id, quantity from books b where b.title = $1;", book.Title).Scan(&book.ID, &book.Quantity)
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

	returnDate, err := time.Parse(time.RFC3339, information["returnDate"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the return date."})
		return
	}

	if err = borrowBook(conn, id, book, returnDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = updateHistory(conn, book, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating the history of the person"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func ReturnBook(c *gin.Context) { //to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = CreateBookReservationsTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = createBorrowedBooksTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&information) //title

	id, _, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	book.Title = information["title"]

	err = conn.QueryRow(context.Background(), "select id, isbn, title, author, year, quantity from books b where b.title = $1;", book.Title).Scan(
		&book.ID, &book.ISBN, &book.Title, &book.Author, &book.Year, &book.Quantity)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the book details"})
		return
	}

	var check int
	err = conn.QueryRow(context.Background(), "delete from borrowed_books where book_id = $1 and user_id = $2 returning id;", book.ID, id).Scan(&check)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{"message": "You can't return a book that you haven't borrowed or you have already returned"})
			return
		}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetHistory(c *gin.Context) { //to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information)

	id, _, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var history []string
	err = conn.QueryRow(context.Background(), "select history books b where b.id = $1;", id).Scan(&history)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the history from the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func ReserveBook(c *gin.Context) { // to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = CreateBookReservationsTable(conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&information) //title & (author | isbn | year | id)

	id, _, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	book.Title = information["title"]

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

	_, err = conn.Exec(context.Background(), "insert into book_reservations (book_id, user_id) values ($1, $2);", book.ID, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reserving the book"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func UpdateBookQuantity(c *gin.Context) { // to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&information) // id && quantity && token

	_, accountType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Error users can't update the quantity of books"})
		return
	}

	book.ID, err = strconv.Atoi(information["id"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the id of the book"})
	}

	book.Quantity, err = strconv.Atoi(information["quantity"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the id of the book"})
		return
	}

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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		}
	} else {
		for range book.Quantity {
			if err = borrowReservedBooks(conn, book); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, nil)
}

func RemoveBook(c *gin.Context) { // to be tested
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = CreateBookTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var information map[string]string
	var book Book
	json.NewDecoder(c.Request.Body).Decode(&information) // id && (title || description)

	_, accountType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error users can't remove books"})
		return
	}

	book.ID, err = strconv.Atoi(information["id"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing the id of the book"})
		return
	}

	_, err = conn.Exec(context.Background(), "delete from books b where b.id = $1", book.ID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing the book"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetBooksOverdue(c *gin.Context) { // to be tested
	information := make(map[string]string)
	json.NewDecoder(c.Request.Body).Decode(&information)

	_, accoutType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if accoutType != "librarian" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User's can't view all of the books that are overdue"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error couldn't connect to the databse"})
		return
	}
	defer conn.Close(context.Background())

	if err = CreateBookTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = CreateAuthTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = createBorrowedBooksTable(conn); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := conn.Query(context.Background(), "select book_id, user_id from borrowed_books bb where current_timestamp > bb.return_date")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the users from the database"})
		return
	}

	var userIDs, bookIDs []int
	for rows.Next() {
		var bookID, userID int
		err = rows.Scan(&bookID, &userID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the data"})
			return
		}

		bookIDs = append(bookIDs, bookID)
		userIDs = append(userIDs, userID)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while working with the data"})
		return
	}

	users := []User{}
	query := "select name, email from authentication where id = $1"
	for i := range userIDs {
		if i == 0 {
			continue
		}

		query = query + " or id = $" + string(byte(i+'1'))
	}
	query = query + ";"
	fmt.Println(query) //for debugging purposes

	rows, err = conn.Query(context.Background(), query, userIDs)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users from the database"})
		return
	}

	i := 0
	for rows.Next() {
		var name, email string
		err = rows.Scan(&name, &email)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the users from the database"})
			return
		}

		users = append(users, User{ID: userIDs[i], Name: name, Email: email})
		i++
	}

	books := []Book{}
	query = "select isbn, title, author, quantity, year from books where id = $1"
	for i := range userIDs {
		if i == 0 {
			continue
		}

		query = query + " or id = $" + string(byte(i+'1'))
	}
	query = query + ";"
	fmt.Println(query) //for debugging purposes

	rows, err = conn.Query(context.Background(), query, bookIDs)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users from the database"})
		return
	}

	i = 0
	for rows.Next() {
		var isbn, title, author string
		var quantity int
		var year uint

		err = rows.Scan(&isbn, &title, &author, &quantity, &year)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the books from the database"})
			return
		}

		books = append(books, Book{ID: bookIDs[i], ISBN: isbn, Title: title, Author: author, Quantity: quantity, Year: year})
		i++
	}

	result := make(map[User]Book)
	for i, user := range users {
		result[user] = books[i]
	}

	c.JSON(http.StatusOK, result)
}
