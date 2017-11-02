//Creator: Tian lin (tcl344)
//Backend for twitter like app for
//Parallel & Dist system class

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type User struct {
	uname string //username
	psw   string //password
}

//Map of all users
//key is username, value is pointer to user

var user_map map[string]*User = make(map[string]*User)

const USERLIST_FILENAME = "user_list.txt"

//used to sperate user messages in files
const USER_MESSAGE_DELIMITER = "<end of message>"

//used to signify end of a communication message
const END_TAG = "<end>"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//===================================================================
func main() {
	/*main function initialize server socket and listens for queries*/

	//load user list for faster access to a list of current users
	load_user_list()
	in, _ := net.Listen("tcp", ":8083")
	defer in.Close()

	for {
		conn, _ := in.Accept()
		fmt.Fprintln(os.Stderr, "Got a connection")
		scanner := bufio.NewScanner(conn)
		query := ""
		for scanner.Scan() {
			fmt.Println("-------------------------------------------")
			newLine := scanner.Text()
			index_of_endtag := strings.Index(newLine, END_TAG)
			if index_of_endtag != -1 {
				//reached end
				query += newLine[0 : len(newLine)-len(END_TAG)] //append without end tag
				fmt.Println(query)
				// finished_query := query
				// query = "" //reset query
				// evaluate(finished_query, conn)
				evaluate(query, conn)
				query = "" //reset query
			} else {
				query += newLine //append to read query
			}
		}
		//interupted
		fmt.Fprintln(os.Stderr, "connection gone!") //print ln have \n
		conn.Close()
	}

}

func load_user_list() {
	/*loads list of existing users from file database into server memory*/
	file, err := os.OpenFile(USERLIST_FILENAME, os.O_CREATE|os.O_RDONLY, 0600)
	check(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted_string := strings.Split(line, ":")
		uname := strings.TrimSpace(splitted_string[0])
		psw := strings.TrimSpace(splitted_string[1])
		// store user in server memory
		user_map[uname] = &User{uname, psw}
	}

}

func evaluate(query string, conn net.Conn) {
	/*
		handles queries
		separate string into args by spliting along delimiter ":"
		Pre-Condition:
		Request will be in the form of "queryname: arg1: arg2: ..."
		Post Condition:
		response with be in the form of  "success/error: response"
	*/
	delimiter := ":"
	parsed_query := strings.Split(query, delimiter)
	//trims white space
	for i := 0; i < len(parsed_query); i++ {
		parsed_query[i] = strings.TrimSpace(parsed_query[i])
	}
	switch parsed_query[0] {
	case "does user exist":
		//args should be query, username
		check_args(parsed_query, 2, conn)
		_, is_exist := user_map[parsed_query[1]]
		if is_exist {
			fmt.Fprintf(conn, "success: true %s \r\n", END_TAG)
		} else {
			fmt.Fprintf(conn, "success: true %s \r\n", END_TAG)
		}
	case "add user":
		//args should be query, username, password
		check_args(parsed_query, 3, conn)
		add_user(parsed_query[1], parsed_query[2], conn)
	case "add message":
		//args should be query, username, message
		check_args(parsed_query, 3, conn)
		add_message(parsed_query[1], parsed_query[2], conn)
	case "read messages":
	default:
		fmt.Fprintln(os.Stderr, "I sent error\n")
		fmt.Fprintf(conn, "error: %s is not a valid query.%s\n", parsed_query[0], END_TAG)

	}

}

//==========================================================================================
//Functions that respond to queries, used in Evalute
//
func check_args(parsed_query []string, num_expected int, conn net.Conn) {
	if len(parsed_query) != num_expected {
		fmt.Fprintf(conn, "error: Not valid # of args%s\n", END_TAG)
		return
	}
}
func add_user(uname string, psw string, conn net.Conn) {
	/*Create new user and write new user info to user list file*/

	_, is_exist := user_map[uname]
	if !is_exist {
		//create user if not exist
		user_map[uname] = &User{uname, psw}
		//open user list file to write to end of it
		file, err := os.OpenFile(USERLIST_FILENAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		check(err)
		defer file.Close()
		text := uname + " : " + psw + "\r\n"
		if _, err = file.WriteString(text); err != nil {
			panic(err)
		}
		//create user data file
		u_file_name := uname + ".txt"
		os.Remove(u_file_name) // clear old junk
		created_file, err := os.Create(u_file_name)
		check(err)
		defer created_file.Close()
		//response
		fmt.Fprintf(conn, "success: I added user %s.%s\n", uname, END_TAG)
	} else {
		//negative response
		fmt.Fprintf(conn, "error: user, %s, already exists.%s\n", uname, END_TAG)
	}
}

func add_message(uname string, new_message string, conn net.Conn) {
	/*
		Add a new message under the user with given uname, by
		writing to database file containing stored messsages the user
	*/
	filename := uname + ".txt"
	message_file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)
	defer message_file.Close()

	//write new message to file
	text_to_write := new_message + USER_MESSAGE_DELIMITER
	if _, err = message_file.WriteString(text_to_write); err != nil {
		fmt.Fprintf(conn, "error: server failed to write\n")
		panic(err)
	} else {
		fmt.Fprintf(conn, "success: added message for%s\n", uname)
	}

}

func read_messages(uname string, num_messages int, conn net.Conn) {
	/*reads messages from user file database*/

	filename := uname + ".txt"
	message_file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0600)
	check(err)
	defer message_file.Close()

	messages_in_byte, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	messages_in_string := string(messages_in_byte)

	message_array := strings.Split(messages_in_string, USER_MESSAGE_DELIMITER)

	recent_messages := message_array[len(message_array)-num_messages:]
	fmt.Fprintf(os.Stderr, "printing message%s\n", recent_messages)
	//give back most recent num_messages of messages
}
