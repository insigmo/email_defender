package kafka_consumer

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
	"github.com/insigmo/email_defender/internal/services/message_handler_service"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type Consumer interface {
	StartConsume()
	Consume(ctx context.Context, conf *config.KafkaConfig) error
}
type ConsumerImp struct {
	client     *kgo.Client
	msgHandler message_handler_service.MessageHandlerService
	logger     *zap.Logger
	conf       *config.KafkaConfig
}

func NewKafkaConsumer(conf *config.KafkaConfig, msgHandler message_handler_service.MessageHandlerService, logger *zap.Logger) (Consumer, error) {
	client, err := createConsumerClient(conf)
	if err != nil {
		return nil, err
	}
	return &ConsumerImp{
		client:     client,
		msgHandler: msgHandler,
		logger:     logger,
		conf:       conf,
	}, nil
}

func (c *ConsumerImp) StartConsume() {
	client, err := createConsumerClient(c.conf)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	defer stop()

	defer func() {
		client.Close()
		c.logger.Info("consumer closed")
	}()

	if err := c.Consume(ctx, c.conf); err != nil && !errors.Is(err, context.Canceled) {
		c.logger.Error("consume error", zap.Error(err))
		os.Exit(1)
	}
}

func (c *ConsumerImp) Consume(ctx context.Context, conf *config.KafkaConfig) error {
	c.logger.Info("consumer started",
		zap.String("group", conf.GroupID),
		zap.String("topic", conf.Topic),
	)

	for {
		// PollFetches блокируется до появления записей или отмены контекста.
		// С BlockRebalanceOnPoll ребаланс заблокирован на время этого вызова.
		fetches := c.client.PollFetches(ctx)

		// Разрешаем ребаланс сразу после получения батча
		c.client.AllowRebalance()

		// Проверяем ошибки poll (non-retriable ошибки возвращаются наружу)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fe := range errs {
				// Ошибка контекста — нормальное завершение
				if errors.Is(fe.Err, context.Canceled) {
					c.logger.Info("context cancelled, stopping consumer")
					return nil
				}
				// Если ошибка топика/партиции — логируем и продолжаем
				c.logger.Error("fetch error",
					zap.String("topic", fe.Topic),
					zap.Int32("partition", fe.Partition),
					zap.Error(fe.Err),
				)
			}
			// Если были только partial ошибки — записи могут всё равно быть
			if fetches.Empty() {
				continue
			}
		}

		if fetches.Empty() {
			// Контекст отменён без ошибок
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		// Обрабатываем записи по партициям
		var processErr error
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			if processErr != nil {
				return // прерываем обход при критической ошибке
			}
			for _, record := range p.Records {
				if err := c.ProcessRecord(ctx, record); err != nil {
					if errors.Is(err, context.Canceled) {
						processErr = err
						return
					}
					// Логируем ошибку обработки, но не останавливаем потребление.
					// В production здесь может быть dead-letter queue или retry логика.
					c.logger.Error("failed to process record",
						zap.String("topic", record.Topic),
						zap.Int32("partition", record.Partition),
						zap.Int64("offset", record.Offset),
						zap.Error(err),
					)
				}
			}
		})

		if processErr != nil {
			return processErr
		}

		// Коммитим офсеты после успешной обработки всего батча.
		// CommitUncommittedOffsets коммитит офсеты всех записей,
		// полученных с момента последнего коммита.
		if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			// Ошибка коммита не критична — Kafka повторит доставку при рестарте.
			// Но нужно логировать для observability.
			c.logger.Error("failed to commit offsets", zap.Error(err))
		} else {
			c.logger.Debug("offsets committed")
		}
	}
}

// ProcessRecord — бизнес-логика обработки одной записи.
// Замените тело этой функции на свою логику.
func (c *ConsumerImp) ProcessRecord(ctx context.Context, r *kgo.Record) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.logger.Info("received record",
		zap.String("topic", r.Topic),
		zap.Int32("partition", r.Partition),
		zap.Int64("offset", r.Offset),
		zap.ByteString("key", r.Key),
		zap.ByteString("value", r.Value),
		zap.String("timestamp", r.Timestamp.Format(time.RFC3339)),
		zap.String("headers", fmt.Sprintf("%v", r.Headers)),
	)
	body, err := c.decode(r)
	if err != nil {
		return err
	}

	title, ok := body["title"]
	if !ok {
		return errors.New("missing title")
	}
	err = c.msgHandler.Update(ctx, title)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConsumerImp) decode(r *kgo.Record) (map[string]string, error) {
	var body map[string]string
	err := sonic.Unmarshal(r.Value, &body)
	if err != nil {
		return body, err
	}
	return body, nil
}

func createConsumerClient(conf *config.KafkaConfig) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(conf.Brokers),
		kgo.ConsumerGroup(conf.GroupID),
		kgo.ConsumeTopics(conf.Topic),

		// Начать с самого начала, если офсет ещё не закоммичен для группы.
		// Используйте kgo.NewOffsetEnd() для чтения только новых сообщений.
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),

		// BlockRebalanceOnPoll гарантирует, что ребаланс не произойдёт
		// между PollFetches и AllowRebalance.
		// Это самый простой и безопасный паттерн для ручного коммита.
		kgo.BlockRebalanceOnPoll(),

		// Отключаем авто-коммит — будем коммитить вручную после обработки батча.
		kgo.DisableAutoCommit(),

		// Тайм-аут ребаланса: сессия консюмера должна успеть обработать батч.
		// Если обработка занимает больше — увеличить это значение.
		kgo.RebalanceTimeout(30*time.Second),

		// FetchMaxBytes и FetchMaxPartitionBytes ограничивают размер батча
		kgo.FetchMaxBytes(5<<20),          // 5 MB на весь fetch
		kgo.FetchMaxPartitionBytes(1<<20), // 1 MB на партицию

		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelWarn, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed creating kafka producer: %w", err)
	}
	return client, nil
}
