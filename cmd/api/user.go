package main

import (
	"chatX/internal/store"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// GetUserHandler godoc
// @Summary      Foydalanuvchilar ro'yxati
// @Description  Joriy userdan tashqari userlarni pagination va search bilan qaytaradi.
// @Tags         users
// @Produce      json
// @Param        Authorization  header    string                true   "Bearer token: Bearer <token>"
// @Param        limit      query     int                false  "Sahifadagi element soni (1..20)" default(20)
// @Param        offset     query     int                false  "Qaysi elementdan boshlab olish" default(0)
// @Param        search     query     string             false  "Username bo'yicha qidiruv (max 10 ta belgi)"
// @Success      200        {object}  map[string]any     "{"data":[...foydalanuvchilar...]}"
// @Failure      400        {object}  map[string]string  "Query param noto'g'ri"
// @Failure      401        {object}  map[string]string  "Authorization Bearer token yuborilmagan yoki noto'g'ri"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /users [get]
func (app *application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	senderID, ok := getUserfromContext(r)
	if !ok {
		app.unauthorizedError(w, r, errors.New("user not found in context"))
		return
	}

	pg := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Search: "",
	}

	query, err := pg.Parse(r)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(query); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	users, err := app.services.UserSrvc.GetUsers(r.Context(), int(senderID.ID), query)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, users); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateUserHandler godoc
// @Summary      Accountni aktivatsiya qilish
// @Description  Foydalanuvchi emailiga yuborilgan activation linkdagi token orqali accountni faollashtiradi.
// @Description  Frontend tokenni URL path orqali yuboradi: `/api/v1/users/activate/{token}`.
// @Description  Bu endpoint body qabul qilmaydi, faqat path param token kerak bo'ladi.
// @Description  Token bir martalik: account faollashgandan keyin token o'chiriladi.
// @Description  Token noto'g'ri, eskirgan yoki allaqachon ishlatilgan bo'lsa 400 qaytadi.
// @Description  Frontend emaildagi linkdan to'g'ridan-to'g'ri ochish uchun `GET`, API style chaqiriq uchun `PUT` ishlatishi mumkin.
// @Tags         users
// @Produce      json
// @Param        token  path      string             true   "Emailga yuborilgan activation token"
// @Success      200    {object}  map[string]any     "{"data":{"message":"account activated successfully"}}"
// @Failure      400    {object}  map[string]string  "{"error":"activation token is required | Not found"}"
// @Failure      500    {object}  map[string]string  "{"error":"internal server error"}"
// @Router       /users/activate/{token} [get]
// @Router       /users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

	token := strings.TrimSpace(chi.URLParam(r, "token"))
	if token == "" {
		app.badRequestError(w, r, errors.New("activation token is required"))
		return
	}

	ctx := r.Context()
	if err := app.services.UserSrvc.UserActivate(ctx, token); err != nil {
		switch err {
		case store.SqlNotfound:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{
		"message": "account activated successfully",
	}); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func getUserfromContext(r *http.Request) (*store.User, bool) {
	user, ok := r.Context().Value("user").(*store.User)
	return user, ok
}
