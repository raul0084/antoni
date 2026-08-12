package agentcore

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

type Agent struct {
	Config         *Config
	Runner         *runner.Runner
	SessionService session.Service
	userSessionIDs sync.Map // Maps userID -> ADK Session ID string
}

func NewAgent(ctx context.Context, cfg *Config) (*Agent, error) {
	// 1. Initialize Gemini Model
	model, err := gemini.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{
		APIKey: cfg.GoogleAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini model: %w", err)
	}

	// 2. Initialize Tools
	logTool, err := NewLogRewriteTool(cfg.LogFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log rewrite tool: %w", err)
	}
	polishCheckTool, err := NewCheckPolishTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create check polish tool: %w", err)
	}

	// 3. Configure the ADK Assistant
	assistant, err := llmagent.New(llmagent.Config{
		Name:        "antoni",
		Model:       model,
		Description: "A friendly assistant that rewrites any message into clear, business-appropriate language.",
		Instruction: "You are a message-polishing assistant. When given any message, rewrite it to sound professional and business-appropriate while staying warm, natural, and human, never stiff or corporate. Detect the language of the input message and respond in that same language, don't translate to a different one. Keep the original meaning, tone of intent, and key details exactly intact. Use contractions and plain language where natural. Avoid buzzwords, over-formal phrasing, and robotic transitions. Keep the length close to the original unless clarity requires more. Use the check_already_polished tool first; if the message is already polished, say so instead of rewriting. Otherwise use log_rewrite to record the change. Output only the rewritten message, with no preamble or explanation.",
		Tools:       []tool.Tool{logTool, polishCheckTool},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create llm agent: %w", err)
	}

	// 4. Set up the memory storage engine and runner
	sessions := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        "antoni-app",
		Agent:          assistant,
		SessionService: sessions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Agent{
		Config:         cfg,
		Runner:         r,
		SessionService: sessions,
	}, nil
}

func (a *Agent) HandleMessage(ctx context.Context, userID string, userPrompt string) (string, error) {
	// Find or dynamically register an ADK session ID for this specific user
	var sessionID string
	if val, ok := a.userSessionIDs.Load(userID); ok {
		sessionID = val.(string)
	} else {
		created, err := a.SessionService.Create(ctx, &session.CreateRequest{
			AppName: "antoni-app",
			UserID:  userID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create session context for %s: %w", userID, err)
		}
		sessionID = created.Session.ID()
		a.userSessionIDs.Store(userID, sessionID)
	}

	// Create the wrapper content type expected by ADK
	msg := genai.NewContentFromText(userPrompt, genai.RoleUser)

	var sb strings.Builder
	// Execute the runner and consume the channel output stream
	for event, err := range a.Runner.Run(ctx, userID, sessionID, msg, agent.RunConfig{}) {
		if err != nil {
			return "", fmt.Errorf("error during agent tool/thought turn: %w", err)
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				sb.WriteString(part.Text)
			}
		}
	}

	return sb.String(), nil
}
