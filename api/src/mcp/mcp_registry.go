package mcp

import (
    "banking-system/src/controllers"
    "encoding/json"
    "fmt"
)

type MCPTool struct {
    Name        string
    Description string
    Execute     func(args map[string]interface{}, userID string) (interface{}, error)
}

type MCPRegistry struct {
    tools  map[string]MCPTool
    ledger *controllers.LedgerController
}

func NewMCPRegistry(l *controllers.LedgerController) *MCPRegistry {
    r := &MCPRegistry{
        tools:  make(map[string]MCPTool),
        ledger: l,
    }
    r.setupTools()
    return r
}

func (r *MCPRegistry) setupTools() {
	r.tools["get_total_balance"] = MCPTool{
		Name:        "get_total_balance",
		Description: "Consulta el saldo REAL y números de cuenta. Úsala SIEMPRE después de transferir para dar el saldo actualizado.",
		Execute: func(args map[string]interface{}, userID string) (interface{}, error) {
			return r.ledger.GetBalancesInternal(userID)
		},
	}

	r.tools["transfer_funds"] = MCPTool{
		Name:        "transfer_funds",
		Description: "Mueve dinero. DEPÓSITO: from='SYSTEM', to='Cuenta'. RETIRO: from='Cuenta', to='SYSTEM'.",
		Execute: func(args map[string]interface{}, userID string) (interface{}, error) {
			from, _ := args["from_account"].(string)
			to, _ := args["to_account"].(string)
			amount, _ := args["amount"].(float64)

			txID, err := r.ledger.InternalTransfer(from, to, amount)
			if err != nil {
				return nil, err
			}
		
			return map[string]interface{}{
				"status":      "success",
				"transfer_id": txID,
				"message":     "Transacción asentada en el Ledger correctamente. ✅",
			}, nil
		},
	}

	r.tools["get_history"] = MCPTool{
		Name:        "get_history",
		Description: "Obtiene los últimos movimientos de una cuenta.",
		Execute: func(args map[string]interface{}, userID string) (interface{}, error) {
			accountNum, _ := args["account_number"].(string)
			if accountNum == "" {
				accountNum, _ = args["account_id"].(string)
			}
			return r.ledger.GetTigerBeetleHistory(userID, accountNum)
		},
	}
}

func (r *MCPRegistry) Call(name string, jsonArgs string, userID string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("error parseando argumentos: %v", err)
	}

	tool, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("la herramienta %s no existe", name)
	}

	result, err := tool.Execute(args, userID)
	if err != nil {
		return fmt.Sprintf(`{"error": "%v"}`, err), nil
	}

	resJSON, _ := json.Marshal(result)
	return string(resJSON), nil
}