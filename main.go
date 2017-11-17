/*Creator: Tian lin (tcl344)
Parallel & Dist system class
user interface of twitter like app
handles client side operations*/

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

//===============================================================================
const USER_COOKIE_NAME = "user_cookie_name"

//used to sperate user messages
const USER_MESSAGE_SEPERATOR = "<end of message>"

//protocol
const END_TAG = "<end>"
const SUCCESS_RESPONSE = "success"
const ERROR_RESPONSE = "error"
const DOES_USER_EXIST = "does user exist"
const CHECK_PASS = "check password"
const DELETE_USER = "delete user"
const ADD_USER = "add user"
const ADD_MESSAGE = "add message"
const READ_MESSAGES = "read messages"

//server info
const SERVER_PORT = "8083"

type User struct {
	uname string //username
	psw   string //password
}

var my_user *User

//-----------------------helper functions ------------------------------------

//checks if there's a user cookie
func is_cookie_exist(r *http.Request) bool {
	cookie, _ := r.Cookie(USER_COOKIE_NAME)
	return cookie != nil
}

//read html file, checks for err, and returns data as string
func load_html(html_file_name string) string {
	page, err := ioutil.ReadFile(html_file_name)
	if err != nil {
		log.Fatal(err)
	}
	return string(page)
}

/*checks cookie and my_user to to see if user recently logged in
return true if valid session.
valid session means my user exist and there exist a cookie the
matches the username of the user*/
func is_valid_session(r *http.Request) bool {
	cookie, _ := r.Cookie(USER_COOKIE_NAME)
	valid_session := (cookie != nil && my_user != nil &&
		my_user.uname == cookie.Value)
	return valid_session
}

//existing user logs out
func log_out(w http.ResponseWriter, r *http.Request) {
	//clears my user and return to login page
	my_user = nil
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

/*
	PostCond:  returns server's response as an
	array of string representing args

	Server response is expected to be in format
	success/failre: message
*/
func query_server(args []string) []string {
	//connect to server
	serverConn, err := net.Dial("tcp", "localhost:"+SERVER_PORT)
	defer serverConn.Close()
	if err != nil {
		log.Fatal("Failed connect to server\n")
	}
	server_scanner := bufio.NewScanner(serverConn)

	//----------------------sending query
	query := ""
	if len(args) == 2 {
		query = fmt.Sprintf("%s:%s %s\n", args[0], args[1], END_TAG)
	} else if len(args) == 3 {
		query = fmt.Sprintf("%s:%s:%s %s\n", args[0], args[1], args[2], END_TAG)
	} else if len(args) == 4 {
		query = fmt.Sprintf("%s:%s:%s:%s %s\n", args[0], args[1], args[2], args[3], END_TAG)
	} else {
		log.Fatal("invalid number of args for query")
	}
	//fmt.Printf("sending: %s\n", query)
	fmt.Fprintf(serverConn, query)

	//-------------recieving response
	response := ""
	for server_scanner.Scan() {
		newLine := server_scanner.Text()
		response += newLine //append to read query
		index_of_endtag := strings.Index(newLine, END_TAG)
		if index_of_endtag != -1 {
			//reached end of response
			break
		}
	}

	response = strings.Replace(response, END_TAG, "", 1) //remove end tag
	delimiter := ":"
	parsed_response := strings.Split(response, delimiter)
	//trims white space at the ends
	for i := 0; i < len(parsed_response); i++ {
		parsed_response[i] = strings.TrimSpace(parsed_response[i])
	}
	return parsed_response
}

//-----------------http+ handling functions------------------------------------------------

/*handles logining in to account.
Get displays the log in page.
Post will authentic user, log in to home page, and create cookie*/
func login(w http.ResponseWriter, r *http.Request) {
	//check if user is logged in with cookies
	if is_valid_session(r) {
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case http.MethodGet:
		//display page for logging in if not logged in
		fmt.Fprintf(w, load_html("./login.html"))

	case http.MethodPost:
		r.ParseForm()
		uname := r.PostFormValue("uname")
		psw := r.PostFormValue("psw")

		//send query to  check if valid user
		args := []string{CHECK_PASS, uname, psw}
		response := query_server(args)
		if response[0] != SUCCESS_RESPONSE {
			//negative response so print the error message
			fmt.Fprintf(w, response[1])
		} else {
			my_user = &User{uname, psw}
			//create and store cookie for user
			cookie := http.Cookie{
				Name:    USER_COOKIE_NAME,
				Value:   my_user.uname,
				Expires: time.Now().Add(1 * time.Hour),
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		}
	}
}

/*handles signing up an account.
Get display the sign up page.
Post will read the username and password and
send a request to server for adding a user*/
func sign_up(w http.ResponseWriter, r *http.Request) {
	//check if user is logged in with cookies
	if is_valid_session(r) {
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case http.MethodGet:
		//display page for signing up
		fmt.Fprintf(w, load_html("./sign_up.html"))

	case http.MethodPost:
		r.ParseForm()
		//send query asking if user exist
		uname := r.PostFormValue("uname")
		psw := r.PostFormValue("psw")
		exist_args := []string{DOES_USER_EXIST, uname}
		exist_response := query_server(exist_args)
		fmt.Println(exist_response)
		is_user_exist := (exist_response[0] == SUCCESS_RESPONSE)

		if is_user_exist {
			http.Redirect(w, r, "/fail_sign_up", http.StatusTemporaryRedirect)
		} else {
			// send query to create new user
			add_user_args := []string{ADD_USER, uname, psw}
			add_response := query_server(add_user_args)
			if add_response[0] == SUCCESS_RESPONSE {
				//w.Header().Set("method", "GET")
				http.Redirect(w, r, "/success_sign_up", http.StatusTemporaryRedirect)
			} else {
				//user doesn't already exist but still got error adding it
				fmt.Fprintf(w, "Server Side Error")
			}
		}
	}
}

//show html page telling user he has succeeded in signing up
func success_sign_up(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, success_page, "You have sucessfully signed up!")
}

//display the fail to sign up message
func fail_sign_up(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, load_html("./fail_sign_up.html"))
}

//======the following http handlers requires user to be logged in
//user's log in is check by is_valid_session

/*Dsplay the home page to leads to various functions
to post message by writing in textbox and submitting,
 browse posts, log out, delete account*/
func home(w http.ResponseWriter, r *http.Request) {
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	fmt.Fprintf(w, load_html("./home.html"))
}

/*store new post and respond with a success message.
Accessed from posting with text box from home page*/
func sucess_new_post(w http.ResponseWriter, r *http.Request) {
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		//need to post from home screen
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
	case http.MethodPost:
		r.ParseForm()

		add_message_args := []string{ADD_MESSAGE, my_user.uname, my_user.psw, r.PostFormValue("message")}
		add_message_response := query_server(add_message_args)
		is_success := (add_message_response[0] == SUCCESS_RESPONSE)

		if is_success {
			fmt.Fprintf(w, "successlly posted new message. Please go back")
		} else {
			fmt.Fprintf(w, "Failed to post new message. Error response is: %s.", add_message_response[1])
		}
	}
}

/*Accessed from homepage
Displays html page with messsages searched by user
input from html page was a username. Sends a read request under that username to the server.*/
func browse(w http.ResponseWriter, r *http.Request) {
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		//shouldnt be here
		fmt.Fprintf(w, "please browse from home page")
	case http.MethodPost:
		r.ParseForm()
		search_uname := r.PostFormValue("uname")
		exist_args := []string{DOES_USER_EXIST, search_uname}
		exist_response := query_server(exist_args)
		is_user_exist := (exist_response[0] == SUCCESS_RESPONSE)
		//
		if is_user_exist {
			page := `<!DOCTYPE html>
				<html lang="en">
				<head>
				    <meta charset="UTF-8">
				    <title>Chat App</title>
				</head>			
				<body>
				    <h1>Browsing post from %s</h1>
				    %s 
				</body>				
				</html>`
			//send a read request for reading 10 messages
			read_messages_args := []string{READ_MESSAGES, search_uname, "10"}
			read_messages_response := query_server(read_messages_args)
			is_success := (read_messages_response[0] == SUCCESS_RESPONSE)
			if is_success {
				//print messages on web page
				messages := read_messages_response[1]
				//put 2 newlines after every message
				messages = strings.Replace(messages, USER_MESSAGE_SEPERATOR, "<br /><br />", -1)
				fmt.Fprintf(w, page, search_uname, messages)
			} else {
				fmt.Fprintf(w, "Failed to read messages. Server responded: %s.", read_messages_response[1])
			}
		} else {
			fmt.Fprintf(w, "No such user exists. Please go back.")
		}

	}

}

/*linked from the home page. This page confirm intent to delete
by asking for username and password again*/
func delete_account(w http.ResponseWriter, r *http.Request) {
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, load_html("./delete_account.html"))

	case http.MethodPost:
		r.ParseForm()
		if my_user == nil {
			log.Fatal("My user don't exist")
		}

		pending_psw := r.PostFormValue("psw")
		if my_user.psw != pending_psw {
			//failed password authentication
			fmt.Fprintf(w, "Incorrect Username or Password. Please go back and retry.")

		} else {
			delete_user_args := []string{DELETE_USER, my_user.uname, my_user.psw, r.PostFormValue("message")}
			delete_user_response := query_server(delete_user_args)
			is_success := (delete_user_response[0] == SUCCESS_RESPONSE)

			if is_success {
				my_user = nil //clear username and password in  memory
				fmt.Fprintf(w, success_page, "You have sucessfully deleted account!")
			} else {
				fmt.Fprintf(w, "Failed to deleted user. Error response is: %s.", delete_user_response[1])
			}
		}

	}
}

//===================================================================================

//main function sets up variables and http handlers
func main() {
	my_user = nil
	http.HandleFunc("/", login)
	http.HandleFunc("/sign_up", sign_up)
	http.HandleFunc("/success_sign_up", success_sign_up)
	http.HandleFunc("/fail_sign_up", fail_sign_up)
	http.HandleFunc("/home", home)
	http.HandleFunc("/sucess_new_post", sucess_new_post)
	http.HandleFunc("/browse", browse)
	http.HandleFunc("/log_out", log_out)
	http.HandleFunc("/delete_account", delete_account)
	http.ListenAndServe(":8080", nil)
}

//==================================================================================
//extra html pages
//success page with a message %s and option to go to login page
var success_page string = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <title>Chat App</title>
</head>

<body>
    <h1>%s</h1>
    <br />
     <form>
    <button formaction="/">Return to Log In Page</button>
	</form>
</body>

</html>
`
