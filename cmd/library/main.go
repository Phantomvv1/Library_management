package main

import (
	"net/http"

	. "github.com/Phantomvv1/Library_management/internal/authentication"
	. "github.com/Phantomvv1/Library_management/internal/books"
	. "github.com/Phantomvv1/Library_management/internal/librarians"
	. "github.com/Phantomvv1/Library_management/internal/reviews"
	. "github.com/Phantomvv1/Library_management/internal/users"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Any("/", func(c *gin.Context) { c.JSON(http.StatusOK, nil) })
	r.GET("/users", GetUsers)
	r.GET("/books", GetBooks)
	r.GET("/librarians", GetLibrarians)
	r.GET("/events", GetEvents)
	r.GET("/event/upcoming", GetUpcomingEvents)
	r.GET("/book", GetBookByID)
	r.GET("/authors", GetAuthors)
	r.GET("/book/availability", IsAvailable)
	r.POST("/review", LeaveReview)
	r.DELETE("/review", DeleteReview)
	r.PUT("/review", EditReview)
	r.POST("/book/review", GetReviewsForBook)
	r.POST("/review/user", GetReviewsOfUser)
	r.POST("/book/rating", GetBookRating)
	r.POST("/book/review/high", GetHighestRatedReviews)
	r.POST("/book/review/low", GetLowestRatedReviews)
	r.POST("/review/vote", VoteForReview)
	r.POST("/review/votes", GetVotesForReview)
	r.POST("/review/votes/sql", GetVotesForReviewSQL)
	r.POST("/book/rating/details", RatingDetails)
	r.POST("/book/rating/details/sql", RatingDetailsSQL)
	r.POST("/book/cancel/reservation", CancelBookReservation)
	r.POST("/user", GetUserByID)
	r.POST("/user/history", GetUserHistory)
	r.POST("/history", GetHistory)
	r.POST("/profile", GetCurrentProfile)
	r.POST("/event/invited", GetInvited)
	r.POST("/book", AddBook)
	r.POST("/signup", SignUp)
	r.POST("/login", LogIn)
	r.POST("/searchbook", SearchForBook)
	r.POST("/edit", EditProfile)
	r.POST("/book/borrow", BorrowBook)
	r.POST("/book/return", ReturnBook)
	r.POST("/book/reserve", ReserveBook)
	r.POST("/event", CreateEvent)
	r.POST("/event/invite", InviteToEvent)
	r.POST("/book/quantity", UpdateBookQuantity)
	r.POST("/book/remove", RemoveBook)
	r.POST("/book/overdue", GetBooksOverdue)

	r.Run(":42069")
}
