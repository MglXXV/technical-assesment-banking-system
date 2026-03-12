package mcp

import (
	"banking-system/src/controllers"
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func ChatHandler(engine *MCPEngine, ledger *controllers.LedgerController) gin.HandlerFunc {
	registry := NewMCPRegistry(ledger)

	return func(c *gin.Context) {
		var input struct {
			Message string                         `json:"message"`
			History []openai.ChatCompletionMessage `json:"history"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Input inválido"})
			return
		}

		val, _ := c.Get("userID")
		userID, ok := val.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
			return
		}

		userAccounts, _ := ledger.GetBalancesInternal(userID)
		accountsInfo := ""
		for _, acc := range userAccounts {
			accountsInfo += fmt.Sprintf("- %s (ID TigerBeetle: %s, Saldo: %v %s)\n",
				acc["account_number"], acc["tb_id"], acc["balance"], acc["currency"])
		}

	
		messages := []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: fmt.Sprintf(`Asistente Nexora. Usuario: %s.
REGLAS:
- Cuentas reales: %s
- PARA RETIRAR: from_account=[CuentaUsuario], to_account="SYSTEM".
- PARA DEPOSITAR: from_account="SYSTEM", to_account=[CuentaUsuario].
- Tras transferir, llama a 'get_total_balance' para dar el saldo REAL.
- Confirma éxito con un ✅.`, userID, accountsInfo),
			},
		}

		
		if input.History != nil {
			limit := 5
			if len(input.History) > limit {
				messages = append(messages, input.History[len(input.History)-limit:]...)
			} else {
				messages = append(messages, input.History...)
			}
		}
		
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input.Message,
		})

		ctx := context.Background()
		
	
		resp, err := engine.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:     engine.Model,
			Messages:  messages,
			Tools:     GetBankingToolsDefinition(),
			ToolChoice: "auto",
			MaxTokens: 300, 
		})

		if err != nil {
			
			c.JSON(http.StatusOK, gin.H{"reply": "❌ Error de créditos en el motor de IA. Por favor, reduce la longitud de tu consulta o recarga créditos."})
			return
		}

		msg := resp.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			messages = append(messages, msg)
			for _, tc := range msg.ToolCalls {
				result, err := registry.Call(tc.Function.Name, tc.Function.Arguments, userID)
				if err != nil {
					result = fmt.Sprintf(`{"error":"%v"}`, err)
				}
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}

			/
			finalResp, err := engine.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:    engine.Model,
				Messages: messages,
				MaxTokens: 300,
			})
			if err == nil {
				msg = finalResp.Choices[0].Message
			}
		}

		
		cleanReply := msg.Content
		re := regexp.MustCompile(`\(?(tool_call|function|tool_calls).*?(\]|\)|\s|$)`)
		cleanReply = re.ReplaceAllString(cleanReply, "")

		c.JSON(http.StatusOK, gin.H{"reply": cleanReply})
	}
}