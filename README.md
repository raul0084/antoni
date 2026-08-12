## 🤖 Antoni
Antoni is a Go-based, multi-turn AI message-polishing assistant powered by Google Gemini and built on top of the official Google Agent Development Kit (ADK).
It intercepts messages and refines them into warm, natural, and business-appropriate language while preserving the original intent, language, and core details. It uses embedded local tools to log rewrites and bypass already polished content. [2] 
------------------------------
## 📂 Project Architecture

jarvis/
├── go.mod                # Module definition and dependencies
├── .env                  # Environment configurations (API keys)
├── .gitignore            # Git exclusion rules
├── internal/             # Protected internal core logic
│   └── agentcore/
│       ├── bot.go        # Universal Agent runner & multi-user session engine
│       ├── config.go     # Configuration loader
│       ├── tools.go      # NewLogRewriteTool & NewCheckPolishTool definitions
│       └── tools_test.go # Automation tests
└── cmd/                  # Application entry points
    └── cli/
        └── main.go       # Interactive terminal REPL loop

------------------------------
## 🛠️ Features

* Platform Agnostic Architecture: Core AI execution logic is housed entirely within internal/agentcore, decoupling it from the terminal interface loop.
* Google ADK & Gemini Engine: Leverages google.golang.org/adk to orchestrate model actions, execute tool calls, and control conversational flows seamlessly. [3] 
* Multi-Turn User Separation: Dynamic thread synchronization isolation via sync.Map. Keeps conversations unique for separate users. [4] 
* Smart Tool Integration:
* check_already_polished: Analyzes input text to determine if changes are necessary.
   * log_rewrite: Records changes dynamically when modifications occur.

------------------------------
## 🚦 Getting Started## 1. Prerequisites

* Go 1.22 or higher installed.
* A Gemini API Key from Google AI Studio. [5] 

## 2. Setup Configuration
Clone the repository and create an .env file at the root of the project directory: [6] 

MODEL_NAME=gemini-2.5-flash # Or your specific Gemini variant
GOOGLE_API_KEY=your_gemini_api_key_here
LOG_FILE_PATH=./rewrites.log

## 3. Running the App
To start the interactive command-line environment and begin chatting with Antoni, execute the following command in your terminal:

go run ./cmd/cli/main.go

