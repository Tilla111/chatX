package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type addMemberRequest struct {
	UserID int64 `json:"user_id" validate:"required,gt=0"`
}

// GetMembersHandler godoc
//
//	@Summary		Chat a'zolarini olish
//	@Description	Berilgan group chat uchun a'zolar ro'yxatini qaytaradi. Faqat chat a'zosi ko'ra oladi.
//	@Tags			members
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer token: Bearer <token>"
//	@Param			chat_id			path		int					true	"Group chat ID"
//	@Success		200				{object}	map[string]any		"{"data":[...a'zolar...]}"
//	@Failure		400				{object}	map[string]string	"chat_id noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		403				{object}	map[string]string	"User chat a'zosi emas"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/groups/{chat_id}/members [get]
func (app *application) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	chatID, err := strconv.ParseInt(chi.URLParam(r, "chat_id"), 10, 64)
	if err != nil || chatID <= 0 {
		app.badRequestError(w, r, errors.New("chat_id must be a positive integer"))
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

	members, err := app.services.MemberSRV.GetByChatID(r.Context(), int(chatID))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, members); err != nil {
		app.internalServerError(w, r, err)
	}
}

// AddMemberHandler godoc
//
//	@Summary		Groupga a'zo qo'shish
//	@Description	Group chatga yangi a'zo qo'shadi. Amalni faqat owner yoki admin bajarishi mumkin.
//	@Tags			members
//	@Accept			json
//	@Produce		json
//	@Param			Authorization	header		string				true	"Bearer token: Bearer <token>"
//	@Param			chat_id			path		int					true	"Group chat ID"
//	@Param			payload			body		addMemberRequest	true	"Qo'shiladigan user ID"
//	@Success		201				{object}	map[string]any		"{"data":{"result":"added","user_id":21}}"
//	@Failure		400				{object}	map[string]string	"Path param yoki body noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		403				{object}	map[string]string	"Ruxsat yo'q"
//	@Failure		404				{object}	map[string]string	"Chat topilmadi"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/groups/{chat_id}/members [post]
func (app *application) AddMemberHandler(w http.ResponseWriter, r *http.Request) {
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

	var req addMemberRequest
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	err = app.services.MemberSRV.Add(r.Context(), senderID.ID, chatID, int(req.UserID))
	if err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		case errors.Is(err, store.SqlForbidden):
			app.forbiddenError(w, r, err)
		case errors.Is(err, service.ErrInvalidMemberChatType), errors.Is(err, service.ErrMemberAlreadyExists):
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	memberUsers, err := app.services.MemberSRV.GetByChatID(r.Context(), chatID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	memberIDs := make([]string, len(memberUsers))
	addedUsername := ""
	for i, user := range memberUsers {
		memberIDs[i] = strconv.FormatInt(user.ID, 10)
		if user.ID == req.UserID {
			addedUsername = user.UserName
		}
	}
	if addedUsername == "" {
		addedUsername = "foydalanuvchi"
	}

	addedByName := senderID.UserName
	if addedByName == "" {
		addedByName = "Kimdir"
	}

	go app.ws.BroadcastMemberAdded(
		int64(chatID),
		req.UserID,
		senderID.ID,
		addedUsername,
		addedByName,
		memberIDs,
	)

	if err := app.jsonResponse(w, http.StatusCreated, map[string]any{
		"result":  "added",
		"user_id": req.UserID,
	}); err != nil {
		app.internalServerError(w, r, err)
	}
}

// DeleteMemberHandler godoc
//
//	@Summary		A'zoni groupdan chiqarish
//	@Description	Groupdan userni chiqaradi. O'zini chiqarish mumkin, boshqa userni esa owner/admin chiqara oladi.
//	@Tags			members
//	@Param			Authorization	header	string	true	"Bearer token: Bearer <token>"
//	@Param			chat_id			path	int		true	"Group chat ID"
//	@Param			user_id			path	int		true	"Chiqariladigan user ID"
//	@Success		204				"Muvaffaqiyatli chiqarildi"
//	@Failure		400				{object}	map[string]string	"Path param noto'g'ri"
//	@Failure		401				{object}	map[string]string	"Authorization Bearer token yuborilmagan yoki noto'g'ri"
//	@Failure		403				{object}	map[string]string	"Ruxsat yo'q"
//	@Failure		404				{object}	map[string]string	"Member yoki chat topilmadi"
//	@Failure		500				{object}	map[string]string	"Ichki server xatosi"
//	@Router			/groups/{chat_id}/{user_id}/member [delete]
func (app *application) DeleteMemberHandler(w http.ResponseWriter, r *http.Request) {
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

	targetID, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil || targetID <= 0 {
		app.badRequestError(w, r, errors.New("user_id must be a positive integer"))
		return
	}

	if err := app.services.MemberSRV.Delete(r.Context(), senderID.ID, chatID, targetID); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		case errors.Is(err, store.SqlForbidden):
			app.forbiddenError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
