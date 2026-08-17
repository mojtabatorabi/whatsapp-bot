package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serverPort = ":8080"

	ollamaURL   = "http://localhost:11434/api/chat"
	ollamaModel = "qwen2.5-coder:7b"

	systemPrompt = "You are a helpful WhatsApp assistant. Always answer in Persian."

	defaultDatabaseURL = "postgres://whatsapp_bot:StrongPassword123@localhost:5432/whatsapp_bot"
)

var db *pgxpool.Pool

// =========================
// Message
// =========================

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// =========================
// WhatsApp
// =========================

type WhatsAppMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

type WhatsAppResponse struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

type ChatRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Message string `json:"message"`
}

// =========================
// Ollama
// =========================

type OllamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type OllamaResponse struct {
	Message Message `json:"message"`
}

// =========================
// Main
// =========================

func main() {

	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	var err error

	db, err = pgxpool.New(ctx, databaseURL)

	if err != nil {
		log.Fatal("Database connection error:", err)
	}

	defer db.Close()

	err = db.Ping(ctx)

	if err != nil {
		log.Fatal("Database ping error:", err)
	}

	fmt.Println("=================================")
	fmt.Println("WhatsApp AI Bot")
	fmt.Println("=================================")
	fmt.Println("Server :", serverPort)
	fmt.Println("Ollama :", ollamaURL)
	fmt.Println("Model  :", ollamaModel)
	fmt.Println("Database: PostgreSQL")
	fmt.Println("=================================")

	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir("./web"),
			),
		),
	)

	http.HandleFunc(
		"/",
		indexHandler,
	)

	http.HandleFunc(
		"/api/chat",
		chatHandler,
	)

	http.HandleFunc(
		"/health",
		healthHandler,
	)

	http.HandleFunc(
		"/webhook/whatsapp",
		whatsappWebhook,
	)

	http.HandleFunc(
		"/simulate/send",
		simulateSend,
	)

	http.HandleFunc(
		"/memory",
		memoryHandler,
	)

	http.HandleFunc(
		"/memory/clear",
		clearMemoryHandler,
	)

	err = http.ListenAndServe(
		serverPort,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}
}

//=============================
//chathandeler_ui
//=============================

func indexHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(
		w,
		r,
		"./web/index.html",
	)
}

func chatHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	var req ChatRequest

	err := json.NewDecoder(
		r.Body,
	).Decode(&req)

	if err != nil {

		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)

		return
	}

	if req.User == "" {
		req.User = "web-user"
	}

	if req.Message == "" {

		http.Error(
			w,
			"Message is required",
			http.StatusBadRequest,
		)

		return
	}

	ctx := r.Context()

	// -----------------------------
	// Load conversation
	// -----------------------------

	messages, err := getMessages(
		ctx,
		req.User,
	)

	if err != nil {

		log.Println(
			"Database error:",
			err,
		)

		http.Error(
			w,
			"Database error: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// -----------------------------
	// System prompt
	// -----------------------------

	if len(messages) == 0 {

		err = saveMessage(
			ctx,
			req.User,
			"system",
			systemPrompt,
		)

		if err != nil {

			http.Error(
				w,
				"Database error: "+err.Error(),
				http.StatusInternalServerError,
			)

			return
		}

		messages = append(
			messages,
			Message{
				Role:    "system",
				Content: systemPrompt,
			},
		)
	}

	// -----------------------------
	// User message
	// -----------------------------

	err = saveMessage(
		ctx,
		req.User,
		"user",
		req.Message,
	)

	if err != nil {

		http.Error(
			w,
			"Database error: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	messages = append(
		messages,
		Message{
			Role:    "user",
			Content: req.Message,
		},
	)

	// -----------------------------
	// Ollama
	// -----------------------------

	reply, err := askOllama(
		messages,
	)

	if err != nil {

		log.Println(
			"Ollama error:",
			err,
		)

		http.Error(
			w,
			"Ollama error: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// -----------------------------
	// Save AI response
	// -----------------------------

	err = saveMessage(
		ctx,
		req.User,
		"assistant",
		reply,
	)

	if err != nil {

		http.Error(
			w,
			"Database error: "+err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// -----------------------------
	// Response
	// -----------------------------

	response := ChatResponse{
		Message: reply,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(
		w,
	).Encode(response)
}

// =========================
// Health
// =========================

func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	ctx := r.Context()

	err := db.Ping(ctx)

	databaseStatus := "ok"

	if err != nil {
		databaseStatus = "error"
	}

	response := map[string]interface{}{
		"status":   "ok",
		"service":  "whatsapp-ai-bot",
		"database": databaseStatus,
		"ollama":   ollamaURL,
		"model":    ollamaModel,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}

// =========================
// WhatsApp Webhook
// =========================

func whatsappWebhook(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	var msg WhatsAppMessage

	err := json.NewDecoder(
		r.Body,
	).Decode(&msg)

	if err != nil {

		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)

		return
	}

	if msg.From == "" || msg.Message == "" {

		http.Error(
			w,
			"from and message are required",
			http.StatusBadRequest,
		)

		return
	}

	fmt.Println()
	fmt.Println("=================================")
	fmt.Println("Incoming message")
	fmt.Println("From   :", msg.From)
	fmt.Println("Message:", msg.Message)

	ctx := r.Context()

	// --------------------------------
	// Check conversation
	// --------------------------------

	messages, err := getMessages(
		ctx,
		msg.From,
	)

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		log.Println(err)

		return
	}

	// --------------------------------
	// Add system prompt
	// --------------------------------

	if len(messages) == 0 {

		err = saveMessage(
			ctx,
			msg.From,
			"system",
			systemPrompt,
		)

		if err != nil {

			http.Error(
				w,
				"Database error",
				http.StatusInternalServerError,
			)

			return
		}

		messages = append(
			messages,
			Message{
				Role:    "system",
				Content: systemPrompt,
			},
		)
	}

	// --------------------------------
	// Save user message
	// --------------------------------

	err = saveMessage(
		ctx,
		msg.From,
		"user",
		msg.Message,
	)

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		return
	}

	messages = append(
		messages,
		Message{
			Role:    "user",
			Content: msg.Message,
		},
	)

	// --------------------------------
	// Ask Ollama
	// --------------------------------

	reply, err := askOllama(
		messages,
	)

	if err != nil {

		log.Println(
			"Ollama error:",
			err,
		)

		http.Error(
			w,
			"Ollama error",
			http.StatusInternalServerError,
		)

		return
	}

	// --------------------------------
	// Save assistant response
	// --------------------------------

	err = saveMessage(
		ctx,
		msg.From,
		"assistant",
		reply,
	)

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		return
	}

	fmt.Println("Reply:", reply)
	fmt.Println("=================================")

	response := WhatsAppResponse{
		To:      msg.From,
		Message: reply,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		response,
	)
}

// =========================
// Save Message
// =========================

func saveMessage(
	ctx context.Context,
	userID string,
	role string,
	content string,
) error {

	_, err := db.Exec(
		ctx,
		`
		INSERT INTO messages
		(user_id, role, content)
		VALUES ($1, $2, $3)
		`,
		userID,
		role,
		content,
	)

	return err
}

// =========================
// Get Messages
// =========================

func getMessages(
	ctx context.Context,
	userID string,
) ([]Message, error) {

	rows, err := db.Query(
		ctx,
		`
		SELECT role, content
		FROM messages
		WHERE user_id = $1
		ORDER BY created_at ASC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []Message

	for rows.Next() {

		var message Message

		err := rows.Scan(
			&message.Role,
			&message.Content,
		)

		if err != nil {
			return nil, err
		}

		messages = append(
			messages,
			message,
		)
	}

	return messages, rows.Err()
}

// =========================
// Memory API
// =========================

func memoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	userID := r.URL.Query().Get("user")

	if userID == "" {

		http.Error(
			w,
			"Missing user parameter",
			http.StatusBadRequest,
		)

		return
	}

	messages, err := getMessages(
		r.Context(),
		userID,
	)

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"user":     userID,
			"messages": messages,
		},
	)
}

// =========================
// Clear Memory
// =========================

func clearMemoryHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodDelete {

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	userID := r.URL.Query().Get("user")

	if userID == "" {

		http.Error(
			w,
			"Missing user parameter",
			http.StatusBadRequest,
		)

		return
	}

	_, err := db.Exec(
		r.Context(),
		`
		DELETE FROM messages
		WHERE user_id = $1
		`,
		userID,
	)

	if err != nil {

		http.Error(
			w,
			"Database error",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"success": true,
			"user":    userID,
		},
	)
}

// =========================
// Ollama
// =========================

func askOllama(
	messages []Message,
) (string, error) {

	requestBody := OllamaRequest{
		Model:    ollamaModel,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(
		requestBody,
	)

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		ollamaURL,
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return "",
			fmt.Errorf(
				"Ollama returned %s",
				resp.Status,
			)
	}

	var result OllamaResponse

	err = json.NewDecoder(
		resp.Body,
	).Decode(&result)

	if err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

// =========================
// Mock WhatsApp
// =========================

func simulateSend(
	w http.ResponseWriter,
	r *http.Request,
) {

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"Method not allowed",
			http.StatusMethodNotAllowed,
		)

		return
	}

	var msg WhatsAppMessage

	err := json.NewDecoder(
		r.Body,
	).Decode(&msg)

	if err != nil {

		http.Error(
			w,
			"Invalid JSON",
			http.StatusBadRequest,
		)

		return
	}

	body, err := json.Marshal(msg)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	resp, err := http.Post(
		"http://localhost:8080/webhook/whatsapp",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	defer resp.Body.Close()

	var result WhatsAppResponse

	err = json.NewDecoder(
		resp.Body,
	).Decode(&result)

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(result)
}

// =========================
// Keep pgx import used
// =========================

var _ = pgx.ErrNoRows
