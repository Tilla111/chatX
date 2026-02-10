package main

import (
	service "chatX/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// MessageCreateHandler godoc
// @Summary      Xabar yuborish
// @Description  Yangi xabar yaratadi va uni guruh a'zolariga WebSocket orqali real-vaqtda tarqatadi.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        payload body service.Message true "Xabar ma'lumotlari"
// @Success      201  {object}  map[string]service.Message "Xabar yaratildi: {"data": {message_object}}"
// @Failure      400  {object}  map[string]string "Noto'g'ri JSON formati"
// @Failure      500  {object}  map[string]string "Server xatosi"
// @Router       /messages [post]
func (app *application) MessageCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req service.Message
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	msg, err := app.services.MessageSRV.Create(ctx, req)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(ctx, int(msg.ChatID))
	if err != nil {
		app.internalServerError(w, r, err)
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

// GetMessagesHandler godoc
// @Summary      Chat xabarlarini olish
// @Description  Berilgan chat_id bo'yicha barcha xabarlar tarixini qaytaradi.
// @Tags         messages
// @Produce      json
// @Param        chat_id  path      int  true  "Chat ID"
// @Success      200      {object}  map[string][]service.MessageDetail "Xabarlar ro'yxati: {"data": [MessageDetail ob'ektlari]}"
// @Failure      400      {object}  map[string]string "Noto'g'ri ID"
// @Failure      500      {object}  map[string]string "Server xatosi"
// @Router       /chats/{chat_id}/messages [get]
func (app *application) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	msg, err := app.services.MessageSRV.GetByChatID(ctx, int64(id))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, msg)
}

// MarkAsReadHandler godoc
// @Summary      Xabarlarni o'qilgan deb belgilash
// @Description  Chatdagi barcha xabarlarni joriy foydalanuvchi uchun o'qilgan holatiga o'tkazadi va bu haqda boshqa a'zolarga xabar beradi.
// @Tags         messages
// @Produce      json
// @Param        chat_id  path      int  true  "Chat ID"
// @Success      200      {object}  map[string]map[string]string "Muvaffaqiyatli: {"data": {"status": "success"}}"
// @Failure      400      {object}  map[string]string "Noto'g'ri ID format"
// @Failure      500      {object}  map[string]string "Server xatosi"
// @Router       /messages/{chat_id} [patch]
func (app *application) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {

	chatIDStr := chi.URLParam(r, "chat_id")
	chatID, _ := strconv.ParseInt(chatIDStr, 10, 64)

	userID := int64(1)

	ctx := r.Context()
	if err := app.services.MessageSRV.MarkChatAsRead(ctx, chatID, userID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(ctx, int(chatID))
	if err != nil {
		app.internalServerError(w, r, err)
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

// MessageUpdateHandler godoc
// @Summary      Xabarni tahrirlash
// @Description  Yuborilgan xabar matnini o'zgartiradi va bu haqda guruh a'zolariga WebSocket orqali xabar beradi.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        id      path      int     true  "Xabar IDsi"
// @Param        payload body      object  true  "Tahrirlash ma'lumotlari (message_text va chat_id)"
// @Success      200     {object}  map[string]map[string]string "Muvaffaqiyatli: {"data": {"result": "updated"}}"
// @Failure      400     {object}  map[string]string "Noto'g'ri JSON yoki ID"
// @Failure      500     {object}  map[string]string "Server xatosi"
// @Router       /messages/{id} [patch]
func (app *application) MessageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	msgID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	userID := int64(1)

	var input struct {
		MessageText string `json:"message_text"`
		ChatID      int64  `json:"chat_id"` // WS uchun kerak
	}
	if err := readJSON(w, r, &input); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := app.services.MessageSRV.UpdateMessage(r.Context(), msgID, userID, input.MessageText); err != nil {
		app.internalServerError(w, r, err)
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

// MessageDeleteHandler godoc
// @Summary      Xabarni o'chirish
// @Description  Xabarni ma'lumotlar bazasidan o'chiradi va WebSocket orqali barcha chat a'zolariga xabar o'chirilganligi haqida signal yuboradi.
// @Tags         messages
// @Produce      json
// @Param        id   path      int  true  "O'chirilishi kerak bo'lgan xabar IDsi"
// @Success      200  {object}  map[string]map[string]string "Muvaffaqiyatli: {"data": {"result": "deleted"}}"
// @Failure      400  {object}  map[string]string "Noto'g'ri ID format"
// @Failure      404  {object}  map[string]string "Xabar topilmadi"
// @Failure      500  {object}  map[string]string "Server xatosi"
// @Router       /messages/{id} [delete]
func (app *application) MessageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	msgID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	userID := int64(1)

	// Xabar haqidagi ma'lumotni bazadan olish (chatID kerakligi uchun)
	msg, _ := app.services.MessageSRV.GetByID(r.Context(), msgID)

	if err := app.services.MessageSRV.DeleteMessage(r.Context(), msgID, userID); err != nil {
		app.internalServerError(w, r, err)
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
