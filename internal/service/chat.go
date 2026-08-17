package service

import (
	"context"

	"github.com/mojtabatorabi/whatsapp-bot/internal/ai"
	"github.com/mojtabatorabi/whatsapp-bot/internal/model"
	"github.com/mojtabatorabi/whatsapp-bot/internal/repository"
)

type ChatService struct {
	messages repository.MessageRepository
	ai       ai.Provider
	prompt   string
}

func NewChatService(
	messages repository.MessageRepository,
	aiProvider ai.Provider,
	systemPrompt string,
) *ChatService {

	return &ChatService{
		messages: messages,
		ai:       aiProvider,
		prompt:   systemPrompt,
	}
}

func (s *ChatService) SendMessage(
	ctx context.Context,
	userID string,
	text string,
) (string, error) {

	messages, err := s.messages.GetMessages(
		ctx,
		userID,
	)

	if err != nil {
		return "", err
	}

	if len(messages) == 0 {

		systemMessage := model.Message{
			Role:    "system",
			Content: s.prompt,
		}

		err = s.messages.Save(
			ctx,
			userID,
			systemMessage,
		)

		if err != nil {
			return "", err
		}

		messages = append(
			messages,
			systemMessage,
		)
	}

	userMessage := model.Message{
		Role:    "user",
		Content: text,
	}

	err = s.messages.Save(
		ctx,
		userID,
		userMessage,
	)

	if err != nil {
		return "", err
	}

	messages = append(
		messages,
		userMessage,
	)

	reply, err := s.ai.Chat(
		ctx,
		messages,
	)

	if err != nil {
		return "", err
	}

	assistantMessage := model.Message{
		Role:    "assistant",
		Content: reply,
	}

	err = s.messages.Save(
		ctx,
		userID,
		assistantMessage,
	)

	if err != nil {
		return "", err
	}

	return reply, nil
}

func (s *ChatService) History(
	ctx context.Context,
	userID string,
) ([]model.Message, error) {

	return s.messages.GetMessages(
		ctx,
		userID,
	)
}

func (s *ChatService) Clear(
	ctx context.Context,
	userID string,
) error {

	return s.messages.Clear(
		ctx,
		userID,
	)
}
