Users:
    As a user, I want to retrieve all users so that I can view the list of registered members. (GET /users)
    As a user, I want to sign up so that I can access library services. (POST /signup)
    As a user, I want to log in so that I can authenticate and use the system. (POST /login)
    As a user, I want to view my profile so that I can see my account information. (POST /profile)
    As a user, I want to edit my profile so that I can keep my information up to date. (POST /edit)
    As a user, I want to delete my account so that I can remove my data from the system. (DELETE /user)
    As an admin, I want to retrieve a specific user’s data so that I can review their activity. (POST /user)

Books
    As a user, I want to view all available books so that I can choose what to read. (GET /books)
    As a user, I want to view the details of a specific book so that I can learn more about it. (GET /book)
    As a user, I want to search for a book so that I can find books that match my interest. (POST /searchbook)
    As a user, I want to know if a book is available so that I can plan to borrow or reserve it. (GET /book/availability)
    As a librarian, I want to add a new book so that it becomes available in the catalog. (POST /book)
    As a librarian, I want to update a book's ID so that I can maintain catalog accuracy. (POST /book/update/id)
    As a librarian, I want to update a book's quantity so that the inventory reflects reality. (POST /book/quantity)
    As a librarian, I want to remove a book so that it is no longer borrowable. (POST /book/remove)
    As a user, I want to borrow a book so that I can read it. (POST /book/borrow)
    As a user, I want to return a book so that I complete my borrowing. (POST /book/return)
    As a user, I want to reserve a book so that I can borrow it immediately after it's available. (POST /book/reserve)
    As a user, I want to cancel my book reservation so that I release the hold. (POST /book/cancel/reservation)
    As a librarian, I want to view books that are overdue so that I can follow up. (POST /book/overdue)

Reviews & Ratings
    As a user, I want to leave a review on a book so that I can share my opinion. (POST /review)
    As a user, I want to edit my review so that I can update my thoughts. (PUT /review)
    As a user, I want to delete my review so that it is no longer visible. (DELETE /review)
    As a user, I want to view reviews for a book so that I can learn from others' experiences. (POST /book/review)
    As a librarian, I want to view peoples reviews so that I can track if they write responsibly. (POST /review/user)
    As a user, I want to view a book’s rating so that I can evaluate its quality. (POST /book/rating)
    As a user, I want to see the highest-rated reviews so that I can read the most helpful feedback. (POST /book/review/high)
    As a user, I want to see the lowest-rated reviews so that I can be aware of possible downsides. (POST /book/review/low)
    As a user, I want to vote for a review so that I can signal its helpfulness. (POST /review/vote)
    As a user, I want to view how many votes a review has so that I know its popularity. (POST /review/votes)
    As a user, I want detailed book rating information so that I can make informed choices. (POST /book/rating/details)

Authors
    As a user, I want to view a list of authors so that I can find books by them. (GET /authors)

Librarians
    As a user or admin, I want to view librarians so that I know who manages the system. (GET /librarians)
    As a librarian, I want to view all peoples borrowing history so that I can track the books they have interacted with. (POST /user/history)

 History
    As a user, I want to retrieve my borrowing history so that I can reflect on past reads. (POST /history)

Other
    As a visitor, I want to access the homepage so that I know the API is running. (ANY /)
