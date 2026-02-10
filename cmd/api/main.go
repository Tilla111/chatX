package main

import (
	"chatX/internal/db"
	"chatX/internal/env"
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"chatX/internal/ws"
	"log"
)

const Version = "v1.0.0"

//	@title			ChatX API
//	@description	Api for gophers Chat Project !
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/api/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {

	env, err := env.Load()
	if err != nil {
		log.Fatalf("Error loading env: %v", err)
	}

	cfg := config{
		Addr: ":" + env.Server.Port,
		DB: DBConfig{
			Addr:         env.Database.Addr,
			Host:         env.Database.Host,
			User:         env.Database.User,
			Password:     env.Database.Password,
			Name:         env.Database.Name,
			MaxIdleConns: env.Database.MaxIdleConns,
			MaxOpenConns: env.Database.MaxOpenConns,
			MaxIdletime:  env.Database.MaxIdletime,
		},
		ENV:     env.App.ENV,
		Upgrade: Upgrader,
	}

	db, err := db.NewPostgres(
		cfg.DB.Addr,
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdletime,
	)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Connected to database successfully")

	hub := ws.NewHub()
	go hub.Run()

	storage := store.NewStorage(db)
	services := service.NewServices(storage)

	app := &application{
		config:   cfg,
		strore:   *storage,
		services: *services,
		ws:       hub,
	}

	handler := app.mount()

	log.Println("Starting server on", cfg.Addr)
	err = app.run(cfg.Addr, handler)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
