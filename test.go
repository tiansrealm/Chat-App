package main

import "fmt"

func main() {
	array := []int{1, 2, 3, 4, 5}

	for i := 0; i < 3; i++ {
		for _, x := range array {
			if x == 3 {
				continue
			}
			fmt.Println(x)
		}
	}
}
