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

func login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		login, err := ioutil.ReadFile("./login.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(login))
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
		sign_up, err := ioutil.ReadFile("./sign_up.html")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(w, string(sign_up))
	case http.MethodPost:
		// Parse the form
		r.ParseForm()
		// Print the values of the form
		fmt.Fprintf(w, "post")
	}
}
var user_map map[string]User
func main() {
	user_map = make(map[string]User)
	http.HandleFunc("/", login)
	http.HandleFunc("/sign_up", sign_up)
	http.ListenAndServe(":8080", nil)
}


