package chat

import (
	"context"
	"log"
	"net/http"

	"banking-system/src/controllers"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
)

func ChatHandler(engine *MCPEngine, ledger *controllers.LedgerController) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Mensaje requerido"})
			return
		}

		ctx := context.Background()
		// Iniciamos una sesión de chat con el modelo
		session := engine.Model.StartChat()

		// Enviamos el mensaje del usuario a Gemini
		resp, err := session.SendMessage(ctx, genai.Text(input.Message))
		if err != nil {
			// Esto imprimirá el error real en tu consola de Docker
			log.Printf("FALLO CRÍTICO GEMINI: %v", err)
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// Analizamos la respuesta de Gemini
		for _, part := range resp.Candidates[0].Content.Parts {
			if fnCall, ok := part.(genai.FunctionCall); ok {
				// --- LÓGICA DE HERRAMIENTAS MCP ---

				// Si la herramienta es crítica (Transferencia), devolvemos una solicitud de confirmación
				if fnCall.Name == "transfer_money" {
					c.JSON(http.StatusOK, gin.H{
						"type":    "CONFIRMATION_REQUIRED",
						"message": "La IA ha preparado una transferencia. ¿Deseas proceder?",
						"tool":    fnCall.Name,
						"args":    fnCall.Args, // Aquí vienen from_id, to_id y amount
					})
					return
				}

				// Si no es crítica (como Saldo), la ejecutamos y devolvemos el resultado a la IA
				result, _ := ExecuteTool(fnCall.Name, fnCall.Args, ledger, c)
				c.JSON(http.StatusOK, gin.H{
					"type":    "DATA_RESPONSE",
					"reply":   "Aquí tienes la información solicitada:",
					"content": result,
				})
				return
			}
		}

		// Si fue solo una respuesta de texto normal
		c.JSON(http.StatusOK, gin.H{
			"type":  "TEXT_RESPONSE",
			"reply": resp.Candidates[0].Content.Parts[0],
		})
	}
}
