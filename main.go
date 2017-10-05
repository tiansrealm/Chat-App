package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type User struct {
	uname string //username
	psw   string //password
}

var user_map map[string]User = make(map[string]User)

var my_user User

var message_map map[string][]string = make(map[string][]string)

//key is username, value is array of messages from that user

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
		fmt.Println("inside sign up")
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
				User{r.PostFormValue("uname"), r.PostFormValue("psw")}

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
		message_map[my_user.uname] = append(message_map[my_user.uname], r.PostFormValue("message"))
		fmt.Fprintf(w, "successlly posted new message. Please go back")
	}
}
func browse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//just print everyone's post
		for username, message_array := range message_map {
			fmt.Fprintf(w, username)
			fmt.Fprintf(w, " Posted\n")
			for _, message := range message_array {

				fmt.Fprintf(w, message)
			}
		}
	case http.MethodPost:
		// Parse the form
		r.ParseForm()
		searched_username := user_map[r.PostFormValue("uname")].uname
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
		message_array := message_map[searched_username]
		if len(message_array) > 10 {
			message_array = message_array[len(message_array)-10:]
		}
		for _, message := range message_array {
			print_data = print_data + message + "\n"
		}
		fmt.Fprintf(w, page, searched_username, print_data)
	}

}
func main() {
	http.HandleFunc("/", login)
	http.HandleFunc("/sign up", sign_up)
	http.HandleFunc("/success sign up", success_sign_up)
	http.HandleFunc("/fail sign up", fail_sign_up)
	http.HandleFunc("/home", home)
	http.HandleFunc("/sucess new post", sucess_new_post)
	http.HandleFunc("/browse", browse)
	http.ListenAndServe(":8080", nil)
}
