package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
type User struct {
	uname    string   //username
	psw      string   //password
	messages []string //twitter posts
}

var user_map map[string]*User = make(map[string]*User)

//key is username

var my_user *User

//=====================================================================================
func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := ioutil.ReadFile("./login.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
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
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		} else {
			fmt.Fprintf(w, "Your username and password combination is not found")
		}
	}
}

func sign_up(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := ioutil.ReadFile("./sign_up.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
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
	fmt.Fprintf(w, success_page, "You have sucessfully signed up!")
}

func fail_sign_up(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("./fail_sign_up.html")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(page))
}

func home(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("./home.html")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(page))

}

func sucess_new_post(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "Please post from home screen")
	case http.MethodPost:
		r.ParseForm()
		my_user.messages = append(my_user.messages, r.PostFormValue("message"))
		user_map[my_user.uname].messages = my_user.messages
		fmt.Fprintf(w, "successlly posted new message. Please go back")
	}
}

func browse(w http.ResponseWriter, r *http.Request) {
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
	switch r.Method {
	case http.MethodGet:
		page, err := ioutil.ReadFile("./delete_account.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
	case http.MethodPost:
		r.ParseForm()
		searched_user, ok := user_map[r.PostFormValue("uname")]
		if ok && searched_user.psw == r.PostFormValue("psw") {
			//delete user
			delete(user_map, r.PostFormValue("uname"))

			//option to go back to login screen
			fmt.Fprintf(w, success_page, "Successfully deleted account.")
		} else {
			fmt.Fprintf(w, "Incorrect Username and Password combination")
		}
	}
}

func log_out(w http.ResponseWriter, r *http.Request) {
	//clears my user and return to login page
	my_user = nil
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
func main() {
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
