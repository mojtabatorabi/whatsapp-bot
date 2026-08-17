package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mojtabatorabi/whatsapp-bot/internal/service"
)

type WhatsAppHandler struct {
	service *service.ChatService
}

type WhatsAppMessage struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

type WhatsAppResponse struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

func NewWhatsAppHandler(
	service *service.ChatService,
) *WhatsAppHandler {

	return &WhatsAppHandler{
		service: service,
	}
}

func (h *WhatsAppHandler) Webhook(
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

	log.Printf(
		"WhatsApp message from=%s message=%s",
		msg.From,
		msg.Message,
	)

	reply, err := h.service.SendMessage(
		r.Context(),
		msg.From,
		msg.Message,
	)

	if err != nil {

		log.Println(
			"WhatsApp chat error:",
			err,
		)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)

		return
	}

	response := WhatsAppResponse{
		To:      msg.From,
		Message: reply,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}
