package main

import (
    "fmt"
    "net/http"
	"io/ioutil"
	"log"
)

type User struct{
	uname string  //username
	psw string  //password
}

var user_map map[string]User

var my_user User

var messages map[string]string //key is username, value is message

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
		if is_valid{
			my_user = pending_user
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
		}else{
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
		if is_exist{
			http.Redirect(w, r, "/fail sign up", http.StatusTemporaryRedirect)
		}else {
			// Create new user
			user_map[r.PostFormValue("uname")] = 
				User{ r.PostFormValue("uname"), r.PostFormValue("psw") }
	
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

func main() {
	user_map = make(map[string]User)
	http.HandleFunc("/", login)
	http.HandleFunc("/sign up", sign_up)
	http.HandleFunc("/success sign up", success_sign_up)
	http.HandleFunc("/fail sign up", fail_sign_up)
	http.HandleFunc("/home", home)
	http.ListenAndServe(":8080", nil)
}


