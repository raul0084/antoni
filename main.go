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

	"google.golang.org/genai"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool" // was missing — needed for []tool.Tool
)

func main() {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	model, err := gemini.NewModel(ctx, cfg.model_name, &genai.ClientConfig{
		APIKey: cfg.google_api_key,
	})
	if err != nil {
		log.Fatal(err)
	}

	logTool, err := NewLogRewriteTool(cfg.log_file_path)
	if err != nil {
		log.Fatal(err)
	}
	polishCheckTool, err := NewCheckPolishTool()
	if err != nil {
		log.Fatal(err)
	}

	assistant, err := llmagent.New(llmagent.Config{
		Name:        "antoni",
		Model:       model,
		Description: "A friendly assistant that rewrites any message into clear, business-appropriate language.",
		Instruction: "You are a message-polishing assistant. When given any message, rewrite it to sound professional and business-appropriate while staying warm, natural, and human, never stiff or corporate. Detect the language of the input message and respond in that same language, don't translate to a different one. Keep the original meaning, tone of intent, and key details exactly intact. Use contractions and plain language where natural. Avoid buzzwords, over-formal phrasing, and robotic transitions. Keep the length close to the original unless clarity requires more. Use the check_already_polished tool first; if the message is already polished, say so instead of rewriting. Otherwise use log_rewrite to record the change. Output only the rewritten message, with no preamble or explanation.",
		Tools:       []tool.Tool{logTool, polishCheckTool},
	})
	if err != nil {
		log.Fatal(err)
	}

	sessions := session.InMemoryService()
	created, err := sessions.Create(ctx, &session.CreateRequest{
		AppName: "app",
		UserID:  "user",
	})
	if err != nil {
		log.Fatal(err)
	}

	r, err := runner.New(runner.Config{
		AppName:        "app",
		Agent:          assistant,
		SessionService: sessions,
	})
	if err != nil {
		log.Fatal(err)
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
			break
		}

		msg := genai.NewContentFromText(input, genai.RoleUser)

		var turnFailed bool
		for event, err := range r.Run(ctx, "user", created.Session.ID(), msg, agent.RunConfig{}) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n[error during this turn: %v]\n", err)
				turnFailed = true
				break
			}
			if event.Content != nil {
				for _, part := range event.Content.Parts {
					fmt.Print(part.Text)
				}
			}
		}
		if turnFailed {
			continue
		}
		fmt.Println()
	}
}
