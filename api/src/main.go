package main

import (
	"banking-system/src/chat"
	"banking-system/src/controllers"
	"banking-system/src/middleware"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// In main.go after tbClient initialization:
func bootstrapLedger(tbClient tigerbeetle.Client) {
	bankVaultID := types.ToUint128(1)

	// Check if vault exists
	accounts, err := tbClient.LookupAccounts([]types.Uint128{bankVaultID})
	if err == nil && len(accounts) == 0 {
		log.Println("Initializing Bank Vault (Account ID: 1)...")
		_, err := tbClient.CreateAccounts([]types.Account{
			{
				ID:     bankVaultID,
				Ledger: 1,
				Code:   1,                                                                // System Reserve Code
				Flags:  types.AccountFlags{CreditsMustNotExceedDebits: false}.ToUint16(), // Vault can have negative balance to "issue" money
			},
		})
		if err != nil {
			log.Printf("Warning: Could not create Bank Vault: %v", err)
		}
	}
}

func main() {
	// 1. Configure PostgreSQL connection
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to Postgres: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// 2. Configure TigerBeetle connection
	tbAddress := os.Getenv("TB_ADDRESS")
	if tbAddress == "" {
		tbAddress = "bank_ledger:3000"
	}

	parts := strings.Split(tbAddress, ":")
	if len(parts) == 2 {
		ips, err := net.LookupIP(parts[0])
		if err == nil && len(ips) > 0 {
			tbAddress = ips[0].String() + ":" + parts[1]
		}
	}

	log.Printf("Attempting to connect to TigerBeetle at IP: %s", tbAddress)

	tbClient, err := tigerbeetle.NewClient(types.ToUint128(0), []string{tbAddress})
	if err != nil {
		log.Fatalf("Error connecting to TigerBeetle: %v", err)
	}
	defer tbClient.Close()
	log.Println("Connected to TigerBeetle")
	bootstrapLedger(tbClient)

	// 3. Configure Gin (Web Server)
	r := gin.Default()
	r.SetTrustedProxies(nil)

	// Setup Routes
	authCtrl := &controllers.AuthController{DB: db, TB: tbClient}
	ledgerCtrl := &controllers.LedgerController{DB: db, TB: tbClient}

	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	engine, err := chat.NewMCPEngine(ctx, option.WithAPIKey(apiKey), option.WithEndpoint("https://generativelanguage.googleapis.com/v1beta"))
	if err != nil {
		log.Fatalf("❌ Error fatal inicializando Gemini: %v", err)
	}

	// Verifica que no sea nil justo antes de pasarla
	if engine == nil || engine.Model == nil {
		log.Fatal("❌ El engine se creó pero el Modelo es nil")
	}

	// Single API Group
	api := r.Group("/api")
	{
		// Public Endpoints
		api.POST("/register", authCtrl.Register)
		api.POST("/login", authCtrl.Login)
		api.GET("/users", authCtrl.ListUsers)

		// Protected Endpoints Sub-group
		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			protected.POST("/logout", authCtrl.Logout)

			protected.POST("/accounts/create", ledgerCtrl.CreateAccount)

			protected.GET("/balance", ledgerCtrl.GetBalance)
			protected.POST("/deposit", ledgerCtrl.Deposit)
			protected.POST("/transfer", ledgerCtrl.Transfer)
			protected.GET("/history", ledgerCtrl.GetHistory)
			protected.POST("/chat", chat.ChatHandler(engine, ledgerCtrl))
			protected.DELETE("/accounts/:id", ledgerCtrl.DeleteAccount)
		}
	}

	// Health check (at the root)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	// Root check to verify server is alive
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Banking API is running")
	})

	log.Println("🚀 API server started on port 8080")
	r.Run("0.0.0.0:8080")
}
