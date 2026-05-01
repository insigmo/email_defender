package message_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/insigmo/email_defender/internal/model"
	"github.com/insigmo/email_defender/internal/services/message_service"
	"go.uber.org/zap"
)

type MessageHandler interface {
	GetMessages(w http.ResponseWriter, r *http.Request)
	SendMessages(w http.ResponseWriter, r *http.Request)
	UpdateMessages(w http.ResponseWriter, r *http.Request)
}

type MessageHandlerImp struct {
	service message_service.MessageService
	logger  *zap.Logger
}

func New(service message_service.MessageService, logger *zap.Logger) MessageHandler {
	return &MessageHandlerImp{
		service: service,
		logger:  logger,
	}
}

func (m *MessageHandlerImp) GetMessages(w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	title := r.URL.Query().Get("title")
	message, err := m.service.Get(ctx, title)
	if err != nil {
		m.logger.Warn("Failed to get message", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}
	fmt.Fprintln(w, message)
}

func (m *MessageHandlerImp) SendMessages(w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		m.logger.Warn("Failed to read body", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}

	var body model.Message

	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		m.logger.Warn("Failed to unmarshal body", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}
	err = m.service.Send(ctx, body.Title, body.Text)
	if err != nil {
		m.logger.Warn("Failed to save message", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}

	fmt.Fprintln(w, "Message successfully sent")
}

func (m *MessageHandlerImp) UpdateMessages(w http.ResponseWriter, r *http.Request) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		m.logger.Warn("Failed to read body", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}

	var body model.Message

	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		m.logger.Warn("Failed to unmarshal body", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}
	err = m.service.Update(ctx, body.Title, body.IsChecked, body.Hash)
	if err != nil {
		m.logger.Warn("Failed to save message", zap.Error(err))
		fmt.Fprintln(w, "Something went wrong")
		return
	}

	fmt.Fprintln(w, "Message successfully sent")
}
