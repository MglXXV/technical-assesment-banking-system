package main

import (
	"banking-system/src/mcp"
	"banking-system/src/controllers"
	"banking-system/src/middleware"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func bootstrapLedger(tbClient tigerbeetle.Client) {
	bankVaultID := types.ToUint128(1)

	accounts, err := tbClient.LookupAccounts([]types.Uint128{bankVaultID})
	if err != nil {
		log.Fatalf("Critical error connecting to TigerBeetle: %v", err)
	}

	if len(accounts) == 0 {
		log.Println("🏦 Initializing Bank Vault (ID: 1)...")
		_, err := tbClient.CreateAccounts([]types.Account{
			{
				ID:     bankVaultID,
				Ledger: 1,
				Code:   1,
				Flags:  0,
			},
		})
		if err != nil {
			log.Printf("❌ Error creating Vault: %v", err)
		} else {
			log.Println("✅ Vault created successfully.")
		}
	} else {
		log.Println("✅ Vault already initialized.")
	}
}

func main() {
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to Postgres: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

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

	tbClient, err := tigerbeetle.NewClient(types.ToUint128(0), []string{tbAddress})
	if err != nil {
		log.Fatalf("Error connecting to TigerBeetle: %v", err)
	}
	defer tbClient.Close()
	log.Println("✅ Connected to TigerBeetle")
	bootstrapLedger(tbClient)

	engine := mcp.NewMCPEngine()
	if engine == nil || engine.Client == nil {
		log.Fatal("❌ Fatal error: Could not initialize OpenAI engine")
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	authCtrl := &controllers.AuthController{DB: db, TB: tbClient}
	ledgerCtrl := &controllers.LedgerController{DB: db, TB: tbClient}

	api := r.Group("/api")
	{
		api.POST("/register", authCtrl.Register)
		api.POST("/login", authCtrl.Login)
		api.GET("/users", authCtrl.ListUsers)

		protected := api.Group("/")
		protected.Use(middleware.AuthRequired())
		{
			protected.POST("/logout", authCtrl.Logout)
			protected.POST("/accounts/create", ledgerCtrl.CreateAccount)
			protected.GET("/balance", ledgerCtrl.GetBalance)
			protected.POST("/deposit", ledgerCtrl.Deposit)
			protected.POST("/transfer", ledgerCtrl.Transfer)
			protected.GET("/history", ledgerCtrl.GetHistory)
			protected.DELETE("/accounts/:id", ledgerCtrl.DeleteAccount)
			protected.POST("/chat", mcp.ChatHandler(engine, ledgerCtrl))
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up"})
	})

	log.Println("🚀 API server started on port 8080")
	r.Run("0.0.0.0:8080")
}
