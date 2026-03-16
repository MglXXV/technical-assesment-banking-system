package mcp

import "github.com/sashabaranov/go-openai"

func GetBankingToolsDefinition() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_total_balance",
				Description: "Consulta el saldo real de todas las cuentas del usuario actual directamente desde el ledger.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_history",
				Description: "Devuelve el historial de transacciones, depósitos, retiros y movimientos de una cuenta bancaria específica.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"account_number": map[string]interface{}{
							"type":        "string",
							"description": "El número de cuenta en formato 4001-XXXX-XXXX.",
						},
					},
					"required": []string{"account_number"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name: "transfer_funds",
				Description: "Ejecuta movimientos de dinero. " +
					"REGLA: NO preguntes al usuario por la cuenta 'SYSTEM', úsala automáticamente. " +
					"1. DEPOSITAR: from_account='SYSTEM', to_account='Cuenta_Usuario'. " +
					"2. RETIRAR: from_account='Cuenta_Usuario', to_account='SYSTEM'. " +
					"3. TRANSFERIR: usa los números de cuenta reales.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_account": map[string]interface{}{
							"type":        "string",
							"description": "Cuenta origen.",
						},
						"to_account": map[string]interface{}{
							"type":        "string",
							"description": "Cuenta destino.",
						},
						"amount": map[string]interface{}{
							"type":        "number",
							"description": "Monto de la operación.",
						},
					},
					"required": []string{"from_account", "to_account", "amount"},
				},
			},
		},
	}
}
