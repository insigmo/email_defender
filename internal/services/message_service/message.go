package message_service

import (
	"context"

	"github.com/insigmo/email_defender/internal/services/db/db_massage"
	"github.com/insigmo/email_defender/internal/services/kafka/kafka_producer"
	"go.uber.org/zap"
)

type MessageService interface {
	Get(ctx context.Context, title string) (string, error)
	Send(ctx context.Context, title, text string) error
	Update(ctx context.Context, title string, isChecked bool, hash string) error
}

type Messanger struct {
	db       db_massage.MessageManager
	producer kafka_producer.Producer
	logger   *zap.Logger
}

func New(db db_massage.MessageManager, producer kafka_producer.Producer, logger *zap.Logger) MessageService {
	return &Messanger{
		db:       db,
		producer: producer,
		logger:   logger,
	}
}

func (m *Messanger) Get(ctx context.Context, title string) (string, error) {
	text, err := m.db.GetMessage(ctx, title)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (m *Messanger) Send(ctx context.Context, title, text string) error {
	err := m.db.SendMessage(ctx, title, text)
	if err != nil {
		m.logger.Error("Failed to send message", zap.Error(err))
		return err
	}
	body := map[string]string{
		"title": title,
		"text":  text,
	}
	err = m.producer.Publish(body)
	if err != nil {
		return err
	}
	return nil
}

func (m *Messanger) Update(ctx context.Context, title string, isChecked bool, hash string) error {
	err := m.db.UpdateMessage(ctx, title, isChecked, hash)
	if err != nil {
		m.logger.Error("Failed to update message", zap.Error(err))
		return err
	}
	return nil
}
