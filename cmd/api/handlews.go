package main

import (
	"chatX/internal/ws"
	"log"
	"net/http"
)

var testUsers = map[string]string{
	"1": "Ali",
	"2": "Vali",
	"3": "Gani",
}

func (app *application) handleWebSocket(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("user_id")

	// Test: Faqat ro'yxatdagi foydalanuvchilarga ruxsat beramiz
	if _, ok := testUsers[id]; !ok {
		http.Error(w, "Foydalanuvchi topilmadi", http.StatusForbidden)
		return
	}

	conn, err := app.config.Upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &ws.Client{
		ID:   id,
		Hub:  app.ws,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client

	// Har bir ulanish uchun 2 ta alohida goroutina
	go client.WritePump()
	go client.ReadPump()
}
