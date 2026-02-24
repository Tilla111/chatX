package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type createMessageRequest struct {
	ChatID      int64  `json:"chat_id" validate:"required,gt=0"`
	MessageText string `json:"message_text" validate:"required,max=4000"`
}

type updateMessageRequest struct {
	MessageText string `json:"message_text" validate:"required,max=4000"`
}

func parsePathInt64(raw string, paramName string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(paramName + " must be a positive integer")
	}
	return id, nil
}

// MessageCreateHandler godoc
// @Summary      Xabar yuborish
// @Description  Joriy foydalanuvchi berilgan chatga yangi xabar yuboradi.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                   true   "Bearer token: Bearer <token>"
// @Param        payload    body      createMessageRequest  true   "Xabar yuborish ma'lumotlari"
// @Success      201        {object}  map[string]any        "{"data":{...xabar...}}"
// @Failure      400        {object}  map[string]string     "Body noto'g'ri"
// @Failure      401        {object}  map[string]string     "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      403        {object}  map[string]string     "User chat a'zosi emas"
// @Failure      500        {object}  map[string]string     "Ichki server xatosi"
// @Router       /messages [post]
func (app *application) MessageCreateHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	var req createMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), req.ChatID, senderID.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if !isMember {
		app.forbiddenError(w, r, errors.New("user is not a member of this chat"))
		return
	}

	msg, err := app.services.MessageSRV.Create(r.Context(), service.Message{
		ChatID:      req.ChatID,
		SenderID:    senderID.ID,
		MessageText: req.MessageText,
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), int(msg.ChatID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	go app.ws.BroadcastChatMessage(
		msg.ChatID,
		msg.ChatName,
		strconv.FormatInt(msg.SenderID, 10),
		msg.SenderName,
		msg.MessageText,
		memberIDs,
	)

	if err := app.jsonResponse(w, http.StatusCreated, msg); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetMessagesHandler godoc
// @Summary      Chat xabarlarini olish
// @Description  Berilgan chatdagi xabarlar tarixini qaytaradi. Faqat chat a'zosi ko'ra oladi.
// @Tags         messages
// @Produce      json
// @Param        Authorization  header    string                true   "Bearer token: Bearer <token>"
// @Param        chat_id    path      int                true   "Chat ID"
// @Success      200        {object}  map[string]any     "{"data":[...xabarlar...]}"
// @Failure      400        {object}  map[string]string  "chat_id noto'g'ri"
// @Failure      401        {object}  map[string]string  "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      403        {object}  map[string]string  "User chat a'zosi emas"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /chats/{chat_id}/messages [get]
func (app *application) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	chatID, err := parsePathInt64(chi.URLParam(r, "chat_id"), "chat_id")
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), chatID, senderID.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if !isMember {
		app.forbiddenError(w, r, errors.New("user is not a member of this chat"))
		return
	}

	msg, err := app.services.MessageSRV.GetByChatID(r.Context(), chatID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, msg); err != nil {
		app.internalServerError(w, r, err)
	}
}

// MarkAsReadHandler godoc
// @Summary      Chatdagi xabarlarni o'qilgan deb belgilash
// @Description  Joriy foydalanuvchi uchun berilgan chatdagi barcha kiruvchi xabarlarni o'qilgan holatiga o'tkazadi.
// @Tags         messages
// @Produce      json
// @Param        Authorization  header    string                true   "Bearer token: Bearer <token>"
// @Param        chat_id    path      int                true   "Chat ID"
// @Success      200        {object}  map[string]any     "{"data":{"status":"success"}}"
// @Failure      400        {object}  map[string]string  "chat_id noto'g'ri"
// @Failure      401        {object}  map[string]string  "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      403        {object}  map[string]string  "User chat a'zosi emas"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /messages/chats/{chat_id}/read [patch]
func (app *application) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	chatID, err := parsePathInt64(chi.URLParam(r, "chat_id"), "chat_id")
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), chatID, senderID.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if !isMember {
		app.forbiddenError(w, r, errors.New("user is not a member of this chat"))
		return
	}

	if err := app.services.MessageSRV.MarkChatAsRead(r.Context(), chatID, senderID.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), int(chatID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	readerID := strconv.FormatInt(senderID.ID, 10)
	for _, user := range memberUsers {
		recipientID := strconv.FormatInt(user.ID, 10)
		if recipientID == readerID {
			continue
		}
		go app.ws.BroadcastReadStatus(chatID, readerID, recipientID)
	}

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{"status": "success"}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// MessageUpdateHandler godoc
// @Summary      Xabarni tahrirlash
// @Description  Joriy foydalanuvchi o'zi yuborgan xabar matnini yangilaydi.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                   true   "Bearer token: Bearer <token>"
// @Param        id         path      int                   true   "Xabar ID"
// @Param        payload    body      updateMessageRequest  true   "Yangilangan xabar matni"
// @Success      200        {object}  map[string]any        "{"data":{"result":"updated"}}"
// @Failure      400        {object}  map[string]string     "ID yoki body noto'g'ri"
// @Failure      401        {object}  map[string]string     "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      404        {object}  map[string]string     "Xabar topilmadi yoki userga tegishli emas"
// @Failure      500        {object}  map[string]string     "Ichki server xatosi"
// @Router       /messages/{id} [patch]
func (app *application) MessageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	msgID, err := parsePathInt64(chi.URLParam(r, "id"), "id")
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	var req updateMessageRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	msg, err := app.services.MessageSRV.GetByID(r.Context(), msgID)
	if err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.services.MessageSRV.UpdateMessage(r.Context(), msgID, senderID.ID, req.MessageText); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), int(msg.ChatID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	go app.ws.BroadcastMessageUpdate(msg.ChatID, msgID, req.MessageText, memberIDs)

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{"result": "updated"}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// MessageDeleteHandler godoc
// @Summary      Xabarni o'chirish
// @Description  Joriy foydalanuvchi o'zi yuborgan xabarni o'chiradi.
// @Tags         messages
// @Produce      json
// @Param        Authorization  header    string                true   "Bearer token: Bearer <token>"
// @Param        id         path      int                true   "Xabar ID"
// @Success      200        {object}  map[string]any     "{"data":{"result":"deleted"}}"
// @Failure      400        {object}  map[string]string  "ID noto'g'ri"
// @Failure      401        {object}  map[string]string  "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      404        {object}  map[string]string  "Xabar topilmadi yoki userga tegishli emas"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /messages/{id} [delete]
func (app *application) MessageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	msgID, err := parsePathInt64(chi.URLParam(r, "id"), "id")
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	msg, err := app.services.MessageSRV.GetByID(r.Context(), msgID)
	if err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.services.MessageSRV.DeleteMessage(r.Context(), msgID, senderID.ID); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), int(msg.ChatID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	go app.ws.BroadcastMessageDelete(msg.ChatID, msgID, memberIDs)

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{"result": "deleted"}); err != nil {
		app.internalServerError(w, r, err)
	}
}
