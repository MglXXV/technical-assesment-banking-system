package chat

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func GetBankingTools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_total_balance",
				Description: "Gets the total available balance in all user accounts.",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "transfer_money",
				Description: "Transfers money to another account. Requires prior user confirmation.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"to_account": map[string]interface{}{"type": "string", "description": "Destination account number"},
						"amount":     map[string]interface{}{"type": "number", "description": "Amount in USD"},
						"confirmed":  map[string]interface{}{"type": "boolean", "description": "Indicates if the user has explicitly confirmed the operation"},
					},
					"required": []string{"to_account", "amount"},
				},
			},
		},

		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "withdraw_money",
				Description: "Records a cash withdrawal from a specific user account.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"amount":       {Type: jsonschema.Number, Description: "Amount in USD"},
						"from_account": {Type: jsonschema.String, Description: "Source account number (e.g., 4001-0001-1000)"},
						"confirmed":    {Type: jsonschema.Boolean, Description: "True only if the user confirmed."},
					},
					Required: []string{"amount", "confirmed"}, // from_account is optional for flexibility
				},
			},
		},

		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "deposit_money",
				Description: "Deposits funds from the system treasury into a user's account.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"amount":     {Type: jsonschema.Number, Description: "Amount to deposit"},
						"to_account": {Type: jsonschema.String, Description: "Destination account number (e.g., 4001-0001-1001)"},
					},
					Required: []string{"amount", "to_account"},
				},
			},
		},

		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "get_transaction_history",
				Description: "Gets the transaction history (movements, deposits, withdrawals) of a specific user account. Use it when the user asks for 'movements', 'history', 'transactions', or 'last movements'.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"account_id": {
							Type:        jsonschema.String,
							Description: "Account number (e.g., 4001-0001-1001)",
						},
					},
					Required: []string{"account_id"},
				},
			},
		},

		{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        "create_account",
				Description: "Creates a new bank account for the user in the Ledger. Use it when the user asks to open or create a savings, checking, or investment account.",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"type": {
							Type:        jsonschema.String,
							Description: "The account type. Can be 'Savings', 'Checking', or 'Investment'.",
						},
					},
					Required: []string{"type"},
				},
			},
		},
	}
}
