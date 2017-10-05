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
		User, is_exist := user_map[r.PostFormValue("uname")]

		if is_exist && User.psw == r.PostFormValue("psw") {
			is_valid = true
		}
		fmt.Fprintf(w, "user validity is %t", is_valid)
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
		// Parse the form
		r.ParseForm()
		// Create new user
		user_map[r.PostFormValue("uname")] = 
			User{ r.PostFormValue("uname"), r.PostFormValue("psw") }

		w.Header().Set("method", "GET")
		http.Redirect(w, r, "/thanks_sign_up", http.StatusTemporaryRedirect)
	}
}

func thanks_sign_up(w http.ResponseWriter, r *http.Request) {
	page, err := ioutil.ReadFile("./thanks_sign_up.html")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, string(page))
}

func main() {
	user_map = make(map[string]User)
	http.HandleFunc("/", login)
	http.HandleFunc("/thanks_sign_up", thanks_sign_up)
	http.HandleFunc("/sign_up", sign_up)
	http.ListenAndServe(":8080", nil)
}


