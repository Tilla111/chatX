package main

import (
	"chatX/internal/store"
	"net/http"
)

// GetUserHandler godoc
// @Summary      Foydalanuvchilar ro'yxati
// @Description  Joriy userdan tashqari userlarni pagination va search bilan qaytaradi.
// @Tags         users
// @Produce      json
// @Param        X-User-ID  header    int                true   "Joriy foydalanuvchi IDsi"
// @Param        limit      query     int                false  "Sahifadagi element soni (1..20)" default(20)
// @Param        offset     query     int                false  "Qaysi elementdan boshlab olish" default(0)
// @Param        search     query     string             false  "Username bo'yicha qidiruv (max 10 ta belgi)"
// @Success      200        {object}  map[string]any     "{"data":[...foydalanuvchilar...]}"
// @Failure      400        {object}  map[string]string  "Query param noto'g'ri"
// @Failure      401        {object}  map[string]string  "X-User-ID yuborilmagan yoki noto'g'ri"
// @Failure      500        {object}  map[string]string  "Ichki server xatosi"
// @Router       /users [get]
func (app *application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := app.requireUserID(w, r)
	if !ok {
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

	users, err := app.services.UserSrvc.GetUsers(r.Context(), int(currentUserID), query)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, users); err != nil {
		app.internalServerError(w, r, err)
	}
}
