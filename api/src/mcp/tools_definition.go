package mcp

import "github.com/sashabaranov/go-openai"

func GetBankingToolsDefinition() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_total_balance",
				Description: "Consulta el estado actual de todas las cuentas del usuario, incluyendo saldos disponibles y números de cuenta.",
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_transaction_history",
				Description: "Devuelve el saldo real de todas las cuentas del usuario actual directamente desde el ledger.",
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
				// Descripción ultra-detallada para evitar el error de inversión de cuentas
				Description: "Ejecuta movimientos de dinero en el ledger. " +
					"REGLAS DE DIRECCIÓN: " +
					"1. DEPOSITAR: from_account='SYSTEM', to_account='Cuenta_del_Usuario'. " +
					"2. RETIRAR: from_account='Cuenta_del_Usuario', to_account='SYSTEM'. " +
					"3. TRANSFERIR: usar números de cuenta reales en ambos campos.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from_account": map[string]interface{}{
							"type":        "string",
							"description": "Cuenta de origen de los fondos. Usar 'SYSTEM' para cargar saldo desde el banco.",
						},
						"to_account": map[string]interface{}{
							"type":        "string",
							"description": "Cuenta de destino de los fondos. Usar 'SYSTEM' para retirar dinero hacia el banco.",
						},
						"amount": map[string]interface{}{
							"type":        "number",
							"description": "Monto positivo en USD a transferir. No usar valores negativos.",
						},
					},
					"required": []string{"from_account", "to_account", "amount"},
				},
			},
		},
	}
}