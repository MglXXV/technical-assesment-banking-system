package chat

import (
	"banking-system/src/controllers"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type ChatRequest struct {
	Message string                         `json:"message"`
	History []openai.ChatCompletionMessage `json:"history"` // Optional
}

func ChatHandler(engine *MCPEngine, ledger *controllers.LedgerController) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Message string `json:"message"`
			History []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"history"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}

		uID, _ := c.Get("userID")
		ctx := context.Background()

		messages := []openai.ChatCompletionMessage{
			{
				Role: "system",
				Content: "You are a professional banking assistant from Nexora Bank. " +
					"CRITICAL RULES YOU MUST ALWAYS FOLLOW:\n\n" +
					"1. BEFORE executing deposit_money, withdraw_money, or transfer_money, " +
					"ALWAYS call get_total_balance first in the same turn to get " +
					"the exact account_numbers from the database.\n" +
					"2. ONLY USE the account_numbers returned by get_total_balance. " +
					"NEVER use the number the user typed directly.\n" +
					"3. If the user mentions an account, identify which of the real accounts " +
					"is the closest match and use it.\n" +
					"4. You can make multiple tool calls in a single response " +
					"(first get_total_balance, then the operation).\n\n" +
					"AVAILABLE TOOLS:\n" +
					"- get_total_balance: ALWAYS CALL FIRST before operating\n" +
					"- get_transaction_history: movement history (parameter: account_id with the exact account_number)\n" +
					"- deposit_money: deposit funds from the bank to the user (to_account: exact account_number from get_total_balance)\n" +
					"- withdraw_money: withdraw funds (from_account: exact account_number from get_total_balance)\n" +
					"- transfer_money: transfer between accounts\n\n" +
					"Respond in Spanish, briefly and professionally. " +
					"Always confirm the amount and account before reporting success.",
			},
		}

		// We only accept user/assistant roles from the frontend, tool/system are ignored
		for _, h := range input.History {
			if h.Content == "" {
				continue
			} // skip nulls
			if h.Role != "user" && h.Role != "assistant" {
				continue
			}
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    h.Role,
				Content: h.Content,
			})
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    "user",
			Content: input.Message,
		})

		req := openai.ChatCompletionRequest{
			Model:     engine.Model,
			Messages:  messages,
			Tools:     GetBankingTools(),
			MaxTokens: 500,
		}

		resp, err := engine.Client.CreateChatCompletion(ctx, req)
		if err != nil {
			c.JSON(500, gin.H{"error": "AI Provider Error: " + err.Error()})
			return
		}

		msg := resp.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			messages = append(messages, msg)

			for _, toolCall := range msg.ToolCalls {
				var args map[string]interface{}
				json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

				result, err := ExecuteTool(c, toolCall.Function.Name, args, ledger, uID.(string))
				if err != nil {
					c.JSON(500, gin.H{"error": "Tool execution failed: " + err.Error()})
					return
				}

				resultJSON, _ := json.Marshal(result)
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(resultJSON),
					ToolCallID: toolCall.ID,
				})
			}

			finalResp, err := engine.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:     engine.Model,
				Messages:  messages,
				MaxTokens: 500,
			})
			if err != nil {
				c.JSON(500, gin.H{"error": "AI final response failed: " + err.Error()})
				return
			}

			finalMsg := finalResp.Choices[0].Message

			replyContent := finalMsg.Content
			if replyContent == "" {
				for _, m := range messages {
					if m.Role == "tool" && m.Content != "" {
						var toolResult map[string]interface{}
						if err := json.Unmarshal([]byte(m.Content), &toolResult); err == nil {
							if status, ok := toolResult["status"].(string); ok && status == "success" {
								if tx, ok := toolResult["tx"].(string); ok {
									replyContent = fmt.Sprintf("✅ Operación realizada exitosamente.\nID de referencia: %s", tx)
								} else if msg, ok := toolResult["message"].(string); ok {
									replyContent = fmt.Sprintf("✅ %s", msg)
								} else {
									replyContent = "✅ Operación realizada exitosamente."
								}
							} else if status == "error" {
								if errMsg, ok := toolResult["message"].(string); ok {
									replyContent = fmt.Sprintf("❌ %s", errMsg)
								}
							}
						}
						break
					}
				}
			}

			if replyContent == "" {
				replyContent = "✅ Comando procesado."
			}

			var cleanHistory []map[string]string
			for _, m := range messages {
				if (m.Role == "user" || m.Role == "assistant") && m.Content != "" {
					cleanHistory = append(cleanHistory, map[string]string{
						"role":    m.Role,
						"content": m.Content,
					})
				}
			}
			cleanHistory = append(cleanHistory, map[string]string{
				"role":    "assistant",
				"content": replyContent,
			})

			c.JSON(200, gin.H{
				"reply":   replyContent,
				"history": cleanHistory,
			})
			return
		}

		// Simple response
		// In ChatHandler, when building cleanHistory, also include the tools context
		var cleanHistory []map[string]string
		for _, m := range messages {
			if m.Role == "user" && m.Content != "" {
				cleanHistory = append(cleanHistory, map[string]string{
					"role":    "user",
					"content": m.Content,
				})
			}
			if m.Role == "assistant" && m.Content != "" {
				cleanHistory = append(cleanHistory, map[string]string{
					"role":    "assistant",
					"content": m.Content,
				})
			}
			// Include tool results as assistant context
			if m.Role == "tool" && m.Content != "" {
				cleanHistory = append(cleanHistory, map[string]string{
					"role":    "assistant",
					"content": fmt.Sprintf("[Tool result: %s]", m.Content),
				})
			}
		}

		c.JSON(200, gin.H{
			"reply":   msg.Content,
			"history": cleanHistory,
		})
	}
}
