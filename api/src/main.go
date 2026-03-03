package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Configurar conexión a PostgreSQL
	dsn := os.Getenv("DB_URL")
	_, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error conectando a Postgres: %v", err)
	}
	log.Println("Conectado a PostgreSQL")

	// 2. Configurar conexión a TigerBeetle
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

	log.Printf("Intentando conectar a TigerBeetle en IP: %s", tbAddress)

	tbClient, err := tigerbeetle.NewClient(types.ToUint128(0), []string{tbAddress})
	if err != nil {
		log.Fatalf("Error conectando a TigerBeetle: %v", err)
	}
	defer tbClient.Close()
	log.Println("Conectado a TigerBeetle")

	// 3. Configurar Gin (Servidor Web)
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"db":     "connected",
			"ledger": "connected",
		})
	})

	log.Println("Servidor API iniciado en el puerto 8080")
	r.Run(":8080")
}
