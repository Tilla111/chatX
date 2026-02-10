package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"chatX/internal/ws"
	"fmt"
	"net/http"
	"time"

	"chatX/docs"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config   config
	strore   store.Storage
	services service.Services
	ws       *ws.Hub
	logger   zap.SugaredLogger
}

type config struct {
	Addr    string
	DB      DBConfig
	ENV     string
	Upgrade websocket.Upgrader
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

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (app *application) mount() *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheck)
		r.Get("/ws", app.handleWebSocket)

		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.Addr)
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
			r.Get("/{chat_id}/members", app.GetMembersHandler)
			r.Delete("/{chat_id}/{user_id}/member", app.DeleteMemberHandler)
		})

		r.Route("/messages", func(r chi.Router) {
			r.Post("/", app.MessageCreateHandler)
			r.Patch("/{chat_id}", app.MarkAsReadHandler)
			r.Patch("/{id}", app.MessageUpdateHandler)
			r.Delete("/{id}", app.MessageDeleteHandler)
		})

		r.Route("/users", func(r chi.Router) {

			r.Route("/{user_id}", func(r chi.Router) {
				r.Get("/", app.GetUserHandler)
			})
		})
	})

	fs := http.FileServer(http.Dir("./web"))
	r.Handle("/*", fs)

	return r

}

func (app *application) run(Addr string, handler *chi.Mux) error {
	// Docs
	docs.SwaggerInfo.Version = Version
	docs.SwaggerInfo.Host = app.config.apiURL
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
