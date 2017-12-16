package main

import "fmt"

var user_map map[int]string = make(map[int]string)

func main() {
	t := Test{"test"}
	t2 := &t
	fmt.Println(t2.name)
}

type Test struct {
	name string
}
