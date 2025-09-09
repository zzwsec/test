package main

import (
	"fmt"
)

func Concat(a, b string) string {
	return a + " " + b
}

func main() {
	result := Concat("Hello", "World")
	fmt.Println(result)
}
