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

			// PARSEO SEGURO DEL MONTO (Evita que la IA rompa Go)
			var amount float64
			switch v := args["amount"].(type) {
			case float64:
				amount = v
			case int:
				amount = float64(v)
			case string:
				fmt.Sscanf(v, "%f", &amount)
			}

			if amount <= 0 {
				return nil, fmt.Errorf("el monto debe ser mayor a cero")
			}

			txID, err := r.ledger.InternalTransfer(from, to, amount)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"status":      "success",
				"transfer_id": txID,
				"message":     "Operación asentada en el Ledger correctamente. ✅",
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
	fmt.Printf("\n--- 🤖 INICIO LLAMADA A HERRAMIENTA ---\n")
	fmt.Printf("▶️ Nombre: %s\n", name)
	fmt.Printf("▶️ Args  : %s\n", jsonArgs)

	var args map[string]interface{}
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		fmt.Printf("❌ ERROR DE JSON: %v\n", err)
		return "", fmt.Errorf("error parseando argumentos: %v", err)
	}

	tool, ok := r.tools[name]
	if !ok {
		fmt.Printf("❌ ERROR: La herramienta no existe\n")
		return "", fmt.Errorf("la herramienta %s no existe", name)
	}

	result, err := tool.Execute(args, userID)
	if err != nil {
		// 🔥 AQUÍ ESTÁ EL SECRETO: Esto imprimirá el error real de TigerBeetle o Postgres
		fmt.Printf("❌ ERROR FATAL DE EJECUCIÓN: %v\n", err)
		fmt.Printf("---------------------------------------\n")
		return fmt.Sprintf(`{"error": "El servidor rechazó la operación por este motivo: %v"}`, err), nil
	}

	fmt.Printf("✅ ÉXITO: Operación completada en el Ledger\n")
	fmt.Printf("---------------------------------------\n")

	resJSON, _ := json.Marshal(result)
	return string(resJSON), nil
}
