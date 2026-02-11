package main

import (
	"chatX/internal/store"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetMembersHandler godoc
// @Summary      Chat a'zolarini olish
// @Description  Berilgan group chat uchun a'zolar ro'yxatini qaytaradi. Faqat chat a'zosi ko'ra oladi.
// @Tags         members
// @Produce      json
// @Param        X-User-ID  header    int                true   "Joriy foydalanuvchi IDsi"
// @Param        chat_id    path      int                true   "Group chat ID"
// @Success      200        {object}  map[string]any     "{"data":[...a'zolar...]}"
// @Failure      400        {object}  map[string]string  "chat_id noto'g'ri"
// @Failure      401        {object}  map[string]string  "X-User-ID yuborilmagan yoki noto'g'ri"
// @Failure      403        {object}  map[string]string  "User chat a'zosi emas"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /groups/{chat_id}/members [get]
func (app *application) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := app.requireUserID(w, r)
	if !ok {
		return
	}

	chatID, err := strconv.ParseInt(chi.URLParam(r, "chat_id"), 10, 64)
	if err != nil || chatID <= 0 {
		app.badRequestError(w, r, errors.New("chat_id must be a positive integer"))
		return
	}

	isMember, err := app.services.MemberSRV.IsMember(r.Context(), chatID, userID)
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

// DeleteMemberHandler godoc
// @Summary      A'zoni groupdan chiqarish
// @Description  Groupdan userni chiqaradi. O'zini chiqarish mumkin, boshqa userni esa owner/admin chiqara oladi.
// @Tags         members
// @Param        X-User-ID  header    int                true   "Amalni bajarayotgan foydalanuvchi IDsi"
// @Param        chat_id    path      int                true   "Group chat ID"
// @Param        user_id    path      int                true   "Chiqariladigan user ID"
// @Success      204        "Muvaffaqiyatli chiqarildi"
// @Failure      400        {object}  map[string]string  "Path param noto'g'ri"
// @Failure      401        {object}  map[string]string  "X-User-ID yuborilmagan yoki noto'g'ri"
// @Failure      403        {object}  map[string]string  "Ruxsat yo'q"
// @Failure      404        {object}  map[string]string  "Member yoki chat topilmadi"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /groups/{chat_id}/{user_id}/member [delete]
func (app *application) DeleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := app.requireUserID(w, r)
	if !ok {
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

	if err := app.services.MemberSRV.Delete(r.Context(), actorID, chatID, targetID); err != nil {
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
