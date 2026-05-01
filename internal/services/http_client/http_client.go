package http_client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bytedance/sonic"
	"github.com/insigmo/email_defender/internal/model"
	"go.uber.org/zap"
)

const (
	contentType = "application/json"
)

type HttpClient interface {
	GetMessage(ctx context.Context, title string) (*model.Message, error)
	UpdateMessage(ctx context.Context, message *model.Message) error
}

type HttpClientImpl struct {
	client  *http.Client
	baseUrl string
	logger  *zap.Logger
}

func NewHttpClient(baseUrl string, logger *zap.Logger) HttpClient {
	client := &http.Client{}

	return &HttpClientImpl{
		client:  client,
		baseUrl: baseUrl,
		logger:  logger,
	}
}

func (h *HttpClientImpl) GetMessage(ctx context.Context, title string) (*model.Message, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	params := url.Values{}
	params.Set("title", title)

	fetchUrl := fmt.Sprintf("%s%s?%s", h.baseUrl, "/message", params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchUrl, nil)

	resp, err := h.client.Do(req)
	if err != nil {
		h.logger.Error("failed to update http client", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()
	var message model.Message
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read http response body: %w", err)
	}

	err = sonic.Unmarshal(b, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal http response body: %w", err)
	}
	return &message, nil
}

func (h *HttpClientImpl) UpdateMessage(ctx context.Context, message *model.Message) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	b, _ := sonic.Marshal(message)

	updateUrl := h.baseUrl + "/update_message"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateUrl, bytes.NewBuffer(b))
	if err != nil || req == nil {
		h.logger.Error("failed to update http client", zap.Error(err))
		return err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := h.client.Do(req)

	if err != nil {
		h.logger.Error("failed to update http client", zap.Error(err))

		return err
	}
	err = resp.Body.Close()
	if err != nil {
		return err
	}
	return nil
}
