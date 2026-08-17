package repository

import (
	"context"

	"github.com/mojtabatorabi/whatsapp-bot/internal/model"
)

type MessageRepository interface {
	Save(
		ctx context.Context,
		userID string,
		message model.Message,
	) error

	GetMessages(
		ctx context.Context,
		userID string,
	) ([]model.Message, error)

	Clear(
		ctx context.Context,
		userID string,
	) error
}
