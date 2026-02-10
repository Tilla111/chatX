package main

import (
	"chatX/internal/store"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func (app *application) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	members, err := app.services.MemberSRV.GetByChatID(ctx, id)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, members); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

func (app *application) DeleteMemberHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	uid, err := strconv.Atoi(chi.URLParam(r, "user_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.services.MemberSRV.Delete(ctx, id, uid)
	if err != nil {
		switch err {
		case store.SqlNotfound:
			app.NotFound(w, r)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
