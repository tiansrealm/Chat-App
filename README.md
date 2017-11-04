# Twitter-App
By - Tian Lin

Written for Parallel and Distributed System Class

Twitter like App using Go lang and http

TO USE
---------------
go build server.go
run server.exe

go build main.go
run main.exe 

use web brower to access http://localhost:8080/  
which is the main page for the web app

------------------------------------------------------------------
Errors fixed from Part 1 !
------------------------------------------------------------------
1. Cookies
-app can tell if user logged in within last hour by comparing cookie and current user data
-if session expire will automatically logout
-if session valid with go to home page without logging in

2. no longer able to delete other users. 

3. Messages limited to 100 characters


------------------------------------------------------------------
Part 2 - Server Side + File Systems
------------------------------------------------------------------
Main
-----
main.go now communicates with the server
Run server before running main when testing
main no longer store any user data, 
except for the username/password of the current logged in user.

server
--------
server is coded in server.go. 

creates a running server that handles request for 
these protocols
//protocol
const DOES_USER_EXIST = "does user exist"
const CHECK_PASS = "check password"
const DELETE_USER = "delete user"
const ADD_USER = "add user"
const ADD_MESSAGE = "add message"
const READ_MESSAGES = "read messages"



Files/Database System
--------------------
users and there passwords are kept in a file called user_list.txt
Each user will have a file keeping all there messages.
The messages for userx will be kept in userx.txt
All operations to the files/database is done through server






------------------------------
web app info from Part 1 
----------------------------
main.go is a user web app.

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

