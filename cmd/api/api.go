package main

import (
	"chatX/internal/mailer"
	service "chatX/internal/usecase"
	"chatX/internal/ws"
	"net/http"
	"time"

	"chatX/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config   config
	services service.Services
	ws       *ws.Hub
	logger   zap.SugaredLogger
	mailer   mailer.Client
}

type config struct {
	Addr    string
	DB      DBConfig
	ENV     string
	Upgrade websocket.Upgrader
	mail    MailConfig
	apiURL  string
}

type DBConfig struct {
	Addr         string
	Host         string
	User         string
	Password     string
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	MaxIdletime  string
}

type MailConfig struct {
	mailtrap  mailtrapConfig
	fromEmail string
	exp       time.Duration
}

type mailtrapConfig struct {
	host     string
	port     int
	username string
	password string
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		switch origin {
		case "http://localhost:8080", "http://127.0.0.1:8080":
			return true
		default:
			return false
		}
	},
}

func (app *application) mount() *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheck)
		r.Get("/ws", app.handleWebSocket)

		docsURL := "/api/v1/swagger/doc.json"
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/chats", func(r chi.Router) {
			r.Post("/", app.CreatechatHandler)
			r.Get("/", app.GetUserChatsHandler)
			r.Delete("/{chat_id}", app.DeleteChatHandler)
			r.Get("/{chat_id}/messages", app.GetMessagesHandler)
		})

		r.Route("/groups", func(r chi.Router) {
			r.Post("/", app.CreateGroupHandler)
			r.Patch("/{chat_id}", app.UpdateChatHandler)
			r.Post("/{chat_id}/members", app.AddMemberHandler)
			r.Get("/{chat_id}/members", app.GetMembersHandler)
			r.Delete("/{chat_id}/{user_id}/member", app.DeleteMemberHandler)
		})

		r.Route("/messages", func(r chi.Router) {
			r.Post("/", app.MessageCreateHandler)
			r.Patch("/{id}", app.MessageUpdateHandler)
			r.Delete("/{id}", app.MessageDeleteHandler)
			r.Patch("/chats/{chat_id}/read", app.MarkAsReadHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Get("/", app.GetUserHandler)
			r.Get("/activate/{token}", app.activateUserHandler)
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/authentication", func(r chi.Router) {
				r.Post("/", app.registerUserHandler)
			})
		})
	})

	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)

	return r

}

func (app *application) run(Addr string, handler *chi.Mux) error {
	host := app.config.apiURL
	if host == "" {
		host = "localhost" + Addr
	}

	// Docs
	docs.SwaggerInfo.Version = Version
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/api/v1"

	srv := &http.Server{
		Addr:         Addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srv.ListenAndServe()
}
