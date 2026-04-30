package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/insigmo/email_defender/internal/api/message_handler"
	"github.com/insigmo/email_defender/internal/db/db_massage"
	"github.com/insigmo/email_defender/internal/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	pool := createDbPool()
	defer pool.Close()
	logger := setupLogger()

	msgHandler := getMessageHandler(pool, logger)
	http.HandleFunc("/message", msgHandler.GetMessages)
	http.HandleFunc("/send_message", msgHandler.SendMessages)
	logger.Info("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func setupLogger() *zap.Logger {
	logger, err := logging.New()
	if err != nil {
		panic(err)
	}
	return logger
}

func createDbPool() *pgxpool.Pool {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	if user == "" || password == "" || dbName == "" {
		panic(errors.New("credentials not found"))
	}
	dbConn := fmt.Sprintf(`postgres://%s:%s@localhost:5432/%s`, user, password, dbName)
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbConn)
	if err != nil {
		panic(err)
	}
	return db
}

func getMessageHandler(db *pgxpool.Pool, logger *zap.Logger) message_handler.MessageHandler {
	messageManager := db_massage.New(db)
	return message_handler.New(messageManager, logger)
}
