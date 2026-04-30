package message_service

import (
	"context"

	"github.com/insigmo/email_defender/internal/db/db_massage"
)

type MessageService interface {
	Get(ctx context.Context, title string) (string, error)
	Send(ctx context.Context, title, text string) error
	Update(ctx context.Context, title string, isChecked bool, hash string) error
}

type Messanger struct {
	db db_massage.MessageManager
}

func New(db db_massage.MessageManager) MessageService {
	return &Messanger{
		db: db,
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
	return m.db.SendMessage(ctx, title, text)
}

func (m *Messanger) Update(ctx context.Context, title string, isChecked bool, hash string) error {
	return m.db.UpdateMessage(ctx, title, isChecked, hash)
}
