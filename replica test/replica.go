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

//used to denote that file is used for holding subscriptions
const SUBSCRIPTION_FILE_TAG = "_sublist.txt"

//protocol
const END_TAG = "<end>"
const DOES_USER_EXIST = "does user exist"
const CHECK_PASS = "check password"
const ADD_USER = "add user"
const DELETE_USER = "delete user"
const ADD_MESSAGE = "add message"
const READ_MESSAGES = "read messages"
const SUB = "subscribe"
const UNSUB = "unsubscribe"
const SUB_FEED = "get sub feed"

/*types of queries
DOES_USER_EXIST: uname  END_TAG
CHECK_PASS: uname: psw  END_TAG
ADD_USER: uname: psw    END_TAG
DELETE_USER: uname: psw END_TAG
READ_MESSAGES: uname: num messages END_TAG
ADD_MESSAGE: uname: psw : message  END_TAG
SUB: uname: psw: sub_uname END_TAG
UNSUB: uname: psw: sub_uname END_TAG
SUB_FED: uname: psw: num_messages END_TAG
*/

//error response
const WRONG_NUM_ARGS = "error: Not valid # of args" + END_TAG + "\n"

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

var query_history = []string{}

//A lock for the above ] query_history array
var query_history_lock = sync.Mutex{}

//============================Primary Functions======================================

/*main function
initialize server socket and listens for queries.
Upon accepting a new client, server will create new go rountine to handle
that client concurrently
*/
func main() {
	//establish connection to the primary replica
	//connect to server
	conn_main_replica, err := net.Dial("tcp", "localhost:8084")
	defer conn_main_replica.Close()
	if err != nil {
		panic("Failed connect to conn_main_replica\n")
	}

	//load user list for faster access to a list of current users
	load_user_list()
	handle_requests(conn_main_replica)
}

/*scans query messages from the cient connection and
calls evalute() to perform appropriate actions
write the response returned by evalute to the conn
*/
func handle_requests(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	query := ""
	for scanner.Scan() {
		fmt.Println("scanning")
		newLine := scanner.Text()
		fmt.Println(newLine)
		index_of_endtag := strings.Index(newLine, END_TAG)
		if index_of_endtag != -1 {
			//reached end of query
			fmt.Println("Got query: " + query)
			query += newLine[0 : len(newLine)-len(END_TAG)] //append without end tag
			_, is_updated := evaluate(query)
			if is_updated {
				fmt.Fprint(conn, "sucess\n")
			} else {
				panic("suppose to recieve query that does updates")
			}
			query = "" //reset query

		} else {
			query += newLine //append to read query
		}
	}
	conn.Close()
}

//below is identical code to main replica---------------------------------------

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
	complete request and return the response, along with bool saying if there was an update
	returns (response, is_updated)
	response with be in the form of  "success/error: response"
*/
func evaluate(query string) (string, bool) {
	delimiter := ":"
	parsed_query := strings.Split(query, delimiter)
	//trims white space at the ends
	for i := 0; i < len(parsed_query); i++ {
		parsed_query[i] = strings.TrimSpace(parsed_query[i])
	}

	query_function := parsed_query[0]

	//check if query function is valid
	valid_queries := []string{DOES_USER_EXIST, CHECK_PASS, ADD_USER, DELETE_USER,
		ADD_MESSAGE, READ_MESSAGES, SUB, UNSUB, SUB_FEED}
	is_valid_query := false
	for _, query := range valid_queries {
		if query_function == query {
			is_valid_query = true
		}
	}
	if !is_valid_query {
		//not a  valid queries
		return fmt.Sprintf("error: %s is not a valid query.%s\n", parsed_query[0], END_TAG), false
	}

	//for all queries, args should start with query, username
	//all queries have >= 2 args
	if !check_args(parsed_query, 2) {
		//check args failed
		return WRONG_NUM_ARGS, false
	}
	uname := parsed_query[1]

	//check the only query with 2 arg, does user exist, else check for >= 3 args
	if query_function == DOES_USER_EXIST {
		return does_user_exist(uname)
	} else {
		// check for more args
		if !check_args(parsed_query, 3) {
			//check args failed
			return WRONG_NUM_ARGS, false
		}
	}

	//------following requires >=3 args; passed checked args 3 above
	if query_function == ADD_USER {
		//doesn't need password authentication
		return add_user(uname, parsed_query[2])
	} else if query_function == READ_MESSAGES {
		//args should be query, username, num_message

		if num_message, convert_err := strconv.Atoi(parsed_query[2]); convert_err != nil {
			return fmt.Sprintf("error: third arg must be integer.%s\n", END_TAG), false
		} else {
			return read_messages(uname, num_message)
		}
	}

	psw := parsed_query[2]

	//following functions needs password authentication
	if !authenticate(uname, psw) {
		//uname and psw don't match
		response := fmt.Sprintf("error: Username and Password combination not found. %s\n", END_TAG)
		return response, false
	}

	switch query_function {
	case CHECK_PASS:
		//reply passed username + password check
		//already passed when called authenticate
		return fmt.Sprintf("success: correct username and password %s\n", END_TAG), false
	case DELETE_USER:
		return delete_user(uname)
	case ADD_MESSAGE:
		//args should be query, username, password, message
		if !check_args(parsed_query, 4) {
			//check args failed
			return WRONG_NUM_ARGS, false
		}
		message := parsed_query[3]
		return add_message(uname, message)
	case SUB:
		if !check_args(parsed_query, 4) {
			//check args failed
			return WRONG_NUM_ARGS, false
		}
		sub_uname := parsed_query[3]
		return subscribe(uname, sub_uname)
	case UNSUB:
		if !check_args(parsed_query, 4) {
			//check args failed
			return WRONG_NUM_ARGS, false
		}
		sub_uname := parsed_query[3]
		return unsubscribe(uname, sub_uname)
	case SUB_FEED:
		if num_messages, convert_err := strconv.Atoi(parsed_query[3]); convert_err != nil {
			return fmt.Sprintf("error: fourth arg must be integer.%s\n", END_TAG), false
		} else {
			return sub_feed(uname, num_messages)
		}
	}
	return fmt.Sprintf("error: unknown error.%s\n", END_TAG), false
}

//==========================================================================================
//Functions that respond to queries, used by Evalute
//============================================================================================

/*checks if num args from query is AT LEAST the num expected
return false if args is wrong, true otherwise*/
func check_args(parsed_query []string, num_expected int) bool {
	return (len(parsed_query) >= num_expected)
}

/*  checks password against username
returns false if not match
Locks user map for reading*/
func authenticate(uname string, psw string) bool {
	user_map_lock.Lock()
	defer user_map_lock.Unlock()
	if _, is_exist := user_map[uname]; is_exist && user_map[uname] == psw {
		return true
	} else {
		return false
	}
}

/*simple check if username from query is a existing user. Does not check password.
respond sucess if user exists, else respond error
Locks usermap*/
func does_user_exist(uname string) (string, bool) {
	user_map_lock.Lock()
	defer user_map_lock.Unlock()
	if _, is_exist := user_map[uname]; is_exist {
		return fmt.Sprintf("success: user exists %s\n", END_TAG), false
	} else {
		return fmt.Sprintf("error: no such user %s\n", END_TAG), false
	}
}

/*Create new user and write new user info to user list file
send error response if user already exists
Locks user map. May lock Userlist file and user message file*/
func add_user(uname string, psw string) (string, bool) {
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
		if open_err != nil {
			return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
		}

		text := uname + " : " + psw + "\r\n"
		if _, write_err := file.WriteString(text); write_err != nil {
			return fmt.Sprintf("error: Server write file error%s\n", END_TAG), false
		}
		//create user data file
		u_file_name := uname + ".txt"
		create_and_lock(u_file_name) // lock user file for deleting and recreating
		defer lock_for_files_map[u_file_name].Unlock()
		os.Remove(u_file_name) // clear old junk
		created_file, create_err := os.Create(u_file_name)
		defer created_file.Close()
		if create_err != nil {
			return fmt.Sprintf("error: Server create error%s\n", END_TAG), false
		} else {
			//response
			return fmt.Sprintf("success: I added user %s.%s\n", uname, END_TAG), true
		}
	} else {
		//negative response
		return fmt.Sprintf("error: user, %s, already exists.%s\n", uname, END_TAG), false
	}
}

/*Add a new message under the user with given uname, by
writing to database file containing stored messsages the user
Locks message file of user*/
func add_message(uname string, new_message string) (string, bool) {
	filename := uname + ".txt"
	create_and_lock(filename) // lock user message file for editing
	defer lock_for_files_map[filename].Unlock()

	message_file, open_err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	defer message_file.Close()
	if open_err != nil {
		return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
	}

	//write new message to file
	newline := "\r\n"
	text_to_write := new_message + newline + USER_MESSAGE_SEPERATOR + newline
	if _, write_err := message_file.WriteString(text_to_write); write_err != nil {
		return fmt.Sprintf("error: server failed to write.%s\n", END_TAG), false
	} else {
		return fmt.Sprintf("success: added message for %s.%s\n", uname, END_TAG), true
	}

}

/*deletes user from userlist file and delete message file asscioated with that user
locks usermap and message file of user that is being deleted*/
func delete_user(uname string) (string, bool) {
	//delete user from server memory
	user_map_lock.Lock()
	delete(user_map, uname)
	user_map_lock.Unlock()
	err := rewrite_userlist() //delete user from user list file
	if err != nil {
		return fmt.Sprintf("error: Server rewrite uselist error%s\n", END_TAG), false
	}

	//delete user message file
	filename := uname + ".txt"
	create_and_lock(filename) // lock the file we want to delete
	defer lock_for_files_map[filename].Unlock()
	os.Remove(filename)
	//repond sucess
	return fmt.Sprintf("success: Deleted user %s.%s\n", uname, END_TAG), true
}

/*reads messages from user file database
locks message file of user*/
func read_messages(uname string, num_messages int) (string, bool) {
	filename := uname + ".txt"
	create_and_lock(filename) // lock user message file
	defer lock_for_files_map[filename].Unlock()

	message_file, open_err := os.OpenFile(filename, os.O_CREATE, 0600) //create file if not exist
	defer message_file.Close()
	if open_err != nil {
		return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
	}

	messages_in_byte, read_err := ioutil.ReadFile(filename)
	if read_err != nil {
		return fmt.Sprintf("error: Server read error%s\n", END_TAG), false
	}

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
	return fmt.Sprintf("success: %s%s\n", response, END_TAG), false
}
func sub_feed(uname string, num_messages int) (string, bool) {
	sub_filename := uname + SUBSCRIPTION_FILE_TAG
	create_and_lock(sub_filename)
	defer lock_for_files_map[sub_filename].Unlock()

	//open sublist file and store subscribed names to sublist
	sublist_file, open_err := os.OpenFile(sub_filename, os.O_CREATE|os.O_RDONLY, 0600)
	defer sublist_file.Close()
	if open_err != nil {
		return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
	}

	sublist := []string{}
	scanner := bufio.NewScanner(sublist_file)
	for scanner.Scan() {
		scanned_name := scanner.Text()
		sublist = append(sublist, scanned_name)
	}
	response := ""
	//append num_messages frome each sub_uname
	for _, sub_uname := range sublist {
		message_filename := sub_uname + ".txt"
		create_and_lock(message_filename) // lock user message file
		defer lock_for_files_map[message_filename].Unlock()

		message_file, open_err := os.OpenFile(message_filename, os.O_CREATE, 0600) //create file if not exist
		defer message_file.Close()
		if open_err != nil {
			return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
		}

		messages_in_byte, read_err := ioutil.ReadFile(message_filename)
		if read_err != nil {
			return fmt.Sprintf("error: Server read error%s\n", END_TAG), false
		}

		messages_in_string := string(messages_in_byte)

		message_array := strings.SplitAfter(messages_in_string, USER_MESSAGE_SEPERATOR)
		message_array = message_array[0 : len(message_array)-1] //last index is empty cause of how splitafter works
		recent_messages := message_array
		if num_messages < len(message_array) {
			//only show recent num messages if there exist more than that
			recent_messages = message_array[len(message_array)-num_messages:]
		}
		response += "<br /><b>Recent messages from - " + sub_uname + "</b><br />"
		for _, message := range recent_messages {
			response += message + "\n"
		}

	}
	return fmt.Sprintf("success: %s%s\n", response, END_TAG), false

}

/*subscribes user to another user by writing sub_uname, the subscribe target,
in the main user's sublist.txt*/
func subscribe(uname string, sub_uname string) (string, bool) {
	filename := uname + SUBSCRIPTION_FILE_TAG
	create_and_lock(filename)
	defer lock_for_files_map[filename].Unlock()

	//open sublist file
	sublist_file, open_err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0600)
	defer sublist_file.Close()
	if open_err != nil {
		return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
	}

	//scan file to see if subscription exists
	scanner := bufio.NewScanner(sublist_file)
	for scanner.Scan() {
		scanned_name := scanner.Text()
		//check if already subscribed
		if scanned_name == sub_uname {
			//already subscribed so do nothing
			return fmt.Sprintf("success: Already subscribed.%s\n", END_TAG), false
		}
	}

	//subscription don't exist, so add subscription
	text := sub_uname + "\n"
	if _, write_err := sublist_file.WriteString(text); write_err != nil {
		return fmt.Sprintf("error: Server write error%s\n", END_TAG), false
	} else {
		return fmt.Sprintf("success: Added subscription.%s\n", END_TAG), true
	}
}

/*un-subscribes user to another user by deleting sub_uname, the subscribe target,
from the main user's sublist.txt*/
func unsubscribe(uname string, sub_uname string) (string, bool) {
	filename := uname + SUBSCRIPTION_FILE_TAG
	create_and_lock(filename)
	defer lock_for_files_map[filename].Unlock()

	//open sublist file
	sublist_file, open_err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0600)
	if open_err != nil {
		return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
	}

	sublist := []string{}
	removed := false
	//scan file to see if subscription exists
	scanner := bufio.NewScanner(sublist_file)
	for scanner.Scan() {
		scanned_name := scanner.Text()

		if scanned_name == sub_uname {
			removed = true //didn't add scanned_name to sublist
			continue
		} else {
			sublist = append(sublist, scanned_name)
		}
	}
	sublist_file.Close()
	//rewrite file if removed sub_uname
	if removed {
		os.Remove(filename)
		new_sublist_file, open_err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0600)
		defer new_sublist_file.Close()
		if open_err != nil {
			return fmt.Sprintf("error: Server open error%s\n", END_TAG), false
		}

		for _, uname := range sublist {
			text := uname + "\n"
			if _, write_err := new_sublist_file.WriteString(text); write_err != nil {
				return fmt.Sprintf("error: Server write error%s\n", END_TAG), false
			}
		}

	}
	return fmt.Sprintf("success: removed subscription.%s\n", END_TAG), removed
}

//-----------------------userlist operations-----------------------------------------------

/*loads list of existing users from file database into server memory
for faster checks that user exist
Locks userlist file and usermap in memory*/
func load_user_list() {
	create_and_lock(USERLIST_FILENAME) // lock userlist file
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
