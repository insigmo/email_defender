package db_massage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getMessageQuery    = `select text from message where title = $1`
	saveMessageQuery   = `insert into message (title, text, checked) values ($1, $2, $3)`
	updateMessageQuery = `update message set checked = $2, hash = $3 where title = $1`
)

type MessageManager interface {
	GetMessage(ctx context.Context, title string) (string, error)
	SendMessage(ctx context.Context, title string, text string) error
	UpdateMessage(ctx context.Context, title string, isChecked bool, hash string) error
}

type MessageManagerImp struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) MessageManager {
	return &MessageManagerImp{db: db}
}

func (m *MessageManagerImp) GetMessage(ctx context.Context, title string) (string, error) {
	row, err := m.db.Query(ctx, getMessageQuery, title)
	if err != nil {
		return "", fmt.Errorf(`query: %s, err: %w`, getMessageQuery, err)
	}
	var res string
	row.Next()
	err = row.Scan(&res)
	if err != nil {
		return "", fmt.Errorf(`scan: %s, err: %w`, getMessageQuery, err)
	}

	return res, nil
}

func (m *MessageManagerImp) SendMessage(ctx context.Context, title string, text string) error {
	_, err := m.db.Exec(ctx, saveMessageQuery, title, text, false)
	if err != nil {
		return fmt.Errorf(`save: %s, err: %w`, getMessageQuery, err)
	}
	return nil
}

func (m *MessageManagerImp) UpdateMessage(ctx context.Context, title string, isChecked bool, hash string) error {
	_, err := m.db.Exec(ctx, updateMessageQuery, title, isChecked, hash)
	if err != nil {
		return fmt.Errorf(`update: %s, err: %w`, updateMessageQuery, err)
	}
	return nil
}
