package controllers

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"gorm.io/gorm"

	"banking-system/src/models"
)

type LedgerController struct {
	DB *gorm.DB
	TB tigerbeetle.Client
}

type UserAccount struct {
	AccountNumber string `json:"account_number"`
	TBID          string `json:"tb_id"`
	Type          string `json:"type"`
	Currency      string `json:"currency"`
}

type DepositInput struct {
	AccountID string `json:"account_id" binding:"required"`
	Amount    uint64 `json:"amount" binding:"required,gt=0"`
}

type TransferInput struct {
	FromAccountID   string `json:"from_account_id" binding:"required"`
	TargetAccountID string `json:"target_account_id" binding:"required"`
	Amount          uint64 `json:"amount" binding:"required,gt=0"`
}

// CreateAccount opens a new TigerBeetle account and links it to the user in Postgres.
func (ctrl *LedgerController) CreateAccount(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var input struct {
		Type string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Account type required"})
		return
	}

	var user models.Users
	if err := ctrl.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	tbID := types.ID()
	var code uint16 = 1000
	switch input.Type {
	case "checking":
		code = 1001
	case "investment":
		code = 1002
	}

	_, err := ctrl.TB.CreateAccounts([]types.Account{{
		ID:     tbID,
		Ledger: 1,
		Code:   code,
		Flags:  types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
	}})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create account in TigerBeetle"})
		return
	}

	var accounts []UserAccount
	if len(user.TBAccountID) > 0 && user.TBAccountID != "[]" {
		json.Unmarshal([]byte(user.TBAccountID), &accounts)
	}

	newAcc := UserAccount{
		AccountNumber: fmt.Sprintf("4001-%04d-%04d", len(accounts)+1, code),
		TBID:          fmt.Sprintf("%032s", tbID.String()), // zero-pad to 32 chars
		Type:          input.Type,
		Currency:      "USD",
	}
	accounts = append(accounts, newAcc)

	updatedJSON, _ := json.Marshal(accounts)
	if err := ctrl.DB.Exec(
		"UPDATE users SET tb_account_id = ? WHERE uuid_user = ?",
		string(updatedJSON), userID,
	).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to sync with Postgres", "details": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Account opened and linked successfully", "account": newAcc})
}

// GetBalance returns the balance of all accounts for the authenticated user.
func (ctrl *LedgerController) GetBalance(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var user models.Users
	if err := ctrl.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var accounts []UserAccount
	json.Unmarshal([]byte(user.TBAccountID), &accounts)

	var tbIDs []types.Uint128
	for _, acc := range accounts {
		id, err := types.HexStringToUint128(acc.TBID)
		if err != nil {
			continue
		}
		tbIDs = append(tbIDs, id)
	}

	tbAccounts, err := ctrl.TB.LookupAccounts(tbIDs)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error querying TigerBeetle"})
		return
	}

	type AccountBalance struct {
		AccountNumber string  `json:"account_number"`
		Balance       float64 `json:"balance"`
		Currency      string  `json:"currency"`
		Type          string  `json:"type"`
		TBID          string  `json:"tb_id"`
	}
	results := make([]AccountBalance, 0, len(tbAccounts))
	for i, a := range tbAccounts {
		credits := a.CreditsPosted.BigInt()
		debits := a.DebitsPosted.BigInt()
		balance := float64(new(big.Int).Sub(&credits, &debits).Int64()) / 100.0
		results = append(results, AccountBalance{
			AccountNumber: accounts[i].AccountNumber,
			Balance:       balance,
			Currency:      accounts[i].Currency,
			Type:          accounts[i].Type,
			TBID:          accounts[i].TBID,
		})
	}

	c.JSON(200, gin.H{"accounts": results})
}

// Deposit credits funds from the bank vault into a user-owned account.
func (ctrl *LedgerController) Deposit(c *gin.Context) {
	var input DepositInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("userID").(string)
	var user models.Users
	if err := ctrl.DB.First(&user, "uuid_user = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var accounts []UserAccount
	json.Unmarshal([]byte(user.TBAccountID), &accounts)
	isOwner := false
	for _, acc := range accounts {
		if acc.TBID == input.AccountID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this account"})
		return
	}

	userAccountID, err := types.HexStringToUint128(input.AccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id format"})
		return
	}

	transfer := types.Transfer{
		ID:              types.ID(),
		DebitAccountID:  types.ToUint128(1),
		CreditAccountID: userAccountID,
		Amount:          types.ToUint128(input.Amount),
		Ledger:          1,
		Code:            1,
	}

	res, err := ctrl.TB.CreateTransfers([]types.Transfer{transfer})
	if err != nil || len(res) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Deposit failed", "details": res})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "transfer_id": transfer.ID.String()})
}

// Transfer moves funds between two TigerBeetle accounts.
func (ctrl *LedgerController) Transfer(c *gin.Context) {
	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer data"})
		return
	}

	senderID := c.MustGet("userID").(string)
	var sender models.Users
	if err := ctrl.DB.First(&sender, "uuid_user = ?", senderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 1. Validate that the source account belongs to the user
	var senderAccounts []UserAccount
	json.Unmarshal([]byte(sender.TBAccountID), &senderAccounts)
	isOwner := false
	for _, acc := range senderAccounts {
		if acc.TBID == input.FromAccountID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of the source account"})
		return
	}

	sourceTBID, errSrc := types.HexStringToUint128(input.FromAccountID)
	if errSrc != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source ID format"})
		return
	}

	// 2. RESOLVE THE DESTINATION ACCOUNT (From Account Number to TigerBeetle ID)
	var targetTBID types.Uint128

	// If it's a "withdrawal", the Dashboard sends "1" as the destination (the bank vault)
	if input.TargetAccountID == "1" {
		targetTBID = types.ToUint128(1)
	} else {
		// Find the user who has this account number in Postgres
		var targetUser models.Users
		errDst := ctrl.DB.Where("tb_account_id::jsonb @> ?", fmt.Sprintf(`[{"account_number":"%s"}]`, input.TargetAccountID)).First(&targetUser).Error
		if errDst != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "The destination account number does not exist in the bank"})
			return
		}

		// Extract the secret tb_id from that user's JSON
		var targetAccounts []UserAccount
		json.Unmarshal([]byte(targetUser.TBAccountID), &targetAccounts)
		targetFound := false
		for _, acc := range targetAccounts {
			if acc.AccountNumber == input.TargetAccountID {
				targetTBID, _ = types.HexStringToUint128(acc.TBID)
				targetFound = true
				break
			}
		}

		if !targetFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Could not resolve the destination account ID"})
			return
		}
	}

	// 3. Execute the transfer in TigerBeetle
	transfer := types.Transfer{
		ID:              types.ID(),
		DebitAccountID:  sourceTBID,
		CreditAccountID: targetTBID,
		Amount:          types.ToUint128(input.Amount),
		Ledger:          1,
		Code:            2,
	}

	res, err := ctrl.TB.CreateTransfers([]types.Transfer{transfer})
	if err != nil || len(res) > 0 {
		reason := "Insufficient funds or Ledger error"
		if len(res) > 0 {
			reason = res[0].Result.String()
		}
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Transfer rejected: " + reason})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transfer successful",
		"transfer_id": transfer.ID.String(),
	})
}

// GetHistory returns paginated transaction history for an account by its TB hex ID.
func (ctrl *LedgerController) GetHistory(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id required"})
		return
	}

	userID := c.MustGet("userID").(string)
	var user models.Users
	if err := ctrl.DB.First(&user, "uuid_user = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var accounts []UserAccount
	json.Unmarshal([]byte(user.TBAccountID), &accounts)
	isOwner := false
	for _, acc := range accounts {
		if acc.TBID == accountID {
			isOwner = true
			break
		}
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this account"})
		return
	}

	tbAccountID, err := types.HexStringToUint128(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account_id format"})
		return
	}

	filter := types.AccountFilter{
		AccountID:    tbAccountID,
		TimestampMin: 0,
		TimestampMax: 0,
		Limit:        100,
		Flags: types.AccountFilterFlags{
			Debits: true, Credits: true, Reversed: true,
		}.ToUint32(),
	}

	transfers, err := ctrl.TB.GetAccountTransfers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying history"})
		return
	}

	var history []gin.H
	for _, t := range transfers {
		tm := time.Unix(0, int64(t.Timestamp))
		entryType := "CREDIT"
		if t.DebitAccountID == tbAccountID {
			entryType = "DEBIT"
		}
		amountBig := types.Uint128(t.Amount).BigInt()
		history = append(history, gin.H{
			"transfer_id": t.ID.String(),
			"type":        entryType,
			"amount":      amountBig.String(),
			"date":        tm.Format("2006-01-02 15:04:05"),
			"code":        t.Code,
			"counterparty": map[string]string{
				"debit":  t.DebitAccountID.String(),
				"credit": t.CreditAccountID.String(),
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{"account_id": accountID, "count": len(history), "history": history})
}

// DeleteAccount removes an account from the user's profile.
// Aborts if the account still has a non-zero balance.
func (ctrl *LedgerController) DeleteAccount(c *gin.Context) {
	// The 'id' in the api.DELETE("/accounts/:id") route will now be the account_number
	accountNumberToDelete := c.Param("id")
	userID := c.MustGet("userID").(string)

	var user models.Users
	if err := ctrl.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	var accounts []UserAccount
	if err := json.Unmarshal([]byte(user.TBAccountID), &accounts); err != nil {
		c.JSON(500, gin.H{"error": "Error processing accounts JSON"})
		return
	}

	var updatedAccounts []UserAccount
	found := false

	for _, acc := range accounts {
		if acc.AccountNumber == accountNumberToDelete {
			found = true

			// SECURITY VALIDATION: Query TigerBeetle to see the real balance
			tbID, err := types.HexStringToUint128(acc.TBID)
			if err == nil {
				tbAccounts, err := ctrl.TB.LookupAccounts([]types.Uint128{tbID})
				if err == nil && len(tbAccounts) > 0 {
					credits := tbAccounts[0].CreditsPosted.BigInt()
					debits := tbAccounts[0].DebitsPosted.BigInt()
					// If the balance is greater than 0, we prevent deletion
					if new(big.Int).Sub(&credits, &debits).Sign() > 0 {
						c.JSON(400, gin.H{"error": "You cannot delete an account that still has funds"})
						return
					}
				}
			}
			// If we find it and the balance is 0, we DO NOT add it to the new array (we delete it)
			continue
		}
		updatedAccounts = append(updatedAccounts, acc)
	}

	if !found {
		c.JSON(404, gin.H{"error": "The account does not exist or does not belong to you"})
		return
	}

	// We convert the new array back to JSON
	updatedJSON, _ := json.Marshal(updatedAccounts)

	// We save in Postgres
	if err := ctrl.DB.Model(&user).Update("tb_account_id", string(updatedJSON)).Error; err != nil {
		c.JSON(500, gin.H{"error": "Error updating the database"})
		return
	}

	c.JSON(200, gin.H{
		"message":         "Account deleted successfully",
		"deleted_account": accountNumberToDelete,
	})
}

// ExtractTbID returns the TB ID of the first account in a JSONB string.
func (lc *LedgerController) ExtractTbID(jsonStr string) (types.Uint128, error) {
	var accountsData []struct {
		TbID string `json:"tb_id"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &accountsData); err != nil || len(accountsData) == 0 {
		return types.Uint128{}, fmt.Errorf("tb_id not found in metadata")
	}
	return types.HexStringToUint128(accountsData[0].TbID)
}

// InternalTransfer moves funds between two accounts identified by:
//   - "1" or "SYSTEM"  → the bank vault (TigerBeetle account 1)
//   - 32-char hex      → a raw TigerBeetle account ID
//   - account number   → e.g. "4001-0001-1001", resolved via Postgres JSONB
//   - UUID             → a user UUID, resolved to their first TB account
func (lc *LedgerController) InternalTransfer(fromIdentifier string, toIdentifier string, amount float64) (string, error) {
	var fromTbID types.Uint128
	var toTbID types.Uint128
	var err error

	if fromIdentifier == "1" || fromIdentifier == "SYSTEM" {
		fromTbID = types.ToUint128(1)
	} else if len(fromIdentifier) >= 31 && len(fromIdentifier) <= 32 {
		padded := fromIdentifier
		if len(padded) == 31 {
			padded = "0" + padded
		}
		fromTbID, err = types.HexStringToUint128(padded)
		if err != nil {
			return "", fmt.Errorf("identificador origen inválido: %v", err)
		}
	} else {
		var fromUser models.Users
		if err := lc.DB.Where("uuid_user = ?", fromIdentifier).First(&fromUser).Error; err != nil {
			return "", fmt.Errorf("source user not found")
		}
		fromTbID, err = lc.ExtractTbID(fromUser.TBAccountID)
		if err != nil {
			return "", fmt.Errorf("error extracting TBAccountID: %v", err)
		}
	}

	if toIdentifier == "1" || toIdentifier == "SYSTEM" {
		toTbID = types.ToUint128(1)
	} else if len(toIdentifier) >= 31 && len(toIdentifier) <= 32 {
		// Pad a 32 chars si tiene 31
		padded := toIdentifier
		if len(padded) == 31 {
			padded = "0" + padded
		}
		toTbID, err = types.HexStringToUint128(padded)
		if err != nil {
			return "", fmt.Errorf("identificador destino inválido: %v", err)
		}
	} else {
		var toUser models.Users
		err = lc.DB.Where("tb_account_id::jsonb @> ?", fmt.Sprintf(`[{"account_number":"%s"}]`, toIdentifier)).First(&toUser).Error
		if err != nil {
			return "", fmt.Errorf("destination account %s not found", toIdentifier)
		}
		var accounts []UserAccount
		json.Unmarshal([]byte(toUser.TBAccountID), &accounts)
		for _, acc := range accounts {
			if acc.AccountNumber == toIdentifier {
				toTbID, err = types.HexStringToUint128(acc.TBID)
				if err != nil {
					return "", fmt.Errorf("malformed tb_id in destination account: %v", err)
				}
				break
			}
		}
	}

	amountCentavos := uint64(amount * 100)
	fmt.Printf("Ledger: From %s -> To %s | Amount: %d cents\n", fromTbID.String(), toTbID.String(), amountCentavos)

	transferID := types.ID()
	res, err := lc.TB.CreateTransfers([]types.Transfer{
		{
			ID:              transferID,
			DebitAccountID:  fromTbID,
			CreditAccountID: toTbID,
			Amount:          types.ToUint128(amountCentavos),
			Ledger:          1,
			Code:            1,
		},
	})
	if err != nil {
		return "", fmt.Errorf("network error with TigerBeetle: %v", err)
	}
	if len(res) > 0 {
		return "", fmt.Errorf("rejected by Ledger: %s", res[0].Result.String())
	}

	return transferID.String(), nil
}

// GetTigerBeetleHistory returns the last 10 transactions for a given account number
// belonging to the specified user. Called by the chat tool get_transaction_history.
func (lc *LedgerController) GetTigerBeetleHistory(userID string, accountNumber string) ([]map[string]interface{}, error) {
	var user models.Users
	err := lc.DB.Where("uuid_user = ?", userID).
		Where("tb_account_id::jsonb @> ?", fmt.Sprintf(`[{"account_number":"%s"}]`, accountNumber)).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("account %s does not exist or does not belong to your user", accountNumber)
	}

	var accounts []UserAccount
	json.Unmarshal([]byte(user.TBAccountID), &accounts)

	var targetTBID types.Uint128
	for _, acc := range accounts {
		if acc.AccountNumber == accountNumber {
			targetTBID, err = types.HexStringToUint128(acc.TBID)
			if err != nil {
				return nil, fmt.Errorf("malformed tb_id for account %s: '%s' (%d chars)",
					accountNumber, acc.TBID, len(acc.TBID))
			}
			break
		}
	}

	filter := types.AccountFilter{
		AccountID:    targetTBID,
		TimestampMin: 0,
		TimestampMax: 0,
		Limit:        10,
		Flags: types.AccountFilterFlags{
			Debits: true, Credits: true, Reversed: true,
		}.ToUint32(),
	}

	transfers, err := lc.TB.GetAccountTransfers(filter)
	if err != nil {
		return nil, err
	}

	var history []map[string]interface{}
	for _, t := range transfers {
		tm := time.Unix(0, int64(t.Timestamp))
		entryType := "CREDIT"
		if t.DebitAccountID == targetTBID {
			entryType = "DEBIT"
		}
		amountBig := types.Uint128(t.Amount).BigInt()
		history = append(history, map[string]interface{}{
			"id":     t.ID.String(),
			"amount": float64(amountBig.Int64()) / 100.0,
			"type":   entryType,
			"date":   tm.Format(time.RFC3339),
		})
	}

	return history, nil
}
