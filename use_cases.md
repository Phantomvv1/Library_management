Users
UC1: Retrieve All Users
    Actor: Admin
    Precondition: Admin is authenticated
    Main Flow: Admin sends a GET request to /users
    Postcondition: List of all registered users is returned

UC2: User Sign Up
    Actor: Guest
    Precondition: Guest is not registered
    Main Flow: Sends a POST request to /signup with required user data
    Postcondition: User account is created

UC3: User Login
    Actor: Registered User
    Precondition: Valid account exists
    Main Flow: Sends a POST request to /login with credentials
    Postcondition: User is authenticated

UC4: View Profile
    Actor: Logged-in User
    Precondition: User is authenticated
    Main Flow: Sends a POST request to /profile
    Postcondition: Current profile data is returned

UC5: Edit Profile
    Actor: Logged-in User
    Precondition: User is authenticated
    Main Flow: Sends a POST request to /edit with updated data
    Postcondition: Profile is updated

UC6: Delete Account
    Actor: Logged-in User or Librarian
    Precondition: User is authenticated
    Main Flow: Sends a DELETE request to /user
    Postcondition: Account is deleted

UC7: View User Borrowing History
    Actor: Logged-in User
    Precondition: User is authenticated
    Main Flow: Sends a POST request to /history
    Postcondition: Returns a list of borrowed books

UC8: Retrieve Specific User Info
    Actor: Admin
    Precondition: Admin is authenticated
    Main Flow: Sends a POST request to /user with user ID
    Postcondition: Specific user’s data is returned


Books
UC9: List All Books
    Actor: User
    Precondition: None
    Main Flow: Sends a GET request to /books
    Postcondition: Returns list of books

UC10: Get Book by ID
    Actor: User
    Precondition: Valid book ID
    Main Flow: Sends a GET request to /book
    Postcondition: Book details returned

UC11: Search for Books
    Actor: User
    Precondition: None
    Main Flow: Sends a POST request to /searchbook with search criteria
    Postcondition: Matching books returned

UC12: Check Book Availability
    Actor: User
    Precondition: Book ID provided
    Main Flow: Sends a GET request to /book/availability
    Postcondition: Availability status returned

UC13: Add Book
    Actor: Librarian
    Precondition: Valid authentication
    Main Flow: Sends a POST request to /book with book details
    Postcondition: Book added to library

UC14: Update Book ID
    Actor: Librarian
    Precondition: Book exists
    Main Flow: Sends a POST request to /book/update/id
    Postcondition: Book ID updated

UC15: Update Book Quantity
    Actor: Librarian
    Precondition: Book exists
    Main Flow: Sends a POST request to /book/quantity
    Postcondition: Inventory updated

UC16: Remove Book
    Actor: Librarian
    Precondition: Book exists
    Main Flow: Sends a POST request to /book/remove
    Postcondition: Book is removed

UC17: Borrow Book
    Actor: User
    Precondition: Book is available
    Main Flow: Sends a POST request to /book/borrow
    Postcondition: Borrow entry created

UC18: Return Book
    Actor: User
    Precondition: Book is borrowed
    Main Flow: Sends a POST request to /book/return
    Postcondition: Book marked as returned

UC19: Reserve Book
    Actor: User
    Precondition: Book is not currently available
    Main Flow: Sends a POST request to /book/reserve
    Postcondition: Reservation created

UC20: Cancel Reservation
    Actor: User
    Precondition: Reservation exists
    Main Flow: Sends a POST request to /book/cancel/reservation
    Postcondition: Reservation cancelled

UC21: Get Overdue Books
    Actor: Librarian
    Precondition: Books are overdue
    Main Flow: Sends a POST request to /book/overdue
    Postcondition: List of overdue books returned



Reviews & Ratings
UC22: Leave Review
    Actor: User
    Precondition: Book must be borrowed or read
    Main Flow: Sends a POST request to /review with review content
    Postcondition: Review is created

UC23: Edit Review
    Actor: User
    Precondition: Review exists and belongs to the user
    Main Flow: Sends a PUT request to /review
    Postcondition: Review is updated

UC24: Delete Review
    Actor: User
    Precondition: Review exists and belongs to the user
    Main Flow: Sends a DELETE request to /review
    Postcondition: Review is deleted

UC25: Get Reviews for Book
    Actor: User
    Precondition: Book ID provided
    Main Flow: Sends a POST request to /book/review
    Postcondition: Reviews returned

UC26: Get Reviews by User
    Actor: User
    Precondition: User ID provided
    Main Flow: Sends a POST request to /review/user
    Postcondition: List of user's reviews returned

UC27: Get Book Rating
    Actor: User
    Precondition: Book ID provided
    Main Flow: Sends a POST request to /book/rating
    Postcondition: Average rating returned

UC28: Get Top Reviews
    Actor: User
    Precondition: Book ID provided
    Main Flow: Sends a POST request to /book/review/high
    Postcondition: Highest-rated reviews returned

UC29: Get Lowest Reviews
    Actor: User
    Precondition: Book ID provided
    Main Flow: Sends a POST request to /book/review/low
    Postcondition: Lowest-rated reviews returned

UC30: Vote on Review
    Actor: User
    Precondition: Review exists
    Main Flow: Sends a POST request to /review/vote
    Postcondition: Vote is recorded

UC31: Get Review Votes
    Actor: Librarian or Admin
    Precondition: Review ID provided
    Main Flow: Sends a POST request to /review/votes
    Postcondition: Vote count returned

UC33: Get Book Rating Details
    Actor: User
    Precondition: Book exists
    Main Flow: Sends a POST request to /book/rating/details
    Postcondition: Detailed ratings returned



Authors
UC35: Get Authors
    Actor: User
    Precondition: None
    Main Flow: Sends a GET request to /authors
    Postcondition: List of authors returned


Events
UC36: View Events
    Actor: User
    Precondition: None
    Main Flow: Sends a GET request to /events
    Postcondition: Events listed

UC37: View Upcoming Events
    Actor: User
    Precondition: None
    Main Flow: Sends a GET request to /event/upcoming
    Postcondition: Upcoming events returned

UC38: Create Event
    Actor: Librarian
    Precondition: User is authorized
    Main Flow: Sends a POST request to /event with event details
    Postcondition: Event created

UC39: Invite to Event
    Actor: Librarian
    Precondition: Event exists
    Main Flow: Sends a POST request to /event/invite
    Postcondition: User invited

UC40: Get Invited Events
    Actor: User
    Precondition: User has invitations
    Main Flow: Sends a POST request to /event/invited
    Postcondition: List of invited users to an event returned



Librarians
UC41: Get Librarians
    Actor: User or Admin
    Precondition: None
    Main Flow: Sends a GET request to /librarians
    Postcondition: Librarian list returned


UC42: View Borrowing History
    Actor: Librarian
    Precondition: Authenticated
    Main Flow: Sends a POST request to /user/history
    Postcondition: Borrowing history of the user returned

Other
UC43: Access API Root
    Actor: Guest
    Precondition: None
    Main Flow: Sends a request to /
    Postcondition: Health check response returned
