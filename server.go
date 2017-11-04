/*Creator: Tian lin (tcl344)
Parallel & Dist system class

Backend for twitter like app
listen for client connection
recieves and respondsto queries
*/
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

//Map of all users
//key is username, value is password

var user_map map[string]string = make(map[string]string)

const USERLIST_FILENAME = "user_list.txt"

//used to sperate user messages
const USER_MESSAGE_SEPERATOR = "<end of message>"

//protocol
//protocol
const END_TAG = "<end>"
const DOES_USER_EXIST = "does user exist"
const CHECK_PASS = "check password"
const DELETE_USER = "delete user"
const ADD_USER = "add user"
const ADD_MESSAGE = "add message"
const READ_MESSAGES = "read messages"

//===================================================================
//Main Functions

/*main function initialize server socket and listens for queries*/
func main() {

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
				fmt.Println("Got query: " + query)
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

/*
	handles queries for
		checking if user exist,
		adding new users,
		write/read message
	separate string into args by spliting along delimiter ":"

	Pre-Condition:
	Request will be in the form of "queryname: arg1: arg2: ..."
	Post Condition:
	response with be in the form of  "success/error: response"
*/
func evaluate(query string, conn net.Conn) {
	delimiter := ":"
	parsed_query := strings.Split(query, delimiter)
	//trims white space at the ends
	for i := 0; i < len(parsed_query); i++ {
		parsed_query[i] = strings.TrimSpace(parsed_query[i])
	}

	query_function := parsed_query[0]

	//for all queries, args should start with query, username
	//all queries have >= 2 args
	if !check_args(parsed_query, 2, conn) {
		//check args failed
		return //skip this query
	}
	uname := parsed_query[1]

	//------------------------checking query now--------------------------
	if query_function == DOES_USER_EXIST {
		//respond sucess if user exists
		if _, is_exist := user_map[uname]; is_exist {
			fmt.Fprintf(conn, "success: user exists %s\n", END_TAG)
		} else {
			fmt.Fprintf(conn, "error: no such user %s\n", END_TAG)
		}
		return
	} else {
		// check for more args
		if !check_args(parsed_query, 3, conn) {
			//check args failed
			return //skip this query
		}
	}

	//------requires >=3 args
	if query_function == ADD_USER {
		//doesn't need password authentication
		if !check_args(parsed_query, 3, conn) {
			//check args failed
			return //skip this query
		}
		add_user(uname, parsed_query[2], conn)
		return
	} else if query_function == READ_MESSAGES {
		//args should be query, username, num_message

		if num_message, convert_err := strconv.Atoi(parsed_query[2]); convert_err != nil {
			fmt.Fprintf(conn, "error: third arg must be integer.%s\n", END_TAG)
		} else {
			read_messages(uname, num_message, conn)
		}
		return
	}

	psw := parsed_query[2]
	//following functions needs password authentication
	if !authenticate(uname, psw, conn) {
		//uname and psw don't match
		return
	}

	switch query_function {
	case CHECK_PASS:
		//reply passed username + password check
		fmt.Fprintf(conn, "success: correct username and password %s\n", END_TAG)
		return
	case DELETE_USER:
		delete_user(uname, conn)
		return
	case ADD_MESSAGE:
		//args should be query, username, password, message
		if !check_args(parsed_query, 4, conn) {
			//check args failed
			return //skip this query
		}
		message := parsed_query[3]
		add_message(uname, message, conn)
		return
	}

	//didn't recongize any valid queries
	fmt.Fprintf(conn, "error: %s is not a valid query.%s\n", parsed_query[0], END_TAG)
}

//==========================================================================================
//Functions that respond to queries, used in Evalute
//============================================================================================

/*checks if num args from query is AT LEAST the num expected
sends an error response if num args don't match expected
return false if args is wrong, true otherwise*/
func check_args(parsed_query []string, num_expected int, conn net.Conn) bool {

	if !(len(parsed_query) >= num_expected) {
		fmt.Fprintf(conn, "error: Not valid # of args%s\n", END_TAG)
		return false
	}
	return true
}

/*  checks password against username and write error to conn if not match
also returns false if not match  */
func authenticate(uname string, psw string, conn net.Conn) bool {
	if _, is_exist := user_map[uname]; is_exist && user_map[uname] == psw {
		return true
	} else {
		fmt.Fprintf(conn, "error: Username and Password combination not found. %s\n", END_TAG)
		return false
	}
}

/*Create new user and write new user info to user list file
send error response if user already exists*/
func add_user(uname string, psw string, conn net.Conn) {
	_, is_exist := user_map[uname]
	if !is_exist {
		//create user if not exist
		user_map[uname] = psw
		//open user list file to write to end of it
		file, open_err := os.OpenFile(USERLIST_FILENAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		defer file.Close()
		check_err_and_repond(open_err, conn)

		text := uname + " : " + psw + "\r\n"
		if _, write_err := file.WriteString(text); write_err != nil {
			check_err_and_repond(write_err, conn)
		}
		//create user data file
		u_file_name := uname + ".txt"
		os.Remove(u_file_name) // clear old junk
		created_file, create_err := os.Create(u_file_name)
		defer created_file.Close()
		check_err_and_repond(create_err, conn)
		//response
		fmt.Fprintf(conn, "success: I added user %s.%s\n", uname, END_TAG)
	} else {
		//negative response
		fmt.Fprintf(conn, "error: user, %s, already exists.%s\n", uname, END_TAG)
	}
}

/*Add a new message under the user with given uname, by
writing to database file containing stored messsages the user*/
func add_message(uname string, new_message string, conn net.Conn) {
	filename := uname + ".txt"
	message_file, open_err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer message_file.Close()
	check_err_and_repond(open_err, conn)

	//write new message to file
	newline := "\r\n"
	text_to_write := new_message + newline + USER_MESSAGE_SEPERATOR + newline
	if _, write_err := message_file.WriteString(text_to_write); write_err != nil {
		fmt.Fprintf(conn, "error: server failed to write.%s\n", END_TAG)
		panic(write_err)
	} else {
		fmt.Fprintf(conn, "success: added message for %s.%s\n", uname, END_TAG)
	}

}

/*deletes user from userlist file and delete message file asscioated with that user
pre-condition: user exists, use authenticate func to check for that*/
func delete_user(uname string, conn net.Conn) {
	//delete user from server memory
	delete(user_map, uname)
	err := rewrite_userlist() //delete user from user list file
	check_err_and_repond(err, conn)
	//delete user message file
	os.Remove(uname + ".txt")
	//repond sucess
	fmt.Fprintf(conn, "success: Deleted user %s.%s\n", uname, END_TAG)
}

/*reads messages from user file database*/
func read_messages(uname string, num_messages int, conn net.Conn) {
	filename := uname + ".txt"
	message_file, open_err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0600)
	defer message_file.Close()
	check_err_and_repond(open_err, conn)

	messages_in_byte, read_err := ioutil.ReadFile(filename)
	check_err_and_repond(read_err, conn)

	messages_in_string := string(messages_in_byte)

	message_array := strings.SplitAfter(messages_in_string, USER_MESSAGE_SEPERATOR)
	recent_messages := message_array
	if num_messages < len(message_array) {
		//only show recent num messages if there exist more than that
		recent_messages = message_array[len(message_array)-num_messages:]
	}
	fmt.Fprintf(os.Stderr, "printing message%s\n", recent_messages)
	//give back most recent num_messages of messages
	response := ""
	for _, message := range recent_messages {
		response += message
	}
	fmt.Fprintf(conn, "success: %s%s\n", response, END_TAG)
}

//-----------------------file operations-----------------------------------------------

/*loads list of existing users from file database into server memory
for faster checks that user exist*/
func load_user_list() {
	file, err := os.OpenFile(USERLIST_FILENAME, os.O_CREATE|os.O_RDONLY, 0600)
	defer file.Close()
	check_err(err)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitted_string := strings.Split(line, ":")
		uname := strings.TrimSpace(splitted_string[0])
		psw := strings.TrimSpace(splitted_string[1])
		// store user in server memory
		user_map[uname] = psw
	}

}

/*rewrite list of existing user from server memory to userlist. Needed after removing a user*/
func rewrite_userlist() error {
	os.Remove(USERLIST_FILENAME) //delete old file
	//rewrite new user list file
	file, err := os.OpenFile(USERLIST_FILENAME, os.O_CREATE|os.O_WRONLY, 0600)
	defer file.Close()
	if err != nil {
		return err
	}
	//
	for uname, psw := range user_map {
		text := uname + " : " + psw + "\r\n"
		if _, write_err := file.WriteString(text); write_err != nil {
			return write_err
		}
	}
	return nil //no errors = success
}

//--------------common functions

//basic check for err
func check_err(err error) {
	if err != nil {
		panic(err)
	}
}

//check for error and notify error  to conn if there is
func check_err_and_repond(err error, conn net.Conn) {
	if err != nil {
		fmt.Fprintf(conn, "error: Server failure%s\n", END_TAG)
		panic(err)
	}
}
