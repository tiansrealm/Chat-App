# Twitter-App
By - Tian Lin

Written for Parallel and Distributed System Class


Twitter like App using Go lang and http

Current Functionalities

Login Page (path = "/")
-Can log in or direct to sign up page
-checks if username and password exists and are correct
-on success will direct to home page


Sign Up Page
-Can sign up new user
-checks if user already exist and direct to a fail signed up page


Home Page
-can post a message
-can browse a user's recent messages, including your own by searching username
-can log out
-can delete account


Browse Page
-displays result from the search from home page

TO USE
go build main.go
or run main.exe if it already exists