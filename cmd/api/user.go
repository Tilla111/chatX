package main

import (
	"chatX/internal/store"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// GetUserHandler godoc
// @Summary      Foydalanuvchilar ro'yxatini olish
// @Description  Tizimdagi foydalanuvchilarni qidirish va sahifalangan ro'yxatini qaytaradi.
// @Tags         users
// @Produce      json
// @Param        user_id  path      int     true   "So'rov yuborayotgan foydalanuvchi IDsi"
// @Param        limit    query     int     false  "Natijalar soni" default(20)
// @Param        offset   query     int     false  "Surilish (offset)" default(0)
// @Param        search   query     string  false  "Qidiruv matni (ism yoki login)"
// @Success      200      {object}  map[string][]service.User "Foydalanuvchilar: {"data": [User ob'ektlari]}"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /users/{user_id} [get]
func (app *application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	pg := store.PaginationQuery{
		Limit:  20,
		Offset: 0,
		Search: "",
	}

	qr, err := pg.Parse(r)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := Validate.Struct(qr); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	users, err := app.services.UserSrvc.GetUsers(ctx, id, qr)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	fmt.Println(users)

	if err := app.jsonResponse(w, http.StatusOK, users); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}
