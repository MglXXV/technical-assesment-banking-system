package chat

import (
	"os"

	"github.com/sashabaranov/go-openai"
)

type MCPEngine struct {
	Client *openai.Client
	Model  string
}

func NewMCPEngine() *MCPEngine {
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	// We configure the OpenAI SDK to talk to OpenRouter
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	client := openai.NewClientWithConfig(config)

	return &MCPEngine{
		Client: client,
		// We use Bytedance Seed 2.0 Mini (Free Version)
		Model: "bytedance-seed/seed-2.0-mini",
	}
}
