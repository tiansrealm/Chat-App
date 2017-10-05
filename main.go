package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type User struct {
	uname    string //username
	psw      string //password
	messages []string
}

var user_map map[string]*User = make(map[string]*User)

var my_user *User

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := ioutil.ReadFile("./login.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
	case http.MethodPost:
		// Parse the form
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
		page, err := ioutil.ReadFile("./sign up.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
	case http.MethodPost:
		// Parse the form
		r.ParseForm()
		//check is user exist
		_, is_exist := user_map[r.PostFormValue("uname")]
		if is_exist {
			http.Redirect(w, r, "/fail sign up", http.StatusTemporaryRedirect)
		} else {
			// Create new user
			user_map[r.PostFormValue("uname")] =
				&User{r.PostFormValue("uname"), r.PostFormValue("psw"), []string{}}

			w.Header().Set("method", "GET")
			http.Redirect(w, r, "/success sign up", http.StatusTemporaryRedirect)
		}
	}
}

func success_sign_up(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("./success sign up.html")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(page))
}
func fail_sign_up(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("./fail sign up.html")
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
		// Parse the form
		r.ParseForm()
		my_user.messages = append(my_user.messages, r.PostFormValue("message"))
		user_map[my_user.uname].messages = my_user.messages
		//fmt.Printf("user has %d message\n", len(my_user.messages))
		fmt.Fprintf(w, "successlly posted new message. Please go back")
	}
}
func browse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//just print everyone's post
		fmt.Fprintf(w, "please browse from home page")
	case http.MethodPost:
		// Parse the form
		r.ParseForm()
		searched_user := user_map[r.PostFormValue("uname")]
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
		//show last ten messages from the searched user
		print_data := ""
		message_array := searched_user.messages
		//fmt.Printf("There are %d messages\n", len(message_array))
		if len(message_array) > 10 {
			message_array = message_array[len(message_array)-10:]
		}
		for _, message := range message_array {
			print_data = print_data + "<p>" + message + "</p>"
		}
		fmt.Fprintf(w, page, searched_user.uname, print_data)
	}

}
func main() {
	user_map["admin"] = &User{"admin", "admin", []string{}} //default user for faster testing
	http.HandleFunc("/", login)
	http.HandleFunc("/sign up", sign_up)
	http.HandleFunc("/success sign up", success_sign_up)
	http.HandleFunc("/fail sign up", fail_sign_up)
	http.HandleFunc("/home", home)
	http.HandleFunc("/sucess new post", sucess_new_post)
	http.HandleFunc("/browse", browse)
	http.ListenAndServe(":8080", nil)
}
