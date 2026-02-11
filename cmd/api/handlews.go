package main

import (
	"chatX/internal/ws"
	"net/http"
	"strconv"
)

func (app *application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := app.requireUserID(w, r)
	if !ok {
		return
	}

	conn, err := app.config.Upgrade.Upgrade(w, r, nil)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	client := &ws.Client{
		ID:   strconv.FormatInt(userID, 10),
		Hub:  app.ws,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Har bir ulanish uchun 2 ta alohida goroutina
	go client.WritePump()
	go client.ReadPump()
}
