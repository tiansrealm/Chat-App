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
	"sync"
)

//--------------CONSTANTS---------------------------------
const USERLIST_FILENAME = "user_list.txt"

//used to sperate user messages
const USER_MESSAGE_SEPERATOR = "<end of message>"

//protocol
const END_TAG = "<end>"
const DOES_USER_EXIST = "does user exist"
const CHECK_PASS = "check password"
const ADD_USER = "add user"
const DELETE_USER = "delete user"
const ADD_MESSAGE = "add message"
const READ_MESSAGES = "read messages"

/*types of queries
DOES_USER_EXIST: uname  END_TAG
CHECK_PASS: uname: psw  END_TAG
ADD_USER: uname: psw    END_TAG
DELETE_USER: uname: psw END_TAG
READ_MESSAGES: uname: num messages END_TAG
ADD_MESSAGE: uname: psw : message  END_TAG
*/

//--------------DATA---------------------------

//Map of all users
//key is username, value is password
var user_map = make(map[string]string)

//A lock for the above user map struct
var user_map_lock = sync.Mutex{}

//Map of locks for locking files read/write operations
//key is filename, value is Mutex
//one unique lock per file
var lock_for_files_map = make(map[string]*sync.Mutex)

//============================Primary Functions======================================

/*main function
initialize server socket and listens for queries.
Upon accepting a new client, server will create new go rountine to handle
that client concurrently
*/
func main() {
	//load user list for faster access to a list of current users
	load_user_list()
	in, _ := net.Listen("tcp", ":8083")
	defer in.Close()

	client_id := 0

	for { //loop accepting clients to accept many clients
		conn, _ := in.Accept()
		fmt.Fprintf(os.Stderr, "Got a connection for Client %d\n", client_id)
		go handle_client(conn, client_id) //new go routine to handle new client
		client_id++
	}

}

/*scans query messages from the cient connection and
calls evalute() to perform appropriate actions
*/
func handle_client(conn net.Conn, client_id int) {
	scanner := bufio.NewScanner(conn)
	query := ""
	for scanner.Scan() {
		newLine := scanner.Text()
		index_of_endtag := strings.Index(newLine, END_TAG)
		if index_of_endtag != -1 {
			//reached end
			query += newLine[0 : len(newLine)-len(END_TAG)] //append without end tag
			//fmt.Println("Got query: " + query)
			evaluate(query, conn)
			query = "" //reset query
		} else {
			query += newLine //append to read query
		}
	}
	//interupted
	fmt.Fprintf(os.Stderr, "Connection gone from Client %d!\n", client_id)
	conn.Close()
}

/*
	handles queries for
		checking if
		user exist,
		adding new users,
		write/read message
	separate string into args by spliting along delimiter ":"

	This function should not access shared data, so it should not be using locks.
	This function directs to other function that will lock files for read/write.
Pre-Condition:
	Request will be in the form of "queryname: arg1: arg2: ..."
Post Condition:
	Either respond to query, or direct to another function that will respond
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

	//check if query function is valid
	valid_queries := []string{DOES_USER_EXIST, CHECK_PASS, ADD_USER, DELETE_USER, ADD_MESSAGE, READ_MESSAGES}
	is_valid_query := false
	for _, query := range valid_queries {
		if query_function == query {
			is_valid_query = true
		}
	}
	if !is_valid_query {
		//not a  valid queries
		fmt.Fprintf(conn, "error: %s is not a valid query.%s\n", parsed_query[0], END_TAG)
		return
	}

	//for all queries, args should start with query, username
	//all queries have >= 2 args
	if !check_args(parsed_query, 2, conn) {
		//check args failed
		return //skip this query
	}
	uname := parsed_query[1]

	//check the only query with 2 arg, does user exist, else check for >= 3 args
	if query_function == DOES_USER_EXIST {
		does_user_exist(uname, conn)
		return
	} else {
		// check for more args
		if !check_args(parsed_query, 3, conn) {
			//check args failed
			return //skip this query
		}
	}

	//------following requires >=3 args; passed checked args 3 above
	if query_function == ADD_USER {
		//doesn't need password authentication
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
		//already passed when called authenticate
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
}

//==========================================================================================
//Functions that respond to queries, used by Evalute
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
also returns false if not match
Locks user map for reading*/
func authenticate(uname string, psw string, conn net.Conn) bool {
	user_map_lock.Lock()
	defer user_map_lock.Unlock()
	if _, is_exist := user_map[uname]; is_exist && user_map[uname] == psw {
		return true
	} else {
		fmt.Fprintf(conn, "error: Username and Password combination not found. %s\n", END_TAG)
		return false
	}
}

/*simple check if username from query is a existing user. Does not check password.
respond sucess if user exists, else respond error
Locks usermap*/
func does_user_exist(uname string, conn net.Conn) {
	user_map_lock.Lock()
	defer user_map_lock.Unlock()
	if _, is_exist := user_map[uname]; is_exist {
		fmt.Fprintf(conn, "success: user exists %s\n", END_TAG)
	} else {
		fmt.Fprintf(conn, "error: no such user %s\n", END_TAG)
	}
}

/*Create new user and write new user info to user list file
send error response if user already exists
Locks user map. May lock Userlist file and user message file*/
func add_user(uname string, psw string, conn net.Conn) {
	user_map_lock.Lock()
	defer user_map_lock.Unlock()
	_, is_exist := user_map[uname]
	if !is_exist {
		//create user if not exist
		user_map[uname] = psw
		//open user list file to write to end of it
		create_and_lock(USERLIST_FILENAME) // lock userlist file for editing
		defer lock_for_files_map[USERLIST_FILENAME].Unlock()
		file, open_err := os.OpenFile(USERLIST_FILENAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		defer file.Close()
		check_err_and_repond(open_err, conn)

		text := uname + " : " + psw + "\r\n"
		if _, write_err := file.WriteString(text); write_err != nil {
			check_err_and_repond(write_err, conn)
		}
		//create user data file
		u_file_name := uname + ".txt"
		create_and_lock(u_file_name) // lock user file for deleting and recreating
		defer lock_for_files_map[u_file_name].Unlock()
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
writing to database file containing stored messsages the user
Locks message file of user*/
func add_message(uname string, new_message string, conn net.Conn) {
	filename := uname + ".txt"
	create_and_lock(filename) // lock user message file for editing
	defer lock_for_files_map[filename].Unlock()

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
locks usermap and message file of user that is being deleted*/
func delete_user(uname string, conn net.Conn) {
	//delete user from server memory
	user_map_lock.Lock()
	delete(user_map, uname)
	user_map_lock.Unlock()
	err := rewrite_userlist() //delete user from user list file
	check_err_and_repond(err, conn)

	//delete user message file
	filename := uname + ".txt"
	create_and_lock(filename) // lock the file we want to delete
	defer lock_for_files_map[filename].Unlock()
	os.Remove(filename)
	//repond sucess
	fmt.Fprintf(conn, "success: Deleted user %s.%s\n", uname, END_TAG)
}

/*reads messages from user file database
locks message file of user*/
func read_messages(uname string, num_messages int, conn net.Conn) {
	filename := uname + ".txt"
	create_and_lock(filename) // lock user message file for reading
	defer lock_for_files_map[filename].Unlock()

	message_file, open_err := os.OpenFile(filename, os.O_CREATE, 0600) //create file if not exist
	defer message_file.Close()
	check_err_and_repond(open_err, conn)

	messages_in_byte, read_err := ioutil.ReadFile(filename)
	check_err_and_repond(read_err, conn)

	messages_in_string := string(messages_in_byte)

	message_array := strings.SplitAfter(messages_in_string, USER_MESSAGE_SEPERATOR)
	message_array = message_array[0 : len(message_array)-1] //last index is empty cause of how splitafter works
	recent_messages := message_array
	if num_messages < len(message_array) {
		//only show recent num messages if there exist more than that
		recent_messages = message_array[len(message_array)-num_messages:]
	}
	//fmt.Fprintf(os.Stderr, "printing message%s\n", recent_messages)
	//give back most recent num_messages of messages
	response := ""
	for _, message := range recent_messages {
		response += message + "\n"
	}
	fmt.Fprintf(conn, "success: %s%s\n", response, END_TAG)
}

//-----------------------userlist operations-----------------------------------------------

/*loads list of existing users from file database into server memory
for faster checks that user exist
Locks userlist file and usermap in memory*/
func load_user_list() {
	create_and_lock(USERLIST_FILENAME) // lock userlist file for reading
	defer lock_for_files_map[USERLIST_FILENAME].Unlock()

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
		user_map_lock.Lock()
		user_map[uname] = psw
		user_map_lock.Unlock()
	}

}

/*rewrite list of existing user from server memory to userlist file. Needed after removing a user
Locks userlist file and usermap in memory*/
func rewrite_userlist() error {
	create_and_lock(USERLIST_FILENAME) // lock userlist file for editing
	defer lock_for_files_map[USERLIST_FILENAME].Unlock()

	os.Remove(USERLIST_FILENAME) //delete old file
	//rewrite new user list file
	file, err := os.OpenFile(USERLIST_FILENAME, os.O_CREATE|os.O_WRONLY, 0600)
	defer file.Close()
	if err != nil {
		return err
	}
	//
	user_map_lock.Lock() //locks user map for reading
	defer user_map_lock.Unlock()
	for uname, psw := range user_map {
		text := uname + " : " + psw + "\r\n"
		if _, write_err := file.WriteString(text); write_err != nil {
			return write_err
		}
	}
	return nil //no errors = success
}

//--------------common functions--------------------------------------------------

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

/*locking function for files. Called when want to lock a file.
creates lock and store into map under filename key if lock doesn't exist
lastly, calls lock on lock associated with filename*/
func create_and_lock(filename string) {
	_, ok := lock_for_files_map[filename]
	if !ok {
		//lock doesn't exist, so create one
		lock_for_files_map[filename] = &sync.Mutex{}
	}
	lock_for_files_map[filename].Lock() //attempt to acquire lock no matter the case
}
