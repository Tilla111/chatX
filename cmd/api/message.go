package main

import (
	service "chatX/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (app *application) MessageCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req service.Message
	if err := readJSON(w, r, &req); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	msg, err := app.services.MessageSRV.Create(ctx, req)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(ctx, int(msg.ChatID))
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	senderIDStr := strconv.FormatInt(msg.SenderID, 10)

	go app.ws.BroadcastChatMessage(
		msg.ChatID,      // int64
		msg.ChatName,    // string
		senderIDStr,     // string
		msg.SenderName,  // string
		msg.MessageText, // string
		memberIDs,       // []string
	)

	app.jsonResponse(w, http.StatusCreated, msg)
}

func (app *application) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	msg, err := app.services.MessageSRV.GetByChatID(ctx, int64(id))
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, msg)
}

func (app *application) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {

	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, _ := strconv.ParseInt(chatIDStr, 10, 64)

	userID := int64(1)

	ctx := r.Context()
	if err := app.services.MessageSRV.MarkChatAsRead(ctx, chatID, userID); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(ctx, int(chatID))
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	// MarkAsReadHandler ichida:
	for _, memberID := range memberIDs {
		go app.ws.BroadcastReadStatus(chatID, strconv.FormatInt(userID, 10), memberID)
	}

	app.jsonResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// MessageUpdateHandler
func (app *application) MessageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	msgID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	userID := int64(1)

	var input struct {
		MessageText string `json:"message_text"`
		ChatID      int64  `json:"chat_id"` // WS uchun kerak
	}
	if err := readJSON(w, r, &input); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := app.services.MessageSRV.UpdateMessage(r.Context(), msgID, userID, input.MessageText); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	// WebSocket orqali hammani xabardor qilish
	memberUsers, _ := app.services.MemberSRV.GetByChatID(r.Context(), int(input.ChatID))
	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}
	go app.ws.BroadcastMessageUpdate(input.ChatID, msgID, input.MessageText, memberIDs)

	app.jsonResponse(w, http.StatusOK, map[string]string{"result": "updated"})
}

func (app *application) MessageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	msgID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	userID := int64(1)

	// Xabar haqidagi ma'lumotni bazadan olish (chatID kerakligi uchun)
	msg, _ := app.services.MessageSRV.GetByID(r.Context(), msgID)

	if err := app.services.MessageSRV.DeleteMessage(r.Context(), msgID, userID); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	// Chat a'zolarini olish
	memberUsers, _ := app.services.MemberSRV.GetByChatID(r.Context(), int(msg.ChatID))
	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	// WebSocket signalini yuborish
	go app.ws.BroadcastMessageDelete(msg.ChatID, msgID, memberIDs)

	app.jsonResponse(w, http.StatusOK, map[string]string{"result": "deleted"})
}
