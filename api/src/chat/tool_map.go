package chat

import (
	"banking-system/src/controllers"
	"banking-system/src/models"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

func ExecuteTool(ctx *gin.Context, name string, args map[string]interface{}, ledger *controllers.LedgerController, userID string) (interface{}, error) {
	switch name {
	case "get_total_balance":
		var user models.Users
		ledger.DB.Where("uuid_user = ?", userID).First(&user)

		var accounts []controllers.UserAccount // Using the struct we defined before
		json.Unmarshal([]byte(user.TBAccountID), &accounts)

		type AccountResult struct {
			AccountNumber string  `json:"account_number"`
			Balance       float64 `json:"balance"`
			Currency      string  `json:"currency"`
		}
		var results []AccountResult

		for _, acc := range accounts {
			// We convert the TBID of this specific account
			tbID, _ := types.HexStringToUint128(acc.TBID)

			// We query TigerBeetle for this specific ID
			tbAcc, _ := ledger.TB.LookupAccounts([]types.Uint128{tbID})

			balance := 0.0
			if len(tbAcc) > 0 {
				credits := tbAcc[0].CreditsPosted.BigInt()
				debits := tbAcc[0].DebitsPosted.BigInt()
				balance = float64(new(big.Int).Sub(&credits, &debits).Int64()) / 100.0
			}

			results = append(results, AccountResult{
				AccountNumber: acc.AccountNumber,
				Balance:       balance,
				Currency:      acc.Currency,
			})
		}

		return map[string]interface{}{"accounts": results, "status": "success"}, nil
	case "get_transaction_history":
		accountID, _ := args["account_id"].(string)
		fmt.Printf("DEBUG: Looking for history for account: %s\n", accountID)

		// GetTigerBeetleHistory already does the DB lookup + TB query internally.
		// Just pass userID + accountNumber — don't duplicate the lookup here.
		history, err := ledger.GetTigerBeetleHistory(userID, accountID)
		if err != nil {
			return map[string]interface{}{"status": "error", "message": err.Error()}, nil
		}

		return map[string]interface{}{
			"status":  "success",
			"history": history,
		}, nil

	case "deposit_money":
		amount, _ := args["amount"].(float64)
		toAcc, _ := args["to_account"].(string)

		var user models.Users
		if err := ledger.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
			return map[string]interface{}{"status": "error", "message": "User not found"}, nil
		}

		var accounts []controllers.UserAccount
		if err := json.Unmarshal([]byte(user.TBAccountID), &accounts); err != nil {
			return map[string]interface{}{"status": "error", "message": "Error processing accounts"}, nil
		}

		var targetTBID string
		for _, acc := range accounts {
			if acc.AccountNumber == toAcc {
				targetTBID = acc.TBID
				break
			}
		}

		if targetTBID == "" {
			return map[string]interface{}{"status": "error", "message": fmt.Sprintf("Account %s not found", toAcc)}, nil
		}

		// FIX: pass targetTBID (hex) not toAcc (account_number)
		txID, err := ledger.InternalTransfer("SYSTEM", targetTBID, amount)
		if err != nil {
			return map[string]interface{}{"status": "error", "message": err.Error()}, nil
		}
		return map[string]interface{}{"status": "success", "tx": txID}, nil

	case "withdraw_money":
		amount, _ := args["amount"].(float64)
		fromAcc, _ := args["from_account"].(string)

		var user models.Users
		if err := ledger.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
			return map[string]interface{}{"status": "error", "message": "User not found"}, nil
		}

		var accounts []controllers.UserAccount
		if err := json.Unmarshal([]byte(user.TBAccountID), &accounts); err != nil {
			return map[string]interface{}{"status": "error", "message": "Error processing accounts"}, nil
		}

		var sourceTBID string
		for _, acc := range accounts {
			if acc.AccountNumber == fromAcc {
				sourceTBID = acc.TBID
				break
			}
		}

		if sourceTBID == "" {
			return map[string]interface{}{"status": "error", "message": fmt.Sprintf("Account %s not found", fromAcc)}, nil
		}

		// FIX: pass sourceTBID (hex) and expose the real error
		txID, err := ledger.InternalTransfer(sourceTBID, "SYSTEM", amount)
		if err != nil {
			return map[string]interface{}{"status": "error", "message": err.Error()}, nil
		}
		return map[string]interface{}{"status": "success", "tx": txID}, nil

	default:
		return nil, fmt.Errorf("tool '%s' not implemented", name)
	}
}
