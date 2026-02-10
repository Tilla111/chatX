package main

import (
	"chatX/internal/store"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GetMembersHandler godoc
// @Summary      Guruh a'zolarini olish
// @Description  Chat ID bo'yicha barcha a'zolar (foydalanuvchilar) ro'yxatini qaytaradi
// @Tags         members
// @Produce      json
// @Param        chat_id  path      int  true  "Chat ID"
// @Success      200      {object}  map[string][]store.User "Muvaffaqiyatli: {"data": [User ob'ektlari]}"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /groups/{chat_id}/members [get]
func (app *application) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	members, err := app.services.MemberSRV.GetByChatID(ctx, id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, members); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeleteMemberHandler godoc
// @Summary      Guruhdan a'zoni o'chirish
// @Description  Guruh suhbatidan foydalanuvchini olib tashlaydi. Faqat guruh admini yoki foydalanuvchining o'zi (chiqib ketish) bajara oladi.
// @Tags         members
// @Param        chat_id  path      int  true  "Chat (guruh) IDsi"
// @Param        user_id  path      int  true  "O'chirilishi kerak bo'lgan foydalanuvchi IDsi"
// @Success      204      "Muvaffaqiyatli o'chirildi"
// @Failure      400      {object}  map[string]string "ID xato yuborilgan"
// @Failure      404      {object}  map[string]string "A'zo yoki chat topilmadi"
// @Failure      500      {object}  map[string]string "Server xatosi"
// @Router       /chats/{chat_id}/{user_id}/member [delete]
func (app *application) DeleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	uid, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.services.MemberSRV.Delete(ctx, id, uid)
	if err != nil {
		switch err {
		case store.SqlNotfound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
