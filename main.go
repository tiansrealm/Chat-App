package main

import (
    "fmt"
    "net/http"
	"io/ioutil"
	"log"
)

type User struct{
	name string
	passw string
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
		// Print the values of the form
		fmt.Fprintf(w, "post")
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

		fmt.Fprintf(w, "Thanks for signing up")
		http.Get("google.com")
	}
}

func thanks_sign_up(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := ioutil.ReadFile("./thanks_sign_up.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(page))
	case http.MethodPost:
		
		http.Get("http://localhost:8080/")
	}
}

func main() {
	user_map = make(map[string]User)
	http.HandleFunc("/", login)
	http.HandleFunc("/thanks_sign_up", thanks_sign_up)
	http.HandleFunc("/sign_up", sign_up)
	http.ListenAndServe(":8080", nil)
}


