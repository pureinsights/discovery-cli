package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFizzBuzz_Table(t *testing.T) {
	cases := map[int]string{
		1:  "1",
		2:  "2",
		3:  "Fizz",
		5:  "Buzz",
		6:  "Fizz",
		10: "Buzz",
		15: "FizzBuzz",
	}
	for n, want := range cases {
		if got := FizzBuzz(n); got != want {
			t.Fatalf("FizzBuzz(%d) = %q; want %q", n, got, want)
		}
	}
}

func TestRun_UsageWhenNoArgs(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); !strings.Contains(got, "usage:") {
		t.Fatalf("expected usage message, got %q", got)
	}
}

func TestRun_PrintsSequence(t *testing.T) {
	var buf bytes.Buffer
	if err := run(&buf, []string{"5"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantLines := []string{"1", "2", "Fizz", "4", "Buzz"}
	got := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(got) != len(wantLines) {
		t.Fatalf("unexpected lines: got %d, want %d", len(got), len(wantLines))
	}
	for i := range wantLines {
		if got[i] != wantLines[i] {
			t.Fatalf("line %d = %q; want %q", i, got[i], wantLines[i])
		}
	}
}

func TestRun_InvalidArg(t *testing.T) {
	var buf bytes.Buffer
	err := run(&buf, []string{"not-a-number"})
	if err == nil || !strings.Contains(err.Error(), "invalid number") {
		t.Fatalf("expected invalid number error, got %v", err)
	}
}

func TestMainFunction_CoversMain(t *testing.T) {
	// Capture stdout to verify main() calls run correctly.
	origArgs, origStdout := os.Args, os.Stdout
	defer func() { os.Args, os.Stdout = origArgs, origStdout }()

	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Args = []string{"cli", "3"}

	main()

	_ = w.Close()
	outBytes, _ := io.ReadAll(r)
	out := string(outBytes)

	// Build expected output dynamically to avoid hardcoding
	var b bytes.Buffer
	_ = run(&b, []string{"3"})
	expected := b.String()

	if out != expected {
		t.Fatalf("main() output mismatch.\nGot:\n%s\nWant:\n%s", out, expected)
	}
}
