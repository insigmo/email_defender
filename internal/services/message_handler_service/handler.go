package message_handler_service

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/insigmo/email_defender/internal/config"
	"github.com/insigmo/email_defender/internal/model"
	"github.com/insigmo/email_defender/internal/services/http_client"
	"go.uber.org/zap"
)

type MessageHandlerService interface {
	Get(ctx context.Context, title string) (*model.Message, error)
	Update(ctx context.Context, msg *model.Message) error
}

type Handler struct {
	conf       *config.KafkaConfig
	httpClient http_client.HttpClient
	logger     *zap.Logger
}

func New(
	httpClient http_client.HttpClient,
	conf *config.KafkaConfig,
	logger *zap.Logger,
) MessageHandlerService {
	return &Handler{
		httpClient: httpClient,
		conf:       conf,
		logger:     logger,
	}
}

func (h *Handler) Get(ctx context.Context, title string) (*model.Message, error) {
	msg, err := h.httpClient.GetMessage(ctx, title)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (h *Handler) Update(ctx context.Context, msg *model.Message) error {
	msg.Hash = fmt.Sprintf("%s", md5.Sum([]byte(msg.Text)))
	msg.IsChecked = true

	err := h.httpClient.UpdateMessage(ctx, msg)
	if err != nil {
		h.logger.Error("Failed to update message", zap.Error(err))
		return err
	}
	return nil
}
