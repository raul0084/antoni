package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogRewrite(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	args := logRewriteArgs{
		Original:  "hey can u send that thing",
		Rewritten: "Could you send that document over when you get a chance?",
	}

	result, err := logRewrite(logPath, args)
	if err != nil {
		t.Fatalf("logRewrite returned unexpected error: %v", err)
	}
	if result.Status != "logged" {
		t.Errorf("expected status 'logged', got %q", result.Status)
	}

	// Verify the file actually contains what we expect.
	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(contents), args.Original) {
		t.Errorf("log file missing original message; got: %s", contents)
	}
	if !strings.Contains(string(contents), args.Rewritten) {
		t.Errorf("log file missing rewritten message; got: %s", contents)
	}
}

func TestLogRewrite_InvalidPath(t *testing.T) {
	// A path that can't exist (a file used as a directory component)
	// should surface as an error, not a panic.
	badPath := filepath.Join(t.TempDir(), "no", "such", "dir", "test.log")

	_, err := logRewrite(badPath, logRewriteArgs{Original: "x", Rewritten: "y"})
	if err == nil {
		t.Error("expected an error for an unwritable path, got nil")
	}
}

// Table-driven test: the standard Go pattern for testing several inputs
// against the same logic without repeating the test body. Instead of one
// TestX function per case (which you might do in Python with parametrize),
// you define a slice of cases and loop over it.
func TestCheckAlreadyPolished(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    bool
	}{
		{name: "ends with period", message: "Please review the attached file.", want: true},
		{name: "ends with question mark", message: "Could you take a look?", want: true},
		{name: "missing punctuation", message: "send it over when you can", want: false},
		{name: "empty message", message: "", want: true}, // nothing to flag
	}

	for _, tc := range cases {
		// t.Run creates a named subtest, so failures report exactly which
		// case failed (e.g. "TestCheckAlreadyPolished/missing_punctuation")
		// instead of just "TestCheckAlreadyPolished".
		t.Run(tc.name, func(t *testing.T) {
			result, err := checkAlreadyPolished(nil, checkPolishArgs{Message: tc.message})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.AlreadyPolished != tc.want {
				t.Errorf("message %q: got AlreadyPolished=%v, want %v",
					tc.message, result.AlreadyPolished, tc.want)
			}
		})
	}
}
