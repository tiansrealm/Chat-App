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
const MESSAGE_ENDTAG = "<end of message>"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//===================================================================
func main() {
	/*main function initialize server socket and listens for queries*/
	load_user_list()
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
func evaluate(query string, conn net.Conn) {
	/*
		handles queries
		separate string into args by spliting along delimiter ":"
		Pre-Condition:
		Request will be in the form of "queryname: arg1: arg2: ..."
		Post Condition:
		response with be in the form of  "success/Error: value"
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
		if len(parsed_query) != 2 {
			fmt.Fprintf(conn, "Error: Not valid # of args\n")
			return
		}
		_, is_exist := user_map[parsed_query[1]]
		if is_exist {
			fmt.Fprintf(conn, "success: true\n")
		} else {
			fmt.Fprintf(conn, "success: false\n")
		}
	case "add user":
		//args should be query, username, password
		if len(parsed_query) != 3 {
			fmt.Fprintf(conn, "Error: Not valid # of args\n")
			return
		}
		add_user(parsed_query[1], parsed_query[2])
		fmt.Fprintf(conn, "success: I added user %s\n", parsed_query[1])
	case "add message":
		//args should be query, username, message
		if len(parsed_query) != 3 {
			fmt.Fprintf(conn, "Error: Not valid # of args\n")
			return
		}
		add_message(parsed_query[1], parsed_query[2])
		fmt.Fprintf(conn, "success: added message for%s\n", parsed_query[1])
	default:
		fmt.Fprintf(conn, "Error:%s is not a valid query\n", parsed_query[0])

	}

}

//==========================================================================================
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

func add_user(uname string, psw string) {
	/*Create new user and write new user info to user list file*/

	_, is_exist := user_map[uname]
	if !is_exist {
		//create user if not exist
		user_map[uname] = &User{uname, psw, []string{}}
	}
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
}

func add_message(uname string, new_message string) {
	/*
		Add a new message under the user with given uname, by
		writing to database file containing stored messsages the user
	*/
	user := user_map[uname]
	filename := uname + ".txt"
	message_file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	check(err)
	defer message_file.Close()
	// if user.messages == nil {
	// 	//load messages for this user since haven't already done so
	// 	reader := bufio.NewReader(file)
	// 	//read until the delimiter for end of a message
	// 	for stored_message, err := reader.ReadString(MESSAGE_ENDTAG) {
	// 		if err == io.EOF{
	// 			break
	// 		}else {check(err)}
	// 		stored_message := stored_message[0:len(stored_message)-len(MESSAGE_ENDTAG)]
	// 		// store user message in server memory
	// 		user.messages = append(user.messages, stored_message)
	// 		user_map[user.uname].messages = user.messages

	// 	}
	// }

	//write new message to file
	text_to_write := new_message + MESSAGE_ENDTAG
	if _, err = file.WriteString(text_to_write); err != nil {
		panic(err)
	}

}

func read_messages(uname string, num_messages int) {
	/*reads messages from user file database*/
	check(err)
	defer message_file.Close()

	messages_in_byte, err := ioutil.ReadFile(filename)
	if err != nil {
		Panic(err)
	}
	messages_in_string := string(messages_in_byte)

	message_array := strings.Split(messages_in_string, MESSAGE_ENDTAG)

	recent_messages := message_array[len(message_array)-num_messages:]

	//give back most recent num_messages of messages
}
