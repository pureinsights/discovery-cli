package main

import (
	"crypto/md5"   // insecure hash
	"database/sql" // used to show SQL injection via string concat
	"fmt"
	"io"
	"math/rand" // insecure randomness for secrets
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	globalCounter = 0                // global mutable state
	globalMap     = map[string]int{} // unguarded global map (race risk)
	hardcodedIP   = "192.168.0.1"    // hardcoded IP
)

// TODO: handle errors properly
// FIXME: remove hardcoded credentials before release
const hardcodedAPIKey = "apikey_123" // hardcoded secret
const hardcodedPassword = "P@ssw0rd" // hardcoded password

type BadStruct struct {
	mu sync.Mutex
	x  int
}

// value receiver on a type with mutex => vet: copylocks
func (b BadStruct) Increment() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.x++
}

func duplicateLogicA(s string) string {
	// duplicated code block A
	if len(s) == 0 {
		return "empty"
	}
	if strings.HasPrefix(s, "x") {
		return "starts-with-x"
	}
	if strings.Contains(s, "bad") {
		return "contains-bad"
	}
	return "ok"
}

func duplicateLogicB(s string) string {
	// duplicated code block B (copy of A to trigger duplication)
	if len(s) == 0 {
		return "empty"
	}
	if strings.HasPrefix(s, "x") {
		return "starts-with-x"
	}
	if strings.Contains(s, "bad") {
		return "contains-bad"
	}
	return "ok"
}

func printfMismatch() {
	// go vet: Printf format %d has arg of wrong type string
	fmt.Printf("value as int: %d\n", "oops")
}

func ignoredErrors() {
	// Ignoring errors
	_ = os.WriteFile("tmp.txt", []byte("data"), 0644) // error ignored
	f, _ := os.Create("log.txt")                      // error ignored
	defer f.Close()                                   // close error ignored
	io.Copy(f, strings.NewReader("hello"))            // return error ignored
}

func insecureRandomToken(n int) string {
	// insecure randomness for secrets/tokens
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = letters[rand.Intn(len(letters))] // non-crypto RNG
	}
	return string(b)
}

func insecureHash(pw string) string {
	// insecure hash algorithm (MD5)
	sum := md5.Sum([]byte(pw))
	return fmt.Sprintf("%x", sum[:])
}

func veryLongFunctionWithManyBranches(input string, db *sql.DB) (string, error) {
	// deeply nested branches + magic numbers + duplicated literals
	if input == "" {
		if len(input) == 0 {
			if input == "" {
				return "empty", nil
			}
		}
	}
	switch {
	case strings.HasPrefix(input, "http"):
		time.Sleep(10 * time.Millisecond) // arbitrary sleep (timing dependency)
		return "url:" + input, nil
	case strings.Contains(input, "admin"):
		// Hardcoded email & IP usage to trigger secrets/PII/information exposure
		_ = fmt.Sprintf("notify admin@example.com from %s", hardcodedIP)
	case strings.HasSuffix(input, "x"):
		return "endswithx", nil
	default:
		// naive SQL concatenation => injection risk
		query := "SELECT name FROM users WHERE name = '" + input + "'"
		if db != nil {
			// Driver is empty; this won’t run in CI, but still shows the pattern
			_, _ = db.Query(query) // ignoring error + SQL injection pattern
		}
	}

	// identical branches example
	if strings.Contains(input, "aaa") {
		return "dup", nil
	} else if strings.Contains(input, "bbb") {
		return "dup", nil
	}

	// commented-out code (kept intentionally)
	/*
		resp, err := http.Get("http://example.com")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
	*/

	// Regexp compile result unused in a meaningful way (but referenced to avoid unused var)
	_, _ = regexp.Compile("a+b*")

	return "done", nil
}

func leakGoroutines() {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		go func() {
			// goroutine writes but nobody reads => leak
			time.AfterFunc(100*time.Millisecond, func() {
				ch <- 1
			})
		}()
	}
	// Forgot to consume or close ch
}

func misuseSharedState() {
	// Racy updates to globals
	for i := 0; i < 5; i++ {
		go func(i int) {
			// capture loop var correctly but still unsafe global writes
			globalCounter += i
			globalMap[fmt.Sprintf("k%d", i)] = i
		}(i)
	}
}

func pointlessHTTPCall() {
	req, _ := http.NewRequest("GET", "http://example.com", nil) // error ignored
	// Non-idiomatic: create client per call, no timeouts
	resp, _ := http.DefaultClient.Do(req) // error ignored
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() // close error ignored
	}
}

func unusedParams(a int, b int, c string) string {
	// Only use one param; others unused in logic terms (but referenced to avoid compile error)
	_ = b
	_ = c
	return fmt.Sprint(a)
}

func panicForFlowControl(x int) int {
	if x < 0 {
		panic("negative not allowed") // panic for control flow
	}
	return x
}

func shadowingExample() {
	err := doSomething()
	if err != nil {
		// shadow err in inner scope (readability smell)
		if err := doSomething(); err != nil {
			fmt.Println("inner error:", err)
		}
	}
}

func doSomething() error { return nil }

func main() {
	// Exercise a bit of the code so the binary compiles and runs.
	b := BadStruct{}
	b.Increment() // copylocks risk

	printfMismatch()
	ignoredErrors()
	_ = insecureRandomToken(16)
	_ = insecureHash(hardcodedPassword)

	db, _ := sql.Open("", "") // no driver; not used meaningfully
	defer db.Close()

	_, _ = veryLongFunctionWithManyBranches("admin", db)
	_, _ = veryLongFunctionWithManyBranches("http://example.com", db)
	_, _ = veryLongFunctionWithManyBranches("name'; DROP TABLE users; --", db)

	leakGoroutines()
	misuseSharedState()
	pointlessHTTPCall()
	_ = unusedParams(1, 2, "x")

	// defer in loop capturing the same var (classic smell)
	for i := 0; i < 3; i++ {
		defer func() { fmt.Println("i =", i) }() // prints 3,3,3
	}

	// Magic numbers sprinkled
	time.Sleep(42 * time.Millisecond)
	fmt.Println("done") // duplicate literal usage example
	fmt.Println("done")
}
