package librarians

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"` // Example: 1999-01-08T04:05:06Z
}

func CreateEventTable(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), "create table if not exists events (id serial primary key not null, name text, description text, invited text, start timestamp);")
	if err != nil {
		log.Println(err)
		return errors.New("Error creating a table for the events")
	}

	return nil
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
	err = CreateEventTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
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
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateEventTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]string
	var event Event
	json.NewDecoder(c.Request.Body).Decode(&information) //name && start && token (descrpition not neccessary)

	_, accountType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can create events"})
		return
	}

	if information["name"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error no name provided"})
		return
	}
	event.Name = information["name"]
	event.Start, err = time.Parse(time.RFC3339, information["start"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error trying to parse the time given"})
		return
	}

	_, err = conn.Exec(context.Background(), "insert into events (name, description, invited, start) values ($1, $2, ' ', $3);", event.Name, event.Description, event.Start)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating the event"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func InviteToEvent(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	err = CreateEventTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]interface{}
	json.NewDecoder(c.Request.Body).Decode(&information) // email && token && eventId

	tokenString, ok := information["token"].(string)
	if !ok {
		log.Println("Token is not of the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error token is not of the correct type"})
		return
	}
	_, accountType, err := ValidateJWT(tokenString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can invite people to events"})
		return
	}

	eventId, ok := information["eventId"].(float64)
	if !ok {
		log.Println("eventId is not of the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error eventId is not of the correct type"})
		return
	}

	email, ok := information["email"].(string)
	if !ok {
		log.Println("Email is not of the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error email is not of the correct type"})
		return
	}

	var user User
	err = conn.QueryRow(context.Background(), "select id, email, name from authentication where email = $1;", email).Scan(&user.ID, &user.Email, &user.Name)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting information from the database"})
		return
	}

	var invited string
	err = conn.QueryRow(context.Background(), "select invited from events e where e.id = $1;", eventId).Scan(&invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inviting the person"})
		return
	}

	//Check if the person has already been invited to the event
	invitedArr := strings.Split(invited, ", ")
	for _, email := range invitedArr {
		if email == user.Email {
			log.Println("The person has already been invited to this event")
			c.JSON(http.StatusConflict, gin.H{"error": "This person has already been invited to this event"})
			return
		}
	}

	if invited == " " {
		invited = user.Email
	} else {
		invited = invited + ", " + user.Email
	}

	_, err = conn.Exec(context.Background(), "update events set invited = $1;", invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inviting the person"})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func GetInvited(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	// NOTE: Creating the table if it doesn't exist
	err = CreateEventTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]interface{}
	var event Event
	json.NewDecoder(c.Request.Body).Decode(&information) // name && id && token

	tokenString, ok := information["token"].(string)
	if !ok {
		log.Println("Token is not of the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error token is not of the correct type"})
		return
	}
	_, accountType, err := ValidateJWT(tokenString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	if accountType != "librarian" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only librarians can view who is invited to an event"})
		return
	}

	id, ok := information["id"].(float64)
	if !ok {
		log.Println("Name is not of the correct type")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error name is not of the correct type"})
		return
	}
	event.ID = int(id)

	var invited string
	err = conn.QueryRow(context.Background(), "select invited from events e where e.id = $1", event.ID).Scan(&invited)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error there is no event with this name"})
		return
	}

	if invited == " " {
		log.Println("No people have been invited to this event")
		c.JSON(http.StatusNotFound, gin.H{"error": "No people have been invited to this event"})
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
	err = CreateEventTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
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

func GetUserHistory(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	err = CreateAuthTable(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var information map[string]string
	json.NewDecoder(c.Request.Body).Decode(&information) //token

	_, accoutnType, err := ValidateJWT(information["token"])
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Error invalid token"})
		return
	}

	if accoutnType != "librarian" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only librarians can view the history of users"})
		return
	}

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

func GetUpcomingEvents(c *gin.Context) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to connect to the database"})
		return
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), "select id, name, description, start from events e where e.start > current_timestamp")
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error unable to get the information from the database"})
		return
	}

	var events []Event
	for rows.Next() {
		event := Event{}
		err = rows.Scan(&event.ID, &event.Name, &event.Description, &event.Start)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error working with the information from the database"})
			return
		}

		events = append(events, event)
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
