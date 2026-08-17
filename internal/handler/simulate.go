package handler

import (
	"encoding/json"
	"net/http"
)

type SimulateHandler struct {
	whatsapp *WhatsAppHandler
}

func NewSimulateHandler(
	whatsapp *WhatsAppHandler,
) *SimulateHandler {

	return &SimulateHandler{
		whatsapp: whatsapp,
	}
}

func (h *SimulateHandler) Send(
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

	// Instead of making an HTTP request to our own server,
	// directly call the WhatsApp webhook handler.

	h.whatsapp.Webhook(w, r)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}
