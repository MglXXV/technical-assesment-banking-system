package chat

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type MCPEngine struct {
	Client *genai.Client
	Model  *genai.GenerativeModel
}

// En src/mcp/engine.go

func NewMCPEngine(ctx context.Context, opts ...option.ClientOption) (*MCPEngine, error) {
	client, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// USA ESTE NOMBRE EXACTO (es el que aparece en tu lista)
	model := client.GenerativeModel("models/gemini-2.0-flash")

	// Ya podemos activar las herramientas con confianza
	model.Tools = GetBankingTools()

	return &MCPEngine{
		Client: client,
		Model:  model,
	}, nil
}
