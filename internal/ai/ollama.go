package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mojtabatorabi/whatsapp-bot/internal/model"
)

type OllamaProvider struct {
	URL    string
	Model  string
	Client *http.Client
}

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []model.Message `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaResponse struct {
	Message model.Message `json:"message"`
}

func NewOllamaProvider(
	url string,
	modelName string,
) *OllamaProvider {

	return &OllamaProvider{
		URL:   url,
		Model: modelName,

		Client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (o *OllamaProvider) Chat(
	ctx context.Context,
	messages []model.Message,
) (string, error) {

	requestBody := ollamaRequest{
		Model:    o.Model,
		Messages: messages,
		Stream:   false,
	}

	data, err := json.Marshal(
		requestBody,
	)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.URL+"/api/chat",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := o.Client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return "",
			fmt.Errorf(
				"ollama returned %s",
				resp.Status,
			)
	}

	var result ollamaResponse

	err = json.NewDecoder(
		resp.Body,
	).Decode(&result)

	if err != nil {
		return "", err
	}

	return result.Message.Content, nil
}
