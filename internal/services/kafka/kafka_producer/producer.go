package kafka_producer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/insigmo/email_defender/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

var (
	jsonContentTypeKey = "content-type"
	jsonContentType    = []byte("application/json")
)

type Producer interface {
	Publish(message map[string]string) error
}

type ProducerImp struct {
	logger *zap.Logger
	client *kgo.Client
	conf   *config.KafkaConfig
}

func NewKafkaProducer(conf *config.KafkaConfig, logger *zap.Logger) (Producer, error) {
	client, err := createProducerClient(conf)
	if err != nil {
		return nil, err
	}

	return &ProducerImp{
		client: client,
		logger: logger,
		conf:   conf,
	}, nil
}

func (p *ProducerImp) Publish(message map[string]string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer cancelFunc()
	title, ok := message["title"]
	if !ok {
		return errors.New("title not found in message")
	}
	record, err := p.packToKafkaRecord(title, message)
	if err != nil {
		return err
	}

	var promise kgo.FirstErrPromise
	p.client.Produce(ctx, record, promise.Promise())
	// Ждём завершения всех промисов перед выходом
	if err := promise.Err(); err != nil {
		p.logger.Error("async produce batch error", zap.Error(err))
		return fmt.Errorf("produce failed: %w", err)
	}
	return nil
}

func (p *ProducerImp) packToKafkaRecord(eventID string, message map[string]string) (*kgo.Record, error) {
	value, err := sonic.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	return &kgo.Record{
		Topic: p.conf.Topic,
		Key:   []byte(eventID),
		Value: value,
		Headers: []kgo.RecordHeader{
			{Key: jsonContentTypeKey, Value: jsonContentType},
		},
	}, nil
}

func createProducerClient(conf *config.KafkaConfig) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(conf.Brokers...),
		kgo.DefaultProduceTopic(conf.Topic),

		// Идемпотентный продюсер (включён по умолчанию — явно для ясности)
		// kgo.DisableIdempotentWrite(), // раскомментировать только если нужен ack=0/1

		// Linger: накапливать записи до 5ms перед отправкой батча.
		// Увеличивает throughput, добавляет небольшую задержку.
		kgo.ProducerLinger(5*time.Millisecond),

		// Максимум буферизованных записей (back-pressure защита)
		kgo.MaxBufferedRecords(10_000),

		// Таймаут доставки записи
		kgo.ProduceRequestTimeout(10*time.Second),
		kgo.RecordDeliveryTimeout(30*time.Second),

		// Сжатие (snappy — хороший баланс скорости и размера)
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),

		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelInfo, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating kafka producer: %w", err)
	}
	return client, nil
}
