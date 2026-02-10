package main

import (
	"chatX/internal/db"
	"chatX/internal/env"
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"chatX/internal/ws"
	"log"
)

const version = "v1.0.0"

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
