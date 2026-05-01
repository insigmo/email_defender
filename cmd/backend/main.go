package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/insigmo/email_defender/internal/services/db/db_massage"
	"github.com/insigmo/email_defender/internal/services/kafka/kafka_producer"
	"github.com/insigmo/email_defender/internal/services/message_service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/insigmo/email_defender/internal/api/message_handler"
	"github.com/insigmo/email_defender/internal/config"
	"github.com/insigmo/email_defender/internal/logging"
)

func main() {
	conf := config.New()
	logger := setupLogger()

	pool := createDbPool(conf.Postgres)
	defer pool.Close()
	producer, err := kafka_producer.NewKafkaProducer(conf.Kafka, logger)
	if err != nil {
		panic(fmt.Errorf("failed to create kafka producer: %w", err))
	}

	dbMessageManager := db_massage.New(pool)
	msgService := message_service.New(dbMessageManager, producer, logger)
	msgHandler := message_handler.New(msgService, logger)

	http.HandleFunc("/message", msgHandler.GetMessages)
	http.HandleFunc("/send_message", msgHandler.SendMessages)
	http.HandleFunc("/update_message", msgHandler.UpdateMessages)

	logger.Info("Starting server on port 8080")
	err = http.ListenAndServe(":8080", nil)
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

func createDbPool(pgConfig *config.PostgresConfig) *pgxpool.Pool {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, pgConfig.ConnURL)
	if err != nil {
		panic(err)
	}
	err = db.Ping(ctx)
	if err != nil {
		panic(fmt.Sprintf("conUrl: '%s', err %v", pgConfig.ConnURL, err))
	}
	return db
}
