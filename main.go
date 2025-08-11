package main

import "fmt"

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b + 1
}

func main() {
	fmt.Println("Hello from Go!")
	fmt.Printf("2 + 3 = %d\n", Add(2, 3))
}
