package main

import (
	"bufio"
	"fmt"
	"net"
	//"os"
	"strconv"
	"strings"
)

const END_TAG = "<end>"

func main() {

	ch := make(chan int)
	num_test := 10
	// for i := 0; i < num_test; i++ {
	// 	os.Remove("test" + strconv.Itoa(num_test) + ".txt")
	// }
	for i := 0; i < num_test; i++ {
		go simulate_client(0, ch)
	}
	for i := 0; i < num_test; i++ {
		<-ch
	}
}

func simulate_client(id int, ch chan int) {
	conn, _ := net.Dial("tcp", "localhost:8083")
	defer conn.Close()
	//scanner := bufio.NewScanner(os.Stdin)
	//scannerConn := bufio.NewScanner(conn)
	query := "add message: tian: a: test abcde message from client #" + strconv.Itoa(id)

	fmt.Fprintf(conn, query+"<end>\n")

	scannerConn := bufio.NewScanner(conn)
	response := ""
	for scannerConn.Scan() {
		newLine := scannerConn.Text()
		response += newLine //append to read query
		index_of_endtag := strings.Index(newLine, END_TAG)
		if index_of_endtag != -1 {
			//reached end of response
			break
		}
	}
	//
	query2 := "read messages: tian: 10"

	fmt.Fprintf(conn, query2+"<end>\n")

	response2 := ""
	for scannerConn.Scan() {
		newLine := scannerConn.Text()
		response2 += newLine //append to read query
		index_of_endtag := strings.Index(newLine, END_TAG)
		if index_of_endtag != -1 {
			//reached end of response
			break
		}
	}
	ch <- 42
}
