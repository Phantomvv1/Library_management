package librarians

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	. "github.com/Phantomvv1/Library_management/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Librarian struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LibrariansCreate struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Event struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Start       string `json:"start"` // Example: 1999-01-08 04:05:06
}

func GetLibrarians(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	_, err = conn.Exec(context.Background(), "create table if not exists authentication (id serial primary key not null, name text, email text, password text, type text, history text);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a table for authentication"})
		return
	}

	var librarianList []Librarian
	rows, err := conn.Query(context.Background(), "select id, email, name from authentication where type = 'librarian';")
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNoContent, gin.H{"error": "There are no librarians"})
			return
		} else {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't read from the database"})
			return
		}
	}

	for rows.Next() {
		var user Librarian
		err = rows.Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process librarians"})
			return
		}
		librarianList = append(librarianList, user)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch librarians"})
		return
	}

	if librarianList == nil {
		log.Println("There are no librarians created.")
		c.JSON(http.StatusNotFound, gin.H{"error": "There are no librarians"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"librarians": librarianList})
}

func CreateEvent(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can create events"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), "create table if not exists events (id primary key serial not null, name text, description text, invited text, start timestamp);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing a table for the events"})
		return
	}

	var event Event
	json.NewDecoder(c.Request.Body).Decode(&event) //name && start (descrpition not neccessary)
	if event.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no name provided"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into events (name, description, invited, start) ($1, $2, NULL, $3);", event.Name, event.Description, event.Start)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating the event"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func InviteToEvent(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can create events"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	_, err = conn.Exec(context.Background(), "create table if not exists events (id primary key serial not null, name text, description text, invited text, start timestamp);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing a table for the events"})
		return
	}

	var user User
	json.NewDecoder(c.Request.Body).Decode(&user) // email && (id || name)

	err = conn.QueryRow(context.Background(), "select id, email, name from authentication where email = $1;", user.Email).Scan(user.ID, user.Email, user.Name)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting information from the database"})
		return
	}

	var invited string
	err = conn.QueryRow(context.Background(), "select invited from events;").Scan(&invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inviting the person"})
		return
	}
	invited = invited + ", " + user.Email

	_, err = conn.Exec(context.Background(), "update events set invited = $1;", invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inviting the person"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetInvited(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can view who is invited to an event"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	_, err = conn.Exec(context.Background(), "create table if not exists events (id primary key serial not null, name text, description text, invited text, start timestamp);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing a table for the events"})
		return
	}

	var event Event
	json.NewDecoder(c.Request.Body).Decode(&event) // name && (description || start)
	var invited string
	err = conn.QueryRow(context.Background(), "select invited from events where name = $1;", event.Name).Scan(&invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error there is no event with this name"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"people invited": invited})
}

func GetEvents(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	_, err = conn.Exec(context.Background(), "create table if not exists events (id serial primary key not null, name text, description text, invited text, start timestamp);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a table for the events"})
		return
	}

	rows, err := conn.Query(context.Background(), "select id, name, description, start from events")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the events from the database"})
		return
	}

	var events []Event
	for rows.Next() {
		var event Event
		err = rows.Scan(&event.ID, &event.Name, &event.Description, &event.Start)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error collecting all the events"})
			return
		}

		events = append(events, event)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": rows.Err()})
		return
	}

	if events == nil {
		log.Println("There are no events created")
		c.JSON(http.StatusNotFound, gin.H{"error": "There are no events created"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func ViewUserHistory(c *gin.Context) {
	if CurrentPrfile.Type != "librarian" {
		log.Println("Only librarians can view the history of users")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can view the history of users"})
		return
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select name, email, history from authentication a where a.type = 'user';")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the users' history"})
		return
	}

	var profiles []Profile
	for rows.Next() {
		var profile Profile
		err = rows.Scan(&profile.Name, &profile.Email, &profile.History)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the users' history"})
			return
		}

		profiles = append(profiles, profile)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the users' history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user history": profiles})
}
