package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type createPrivateChatRequest struct {
	ReceiverID int64 `json:"receiver_id" validate:"required,gt=0"`
}

type createGroupRequest struct {
	MemberIDs   []int64 `json:"member_ids" validate:"required,min=1,dive,gt=0"`
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description" validate:"max=255"`
}

type updateGroupRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=255"`
}

// CreatePrivateChatHandler godoc
//
//	@Summary		Private chat yaratish
//	@Description	Ikki foydalanuvchi orasida private chat yaratadi. Agar chat oldin yaratilgan bo'lsa, o'sha chat_id qaytadi.
//	@Tags			chats
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string						true	"Bearer token: Bearer <token>"
//	@Param			payload			body		createPrivateChatRequest	true	"Private chat uchun receiver ma'lumoti"
//	@Success		201				{object}	map[string]any				"{"data":{"chat_id":12}}"
//	@Failure		400				{object}	map[string]string			"So'rov noto'g'ri"
//	@Failure		401				{object}	map[string]string			"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		500				{object}	map[string]string			"Ichki server xatosi"
//	@Router			/chats [post]
func (app *application) CreatechatHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	var req createPrivateChatRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	chatID, err := app.services.ChatSRVC.CreatePrivateChat(r.Context(), senderID.ID, req.ReceiverID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, map[string]int64{"chat_id": chatID}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// CreateGroupHandler godoc
//
//	@Summary		Group chat yaratish
//	@Description	Yangi group chat yaratadi, joriy userni owner qiladi va `member_ids` dagi userlarni qo'shadi.
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer token: Bearer <token> (owner bo'ladi)"
//	@Param			payload			body		createGroupRequest	true	"Group yaratish ma'lumotlari"
//	@Success		201				{object}	map[string]any		"{"data":{"chat_id":17}}"
//	@Failure		400				{object}	map[string]string	"So'rov noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/groups [post]
func (app *application) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {

	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	var req createGroupRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	group := service.Group{
		SenderID:    senderID.ID,
		ReceiverID:  req.MemberIDs,
		Name:        req.Name,
		Description: req.Description,
	}

	chatID, err := app.services.ChatSRVC.CreateGroupChat(r.Context(), &group)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, map[string]int64{"chat_id": chatID}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetUserChatsHandler godoc
//
//	@Summary		Joriy user chatlari
//	@Description	Joriy foydalanuvchiga tegishli private va group chatlar ro'yxatini qaytaradi.
//	@Tags			chats
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer token: Bearer <token>"
//	@Param			search			query		string				false	"Chat nomi bo'yicha qidiruv"
//	@Success		200				{object}	map[string]any		"{"data":[...chatlar...]}"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/chats [get]
func (app *application) GetUserChatsHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}
	searchTerm := r.URL.Query().Get("search")
	chats, err := app.services.ChatSRVC.GetUserChats(r.Context(), senderID.ID, searchTerm)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, chats); err != nil {
		app.internalServerError(w, r, err)
	}
}

// UpdateChatHandler godoc
//
//	@Summary		Group ma'lumotlarini yangilash
//	@Description	Berilgan `chat_id` bo'yicha group nomi va description qiymatlarini yangilaydi.
//	@Tags			groups
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer token: Bearer <token>"
//	@Param			chat_id			path		int					true	"Group chat ID"
//	@Param			payload			body		updateGroupRequest	true	"Yangilanadigan qiymatlar"
//	@Success		200				{object}	map[string]any		"{"data":{"result":"updated"}}"
//	@Failure		400				{object}	map[string]string	"ID yoki body noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		404				{object}	map[string]string	"Group topilmadi"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/groups/{chat_id} [patch]
func (app *application) UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	chatID, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil || chatID <= 0 {
		app.badRequestError(w, r, errors.New("chat_id must be a positive integer"))
		return
	}

	var req updateGroupRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	group := service.Chatgroup{
		ChatID:      chatID,
		GroupName:   req.Name,
		Description: req.Description,
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), int64(chatID), senderID.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if !isMember {
		app.forbiddenError(w, r, errors.New("user is not a member of this chat"))
		return
	}

	_, err = app.services.ChatSRVC.Updatechat(r.Context(), &group)
	if err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{"result": "updated"}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// DeleteChatHandler godoc
//
//	@Summary		Chatni o'chirish
//	@Description	Berilgan `chat_id` bo'yicha chatni o'chiradi.
//	@Tags			chats
//	@Param			Authorization	header	string	true	"Bearer token: Bearer <token>"
//	@Param			chat_id			path	int		true	"Chat ID"
//	@Success		204				"Muvaffaqiyatli o'chirildi"
//	@Failure		400				{object}	map[string]string	"chat_id noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		404				{object}	map[string]string	"Chat topilmadi"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/chats/{chat_id} [delete]
func (app *application) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}
	chatID, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil || chatID <= 0 {
		app.badRequestError(w, r, errors.New("chat_id must be a positive integer"))
		return
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), int64(chatID), senderID.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if !isMember {
		app.forbiddenError(w, r, errors.New("user is not a member of this chat"))
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), chatID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
	}

	if err := app.services.ChatSRVC.DeleteChat(r.Context(), chatID); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	deletedByName := senderID.UserName
	if deletedByName == "" {
		deletedByName = "Kimdir"
	}

	go app.ws.BroadcastChatDelete(int64(chatID), senderID.ID, deletedByName, memberIDs)

	w.WriteHeader(http.StatusNoContent)
}
