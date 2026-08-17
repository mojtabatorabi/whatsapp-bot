package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mojtabatorabi/whatsapp-bot/internal/service"
)

type ChatHandler struct {
	service *service.ChatService
}

type ChatRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Message string `json:"message"`
}

func NewChatHandler(
	service *service.ChatService,
) *ChatHandler {

	return &ChatHandler{
		service: service,
	}
}

func (h *ChatHandler) Chat(
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

	reply, err := h.service.SendMessage(
		r.Context(),
		req.User,
		req.Message,
	)

	if err != nil {

		log.Println(
			"chat error:",
			err,
		)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(
		ChatResponse{
			Message: reply,
		},
	)
}
