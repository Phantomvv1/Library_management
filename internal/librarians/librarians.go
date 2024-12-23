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
}

func GetLibrarians(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}
	defer conn.Close(context.Background())

	var userList []Librarian
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
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.Println(rows.Err())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch librarians"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": userList})
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

	_, err = conn.Exec(context.Background(), "create table if not exists events (id primary key serial not null, name text, description text, invited text);")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error createing a table for the events"})
		return
	}

	var event Event
	json.NewDecoder(c.Request.Body).Decode(&event) //name (descrpition not neccessary)
	if event.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no name provided"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into events (name, description, invited) ($1, $2, NULL);", event.Name, event.Description)
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
	invited = invited + " " + user.Email

	_, err = conn.Exec(context.Background(), "update events set invited = $1;", invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inviting the person"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
