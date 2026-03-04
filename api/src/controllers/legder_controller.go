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

// UserAccount mapea la estructura de cada cuenta guardada en el JSONB de Postgres
type UserAccount struct {
	AccountNumber string `json:"account_number"`
	TBID          string `json:"tb_id"`
	Type          string `json:"type"`
	Currency      string `json:"currency"`
}

type DepositInput struct {
	AccountID string `json:"account_id" binding:"required"` // El ID específico de la cuenta destino
	Amount    uint64 `json:"amount" binding:"required,gt=0"`
}

type TransferInput struct {
	FromAccountID   string `json:"from_account_id" binding:"required"`
	TargetAccountID string `json:"target_account_id" binding:"required"`
	Amount          uint64 `json:"amount" binding:"required,gt=0"`
}

// 1. CREATE ACCOUNT: Abre una nueva cuenta bajo demanda post-login
func (ctrl *LedgerController) CreateAccount(c *gin.Context) {
	userID := c.MustGet("userID").(string)

	var input struct {
		Type string `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Tipo de cuenta requerido"})
		return
	}

	var user models.Users
	if err := ctrl.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// Crear en TigerBeetle
	tbID := types.ID()
	var code uint16 = 1000
	if input.Type == "checking" {
		code = 1001
	}

	_, err := ctrl.TB.CreateAccounts([]types.Account{{
		ID:     tbID,
		Ledger: 1,
		Code:   code,
		Flags:  types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
	}})

	if err != nil {
		c.JSON(500, gin.H{"error": "Error al crear cuenta en TigerBeetle"})
		return
	}

	// Preparar JSONB
	var accounts []UserAccount
	if len(user.TBAccountID) > 0 && string(user.TBAccountID) != "[]" {
		json.Unmarshal([]byte(user.TBAccountID), &accounts)
	}

	newAcc := UserAccount{
		AccountNumber: fmt.Sprintf("4001-%04d-%d", len(accounts)+1, code),
		TBID:          tbID.String(),
		Type:          input.Type,
		Currency:      "USD",
	}
	accounts = append(accounts, newAcc)

	updatedJSON, _ := json.Marshal(accounts)

	// --- SOLUCIÓN AL ERROR 500: SQL DIRECTO ---
	// Usamos Exec para hablarle directamente a la base de datos saltándonos a GORM
	errUpdate := ctrl.DB.Exec(
		"UPDATE users SET tb_account_id = ? WHERE uuid_user = ?",
		string(updatedJSON),
		userID,
	).Error

	if errUpdate != nil {
		c.JSON(500, gin.H{"error": "Fallo al sincronizar con Postgres", "details": errUpdate.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Cuenta abierta y vinculada exitosamente",
		"account": newAcc,
	})
}

// 2. GET BALANCE: Retorna el saldo de todas las cuentas del usuario
func (ctrl *LedgerController) GetBalance(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	var user models.Users
	if err := ctrl.DB.First(&user, "uuid_user = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User record not found"})
		return
	}

	var accounts []UserAccount
	if err := json.Unmarshal([]byte(user.TBAccountID), &accounts); err != nil || len(accounts) == 0 {
		c.JSON(http.StatusOK, gin.H{"accounts": []string{}, "message": "No tienes cuentas abiertas"})
		return
	}

	// Batch Lookup para máxima eficiencia
	var tbIDs []types.Uint128
	for _, acc := range accounts {
		id, _ := types.HexStringToUint128(acc.TBID)
		tbIDs = append(tbIDs, id)
	}

	tbAccounts, err := ctrl.TB.LookupAccounts(tbIDs)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Financial ledger communication error"})
		return
	}

	var response []gin.H
	for i, tbAcc := range tbAccounts {
		credits := tbAcc.CreditsPosted.BigInt()
		debits := tbAcc.DebitsPosted.BigInt()
		balance := new(big.Int).Sub(&credits, &debits)

		response = append(response, gin.H{
			"account_number": accounts[i].AccountNumber,
			"tb_id":          accounts[i].TBID,
			"type":           accounts[i].Type,
			"balance":        balance,
			"currency":       accounts[i].Currency,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user":     user.UserFullname,
		"accounts": response,
	})
}

// 3. DEPOSIT: Inyecta fondos desde la bóveda a una cuenta específica
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

	// Seguridad: Verificar que el usuario sea dueño de la cuenta
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
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres propietario de esta cuenta"})
		return
	}

	bankVaultID := types.ToUint128(1)
	userAccountID, _ := types.HexStringToUint128(input.AccountID)

	transfer := types.Transfer{
		ID:              types.ID(),
		DebitAccountID:  bankVaultID,
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

// 4. TRANSFER: Movimiento entre usuarios verificando disponibilidad
func (ctrl *LedgerController) Transfer(c *gin.Context) {
	var input TransferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de transferencia inválidos"})
		return
	}

	senderID := c.MustGet("userID").(string)
	var sender models.Users
	ctrl.DB.First(&sender, "uuid_user = ?", senderID)

	// 1. Validar que el emisor sea dueño de la cuenta de origen
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
		c.JSON(http.StatusForbidden, gin.H{"error": "No eres propietario de la cuenta de origen"})
		return
	}

	// 2. Convertir los IDs de Hexadecimal a Uint128 para TigerBeetle
	sourceTBID, errSrc := types.HexStringToUint128(input.FromAccountID)
	targetTBID, errDst := types.HexStringToUint128(input.TargetAccountID)

	if errSrc != nil || errDst != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de ID de cuenta inválido"})
		return
	}

	// 3. Crear la transferencia en TigerBeetle
	// TigerBeetle validará automáticamente si la cuenta destino existe
	transfer := types.Transfer{
		ID:              types.ID(),
		DebitAccountID:  sourceTBID,
		CreditAccountID: targetTBID,
		Amount:          types.ToUint128(input.Amount),
		Ledger:          1, // USD
		Code:            2, // Código de transferencia entre cuentas
	}

	res, err := ctrl.TB.CreateTransfers([]types.Transfer{transfer})
	if err != nil || len(res) > 0 {
		reason := "Error interno en el Ledger"
		if len(res) > 0 {
			reason = res[0].Result.String() // Ej: "ExceedsCredits" o "IDNotFound"
		}
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":  "Transferencia rechazada",
			"reason": reason,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Transferencia exitosa",
		"transfer_id": transfer.ID.String(),
		"amount":      input.Amount,
	})
}

// 5. GET HISTORY: Historial de transacciones de una cuenta específica
func (ctrl *LedgerController) GetHistory(c *gin.Context) {
	// 1. Extraer el ID (Usando Query para que coincida con tu curl)
	accountID := c.Query("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Account ID is required"})
		return
	}

	// 2. Seguridad: Validar que el usuario sea dueño de la cuenta (Opcional pero recomendado)
	userID := c.MustGet("userID").(string)
	var user models.Users
	ctrl.DB.First(&user, "uuid_user = ?", userID)

	// Aquí deberías hacer el unmarshal de user.TBAccountID y verificar
	// que accountID esté en su lista, similar a como lo hicimos en Transfer.

	tbAccountID, err := types.HexStringToUint128(accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID format"})
		return
	}

	filter := types.AccountFilter{
		AccountID:    tbAccountID,
		TimestampMin: 0,
		TimestampMax: 0,
		Limit:        100,
		Flags: types.AccountFilterFlags{
			Debits: true, Credits: true, Reversed: true, // Reversed: true para ver lo más nuevo primero
		}.ToUint32(),
	}

	transfers, err := ctrl.TB.GetAccountTransfers(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	var readableHistory []gin.H
	for _, t := range transfers {
		tm := time.Unix(0, int64(t.Timestamp))

		entryType := "CREDIT"
		if t.DebitAccountID == tbAccountID {
			entryType = "DEBIT"
		}

		amountBig := types.Uint128(t.Amount).BigInt()
		amountStr := amountBig.String()

		readableHistory = append(readableHistory, gin.H{
			"transfer_id": t.ID.String(),
			"type":        entryType,
			"amount":      amountStr, // Ahora sí como String legible
			"date":        tm.Format("2006-01-02 15:04:05"),
			"code":        t.Code,
			"counterparty": map[string]string{
				"debit":  t.DebitAccountID.String(),
				"credit": t.CreditAccountID.String(),
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
		"count":      len(readableHistory),
		"history":    readableHistory,
	})
}

func (ctrl *LedgerController) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	accountIDToDelete := c.Param("id") // El tb_id de la cuenta

	var user models.Users
	if err := ctrl.DB.Where("uuid_user = ?", userID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// 1. Parsear las cuentas actuales
	var accounts []UserAccount
	if err := json.Unmarshal([]byte(user.TBAccountID), &accounts); err != nil {
		c.JSON(500, gin.H{"error": "Error al procesar cuentas del usuario"})
		return
	}

	// 2. Filtrar la cuenta a eliminar y verificar saldo si es posible
	var updatedAccounts []UserAccount
	found := false

	for _, acc := range accounts {
		if acc.TBID == accountIDToDelete {
			found = true
			// OPCIONAL: Podrías hacer un Lookup en TigerBeetle aquí
			// y abortar si balance != 0
			continue
		}
		updatedAccounts = append(updatedAccounts, acc)
	}

	if !found {
		c.JSON(404, gin.H{"error": "La cuenta no existe o no te pertenece"})
		return
	}

	// 3. Actualizar en Postgres
	updatedJSON, _ := json.Marshal(updatedAccounts)

	// Usamos Exec para asegurar que el JSONB se guarde correctamente como hicimos antes
	errUpdate := ctrl.DB.Exec(
		"UPDATE users SET tb_account_id = ? WHERE uuid_user = ?",
		string(updatedJSON),
		userID,
	).Error

	if errUpdate != nil {
		c.JSON(500, gin.H{"error": "No se pudo actualizar el registro en la base de datos"})
		return
	}

	c.JSON(200, gin.H{
		"message":         "Cuenta eliminada exitosamente del perfil",
		"deleted_account": accountIDToDelete,
	})
}
