package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mojtabatorabi/whatsapp-bot/internal/ai"
	"github.com/mojtabatorabi/whatsapp-bot/internal/config"
	"github.com/mojtabatorabi/whatsapp-bot/internal/handler"
	"github.com/mojtabatorabi/whatsapp-bot/internal/repository"
	"github.com/mojtabatorabi/whatsapp-bot/internal/service"
)

func main() {

	cfg := config.Load()

	ctx := context.Background()

	db, err := pgxpool.New(
		ctx,
		cfg.DatabaseURL,
	)

	if err != nil {
		log.Fatal(
			"database connection error:",
			err,
		)
	}

	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal(
			"database ping error:",
			err,
		)
	}

	messageRepository :=
		repository.NewPostgresMessageRepository(
			db,
		)

	ollamaProvider :=
		ai.NewOllamaProvider(
			cfg.OllamaURL,
			cfg.OllamaModel,
		)

	chatService :=
		service.NewChatService(
			messageRepository,
			ollamaProvider,
			cfg.SystemPrompt,
		)

	chatHandler :=
		handler.NewChatHandler(
			chatService,
		)

	whatsappHandler :=
		handler.NewWhatsAppHandler(
			chatService,
		)

	simulateHandler :=
		handler.NewSimulateHandler(
			whatsappHandler,
		)

	mux := http.NewServeMux()

	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(
				http.Dir("./web"),
			),
		),
	)

	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {

			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

			http.ServeFile(
				w,
				r,
				"./web/index.html",
			)
		},
	)

	mux.HandleFunc(
		"/api/chat",
		chatHandler.Chat,
	)

	mux.HandleFunc(
		"/webhook/whatsapp",
		whatsappHandler.Webhook,
	)

	mux.HandleFunc(
		"/simulate/send",
		simulateHandler.Send,
	)

	mux.HandleFunc(
		"/health",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.Write([]byte(
				`{"status":"ok"}`,
			))
		},
	)

	log.Println(
		"WhatsApp AI Bot listening on",
		cfg.ServerPort,
	)

	err = http.ListenAndServe(
		cfg.ServerPort,
		mux,
	)

	if err != nil {
		log.Fatal(err)
	}
}
