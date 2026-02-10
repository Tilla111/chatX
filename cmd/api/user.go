package main

import (
	"chatX/internal/store"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

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
