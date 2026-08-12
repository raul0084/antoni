package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/raul0084/antoni/internal/agentcore"
)

func main() {
	ctx := context.Background()

	cfg, err := agentcore.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}

	// Construct our platform agnostic brain package
	antoniAgent, err := agentcore.NewAgent(ctx, &cfg)
	if err != nil {
		log.Fatalf("failed to spin up antoni: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\n Antoni >>> write the message you want to rewrite \n\n> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("\ngoodbye!")
				break
			}
			log.Fatalf("failed to read input: %v", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println("\n Antoni >>> Goodbye")
			break
		}

		// Simple, clear extraction passing a unique string ID ("terminal-user")
		reply, err := antoniAgent.HandleMessage(ctx, "terminal-user", input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n[error during this turn: %v]\n", err)
			continue
		}

		fmt.Println(reply)
	}
}
