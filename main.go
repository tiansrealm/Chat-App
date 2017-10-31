//Creator: Tian lin (tcl344)
//Main function of twitter like app for
//Parallel & Dist system class

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//==================================================================================

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

//===============================================================================
const USER_COOKIE_NAME = "user_cookie_name"

type User struct {
	uname    string   //username
	psw      string   //password
	messages []string //twitter posts
}

//Map of all users
//key is username, value is pointer to user
var user_map map[string]*User = make(map[string]*User)

var my_user *User

//=====================================================================================
//helper functions

func is_cookie_exist(r *http.Request) bool {
	//checks if there's a user cookie
	cookie, err := r.Cookie(USER_COOKIE_NAME)
	if err != nil {
		log.Println(err)
	}
	return cookie != nil
}

func load_html(html_file_name string) string {
	//read html file, checks for err, and returns data as string
	page, err := ioutil.ReadFile(html_file_name)
	if err != nil {
		log.Fatal(err)
	}
	return string(page)
}

func is_valid_session(r *http.Request) bool {
	//checks cookie and my_user to to see if user recently logged in
	//return true if valid session, my user exist and has a cookie

	cookie, err := r.Cookie(USER_COOKIE_NAME)
	if err != nil {
		log.Println(err)
	}
	valid_session := (is_cookie_exist(r) && my_user != nil &&
		my_user.uname == cookie.Value)
	return valid_session
}
func log_out(w http.ResponseWriter, r *http.Request) {
	//clears my user and return to login page
	my_user = nil
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

//================================================================================
//html handling functions

func login(w http.ResponseWriter, r *http.Request) {
	//handles logining in to account

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
		// check if valid user
		is_valid := false
		pending_user, is_exist := user_map[r.PostFormValue("uname")]

		if is_exist && pending_user.psw == r.PostFormValue("psw") {
			is_valid = true
		}
		if is_valid {
			my_user = pending_user
			//create and store cookie for user
			cookie := http.Cookie{
				Name:    USER_COOKIE_NAME,
				Value:   my_user.uname,
				Expires: time.Now().Add(1 * time.Hour),
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		} else {
			fmt.Fprintf(w, "Your username and password combination is not found")
		}
	}
}

func sign_up(w http.ResponseWriter, r *http.Request) {
	//handles signing up an account

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
		//check is user exist
		_, is_exist := user_map[r.PostFormValue("uname")]
		if is_exist {
			http.Redirect(w, r, "/fail_sign_up", http.StatusTemporaryRedirect)
		} else {
			// Create new user
			user_map[r.PostFormValue("uname")] =
				&User{r.PostFormValue("uname"), r.PostFormValue("psw"), []string{}}

			w.Header().Set("method", "GET")
			http.Redirect(w, r, "/success_sign_up", http.StatusTemporaryRedirect)
		}
	}
}

func success_sign_up(w http.ResponseWriter, r *http.Request) {
	//tells user he has succeeded in signing up
	fmt.Fprintf(w, success_page, "You have sucessfully signed up!")
}

func fail_sign_up(w http.ResponseWriter, r *http.Request) {
	//display the fail to sign up message
	fmt.Fprintf(w, load_html("./fail_sign_up.html"))
}

//======the following http handlers requires user to be logged in
//user's log in is check by check_session func

func home(w http.ResponseWriter, r *http.Request) {
	//this func display the home page where you have various functions
	// to post, browse posts, log out, delete account
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	fmt.Fprintf(w, load_html("./home.html"))
}

func sucess_new_post(w http.ResponseWriter, r *http.Request) {
	//store new post and respond with a success message

	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		//need post from home screen
		http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
	case http.MethodPost:
		r.ParseForm()
		my_user.messages = append(my_user.messages, r.PostFormValue("message"))
		user_map[my_user.uname].messages = my_user.messages
		fmt.Fprintf(w, "successlly posted new message. Please go back")
	}
}

func browse(w http.ResponseWriter, r *http.Request) {
	//displays messsages searched by user from the home page

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
		searched_user, ok := user_map[r.PostFormValue("uname")]
		if ok {
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
			//show recent ten messages from the searched user
			print_data := ""
			message_array := searched_user.messages
			if len(message_array) > 10 {
				message_array = message_array[len(message_array)-10:]
			}
			for _, message := range message_array {
				print_data = print_data + "<p>" + message + "</p>"
			}
			fmt.Fprintf(w, page, searched_user.uname, print_data)
		} else {
			fmt.Fprintf(w, "No such user exists. Please go back.")
		}

	}

}

func delete_account(w http.ResponseWriter, r *http.Request) {
	//linked from the home page. This page confirm intent to delete
	//by asking for username and password again
	if !is_valid_session(r) {
		log_out(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, load_html("./delete_account.html"))

	case http.MethodPost:
		r.ParseForm()
		user, ok := user_map[my_user.uname]
		if !ok {
			log.Fatal("Deleting user that doesn't exist")
		}
		if user.psw == r.PostFormValue("psw") {
			//delete user
			my_user = nil
			fmt.Fprintf(w, success_page, "You have sucessfully deleted account!")
		} else {
			fmt.Fprintf(w, "Incorrect Password")
		}
	}
}

//===================================================================================
func main() {
	//sets up web app
	user_map["admin"] = &User{"admin", "admin", []string{}} //default user for faster testing
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
