package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const END_TAG = "<end>"

func main() {
	conn, _ := net.Dial("tcp", "localhost:8083")
	defer conn.Close()
	scanner := bufio.NewScanner(os.Stdin)
	scannerConn := bufio.NewScanner(conn)
	for {
		query := ""
		for scanner.Scan() {
			newLine := scanner.Text()
			query += newLine //append to read query
			index_of_endtag := strings.Index(newLine, END_TAG)
			if index_of_endtag != -1 {
				//end of query
				break
			}
		}
		//send query
		fmt.Fprintf(conn, query+"\n")

		response := ""
		for scannerConn.Scan() {
			newLine := scannerConn.Text()
			response += newLine + "\n" //append to read query
			index_of_endtag := strings.Index(newLine, END_TAG)
			if index_of_endtag != -1 {
				//reached end of response
				break
			}
		}
		response = strings.Replace(response, END_TAG, "", 1) //remove tag
		fmt.Print(response)
	}
}
