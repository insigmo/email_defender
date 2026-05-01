package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/insigmo/email_defender/internal/services/http_client"
	"github.com/insigmo/email_defender/internal/services/kafka/kafka_consumer"
	"github.com/insigmo/email_defender/internal/services/message_handler_service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/insigmo/email_defender/internal/config"
	"github.com/insigmo/email_defender/internal/logging"
)

func main() {
	conf := config.New()
	logger := setupLogger()
	client := http_client.NewHttpClient(conf.Service.BaseUrl, logger)

	msgHandler := message_handler_service.New(client, logger)
	consumer, err := kafka_consumer.NewKafkaConsumer(conf.Kafka, msgHandler, logger)
	if err != nil {
		panic(fmt.Errorf("failed to create kafka producer: %w", err))
	}
	consumer.StartConsume()

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
	return db
}
