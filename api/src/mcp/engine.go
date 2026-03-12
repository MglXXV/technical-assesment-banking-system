package mcp

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

	model := os.Getenv("MCP_MODEL")
	if model == "" {
		model = "bytedance-seed/seed-2.0-mini" // Default free
	}

	
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://openrouter.ai/api/v1"

	client := openai.NewClientWithConfig(config)

	return &MCPEngine{
		Client: client,
		Model:  model,
	}
}