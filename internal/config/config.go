package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Kafka    *KafkaConfig
	Postgres *PostgresConfig
	Service  *ServiceConfig
}

func New() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbConn := os.Getenv("POSTGRES_CONNECTION_URL")
	if dbConn == "" {
		panic(errors.New("please provide a connection string to postgres"))
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		panic(errors.New("please provide a kafka brokers"))
	}

	return &Config{
		Kafka: &KafkaConfig{
			Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
			Topic:   os.Getenv("KAFKA_TOPIC"),
			GroupID: os.Getenv("KAFKA_GROUP_ID"),
		},
		Postgres: &PostgresConfig{
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
			ConnURL:  os.Getenv("POSTGRES_CONNECTION_URL"),
		},
		Service: &ServiceConfig{
			Host:    os.Getenv("SERVICE_HOST"),
			Port:    os.Getenv("SERVICE_PORT"),
			BaseUrl: os.Getenv("SERVICE_BASE_URL"),
		},
	}
}

type KafkaConfig struct {
	Brokers []string `yaml:"KAFKA_BROKERS"`
	Topic   string   `yaml:"KAFKA_TOPIC"`
	GroupID string   `yaml:"KAFKA_GROUP_ID"`
}

type PostgresConfig struct {
	User     string `yaml:"POSTGRES_USER"`
	Password string `yaml:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"POSTGRES_DB_NAME"`
	ConnURL  string `yaml:"POSTGRES_CONNECTION_URL"`
}

type ServiceConfig struct {
	Host    string `yaml:"SERVICE_HOST"`
	Port    string `yaml:"SERVICE_PORT"`
	BaseUrl string `yaml:"SERVICE_BASE_URL"`
}
