# whatsapp-bot
                    ┌─────────────────────┐
                    │       Users         │
                    └──────────┬──────────┘
                               │
                 ┌─────────────┴─────────────┐
                 │                           │
             Web Chat                   WhatsApp
                 │                           │
                 └─────────────┬─────────────┘
                               │
                         ┌─────▼─────┐
                         │  API/HTTP  │
                         │    Go      │
                         └─────┬─────┘
                               │
                    ┌──────────▼──────────┐
                    │   Application       │
                    │                     │
                    │ Chat Service        │
                    │ Conversation        │
                    │ Message Service     │
                    └───────┬──────┬──────┘
                            │      │
              ┌─────────────┘      └─────────────┐
              ▼                                   ▼
       ┌──────────────┐                    ┌──────────────┐
       │ PostgreSQL   │                    │    Ollama    │
       │              │                    │              │
       │ users        │                    │ Qwen/Llama   │
       │ conversations│                    │              │
       │ messages     │                    └──────────────┘
       └──────────────┘

                         INTERNET
                            │
                 ┌──────────┴──────────┐
                 │                     │
             WhatsApp                Web
                 │                     │
                 ▼                     ▼
          ┌─────────────────────────────────┐
          │            Go API               │
          │                                 │
          │  Handler                        │
          │      │                          │
          │  Application Service            │
          │      │                          │
          │  ┌───┴───────────────┐          │
          │  │                   │          │
          │ Repository         AI Provider  │
          │  │                   │          │
          └──┼───────────────────┼──────────┘
             │                   │
             ▼                   ▼
        PostgreSQL            Ollama
             │                   │
             │                   ▼
             │                Qwen
             │
             ▼
       Conversation
       Message
       User
