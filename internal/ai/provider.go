package ai

import (
	"context"

	"github.com/mojtabatorabi/whatsapp-bot/internal/model"
)

type Provider interface {
	Chat(
		ctx context.Context,
		messages []model.Message,
	) (string, error)
}
