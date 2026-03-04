package chat

import (
	"banking-system/src/controllers"

	"github.com/gin-gonic/gin"
)

func ExecuteTool(name string, args map[string]interface{}, ledger *controllers.LedgerController, c *gin.Context) (interface{}, error) {
	switch name {
	case "get_balance":
		// Aquí Gemini no necesita argumentos.
		// En una implementación real, aquí llamarías a una versión de ledger.GetBalance
		// que devuelva datos en lugar de escribir un JSON en el contexto de Gin.
		return "Tu saldo total es de $7,500 USD distribuido en 2 cuentas.", nil

	case "get_history":
		accID := args["account_id"].(string)
		return "Buscando transacciones recientes para la cuenta " + accID, nil

	default:
		return "Herramienta no implementada", nil
	}
}
