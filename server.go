//Creator: Tian lin (tcl344)
//server of twitter like app for
//Parallel & Dist system class

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type User struct {
	uname    string   //username
	psw      string   //password
	messages []string //twitter posts
}

//Map of all users
//key is username, value is pointer to user
var user_map map[string]*User = make(map[string]*User)

const USERLIST_FILENAME = "user_list.txt"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	loadUserData()
	in, _ := net.Listen("tcp", ":8083")
	defer in.Close()

	for {
		conn, _ := in.Accept()
		fmt.Fprintln(os.Stderr, "Got a connection")
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			query := scanner.Text()
			fmt.Println("got " + query)
			evaluate(query, conn)

		}
		//interupted
		fmt.Fprintln(os.Stderr, "connection gone!")
		conn.Close()
	}

}

func loadUserData() {

	file, err := os.OpenFile(USERLIST_FILENAME, os.O_CREATE|os.O_RDONLY, 0600)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted_string := strings.Split(line, ":")
		uname := strings.TrimSpace(splitted_string[0])
		psw := strings.TrimSpace(splitted_string[1])
		// store user in server cache
		user_map[uname] = &User{uname, psw, []string{}}
	}

}

func add_user(uname string, psw string) {
	//create new user and write to database file
	user_map[uname] = &User{uname, psw, []string{}}
	//open file to write to end of it
	file, err := os.OpenFile(USERLIST_FILENAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)
	defer file.Close()
	text := uname + " : " + psw + "\r\n"
	if _, err = file.WriteString(text); err != nil {
		panic(err)
	}
}
func evaluate(query string, conn net.Conn) {
	//handles queries
	//separate string into args by spliting along delimiter ":"
	delimiter := ":"
	parsed_query := strings.Split(query, delimiter)
	//trims white space
	for i := 0; i < len(parsed_query); i++ {
		parsed_query[i] = strings.TrimSpace(parsed_query[i])
	}
	switch parsed_query[0] {
	case "create":
		//args should be query, username and password
		if len(parsed_query) != 2 {
			fmt.Fprintf(conn, "Error: Not valid args\n")
			return
		}
		fmt.Fprintf(conn, "I got %s\n", parsed_query[0])
	case "add user":
		//args should be query, username, password
		if len(parsed_query) != 3 {
			fmt.Fprintf(conn, "Error: Not valid args\n")
			return
		}
		add_user(parsed_query[1], parsed_query[2])
		fmt.Fprintf(conn, "I added user %s\n", parsed_query[1])
	default:
		fmt.Fprintf(conn, "%s is not a valid query\n", parsed_query[0])

	}

}
func writeToFile(filename string) {

}

func createFile(filename string) {

}
