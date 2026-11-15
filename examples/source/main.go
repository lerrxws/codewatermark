package main

import "fmt"

func main() {
	numbers := []int{2, 4, 6}
	total := 0

	for _, number := range numbers {
		total += number
	}

	fmt.Println("Total:", total)
}
