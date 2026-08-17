package config

import "os"

type Config struct {
	ServerPort   string
	DatabaseURL  string
	OllamaURL    string
	OllamaModel  string
	SystemPrompt string
}

func Load() Config {

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		databaseURL = "postgres://whatsapp_bot:StrongPassword123@localhost:5432/whatsapp_bot"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")

	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")

	if ollamaModel == "" {
		ollamaModel = "qwen2.5-coder:7b"
	}

	serverPort := os.Getenv("SERVER_PORT")

	if serverPort == "" {
		serverPort = ":8080"
	}

	return Config{
		ServerPort:   serverPort,
		DatabaseURL:  databaseURL,
		OllamaURL:    ollamaURL,
		OllamaModel:  ollamaModel,
		SystemPrompt: "You are a helpful WhatsApp assistant. Always answer in Persian.",
	}
}
