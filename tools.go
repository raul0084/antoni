package main

import (
	"fmt"
	"os"
	"time"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

// --- Tool 1: log_rewrite ---
//
// Every ADK tool follows the same three-part shape:
//   1. An Args struct describing the tool's inputs (the LLM fills these in)
//   2. A Result struct describing what the tool hands back to the LLM
//   3. A plain Go function doing the actual work
//
// The jsonschema tags are NOT decoration — they're the description the LLM
// reads to decide what to pass in. Vague tags produce vague tool calls.

type logRewriteArgs struct {
	Original  string `json:"original" jsonschema:"The original message before rewriting."`
	Rewritten string `json:"rewritten" jsonschema:"The polished, business-appropriate version."`
}

type logRewriteResult struct {
	Status string `json:"status"`
}

// logRewrite is the actual implementation. Notice it takes a plain string
// logPath argument rather than reaching into a global variable — this is
// what makes it testable (see tools_test.go: we pass a temp file path in
// tests, and the real log path in production).
func logRewrite(logPath string, args logRewriteArgs) (logRewriteResult, error) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return logRewriteResult{}, fmt.Errorf("opening log file: %w", err)
	}
	defer f.Close() // runs when the function returns, success or error alike

	entry := fmt.Sprintf("[%s] ORIGINAL: %s | REWRITTEN: %s\n",
		time.Now().Format(time.RFC3339), args.Original, args.Rewritten)
	if _, err := f.WriteString(entry); err != nil {
		return logRewriteResult{}, fmt.Errorf("writing log entry: %w", err)
	}
	return logRewriteResult{Status: "logged"}, nil
}

// NewLogRewriteTool wraps logRewrite into the shape ADK expects. This is
// where we "capture" the configured log path via a closure, so the
// ADK-facing function signature stays exactly what functiontool.New wants
// (agent.Context, Args) -> (Result, error), while still using our config.
//
// Return type is tool.Tool (an interface), not *functiontool.Tool — the
// functiontool package doesn't export a concrete Tool type. functiontool.New
// returns something that satisfies the tool.Tool interface, and that
// interface is what llmagent.Config.Tools expects anyway, so this is also
// the more correct type to hand upward.
func NewLogRewriteTool(logPath string) (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "log_rewrite",
			Description: "Logs the original and rewritten message to a local file for record-keeping. Call this after every rewrite.",
		},
		func(ctx agent.Context, args logRewriteArgs) (logRewriteResult, error) {
			return logRewrite(logPath, args)
		},
	)
}

// --- Tool 2: check_already_polished ---
//
// A second example so the pattern is clear: a lightweight heuristic tool
// that lets the agent skip rewriting when the input already reads as
// professional. This is a good "second tool" to add because it changes
// agent behavior (via your Instruction) rather than just logging.

type checkPolishArgs struct {
	Message string `json:"message" jsonschema:"The message to check for existing professionalism."`
}

type checkPolishResult struct {
	AlreadyPolished bool   `json:"already_polished"`
	Reason          string `json:"reason"`
}

func checkAlreadyPolished(ctx agent.Context, args checkPolishArgs) (checkPolishResult, error) {
	// Deliberately simple heuristic — this is a placeholder you'd likely
	// replace with something smarter later (e.g. checking for excessive
	// slang, all-caps, missing punctuation).
	if len(args.Message) > 0 && args.Message[len(args.Message)-1] != '.' &&
		args.Message[len(args.Message)-1] != '?' && args.Message[len(args.Message)-1] != '!' {
		return checkPolishResult{AlreadyPolished: false, Reason: "missing terminal punctuation"}, nil
	}
	return checkPolishResult{AlreadyPolished: true, Reason: "no obvious issues detected"}, nil
}

func NewCheckPolishTool() (tool.Tool, error) {
	return functiontool.New(
		functiontool.Config{
			Name:        "check_already_polished",
			Description: "Checks whether a message already reads as professional and business-appropriate, to avoid unnecessary rewriting.",
		},
		checkAlreadyPolished,
	)
}
