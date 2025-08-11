package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// FizzBuzz returns "Fizz", "Buzz", "FizzBuzz", or the number.
func FizzBuzz(n int) string {
	switch {
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Buzz"
	case n%5 == 0:
		return "Buzz"
	default:
		return strconv.Itoa(n)
	}
}

// run prints FizzBuzz from 1..N based on args[0] and writes to out.
func run(out io.Writer, args []string) error {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(out, "usage: provide a positive integer, e.g. `cli 5`")
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(args[0]))
	if err != nil || n < 1 {
		return fmt.Errorf("invalid number: %q", args[0])
	}
	for i := 1; i <= n; i++ {
		_, _ = fmt.Fprintln(out, FizzBuzz(i))
	}
	return nil
}

func main() {
	// Keep main side-effect free (no os.Exit) so tests can call it.
	_ = run(os.Stdout, os.Args[1:])
}
