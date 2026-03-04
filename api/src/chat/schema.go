package chat

import "github.com/google/generative-ai-go/genai"

// GetBankingTools define las capacidades del agente
func GetBankingTools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "get_balance",
					Description: "Muestra el saldo disponible de todas las cuentas del usuario.",
				},
				{
					Name:        "get_history",
					Description: "Muestra la lista de transacciones recientes de una cuenta específica.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"account_id": {Type: genai.TypeString, Description: "ID de TigerBeetle de la cuenta"},
						},
						Required: []string{"account_id"},
					},
				},
				{
					Name:        "transfer_money",
					Description: "Mueve dinero entre cuentas. Requiere confirmación explícita del usuario.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"from_id": {Type: genai.TypeString, Description: "ID origen"},
							"to_id":   {Type: genai.TypeString, Description: "ID destino"},
							"amount":  {Type: genai.TypeNumber, Description: "Monto total"},
						},
						Required: []string{"from_id", "to_id", "amount"},
					},
				},
			},
		},
	}
}
