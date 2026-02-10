package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type ChatReq struct {
	SenderID   int64 `json:"sender_id" validate:"required,max=255"`
	ReceiverID int64 `json:"receiver_id" validate:"required,max=255"`
}

func (app *application) CreatechatHandler(w http.ResponseWriter, r *http.Request) {

	var req ChatReq
	if err := readJSON(w, r, &req); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()

	id, err := app.services.ChatSRVC.CreatePrivateChat(ctx, req.SenderID, req.ReceiverID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			app.BadRequest(w, r, err)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, id)
	if err != nil {
		app.InternalServerError(w, r, err)
	}

}

type groupReq struct {
	SenderID    int64   `json:"sender_id" validate:"required,max=255"`
	ReceiverID  []int64 `json:"receiver1_id" validate:"required,max=255"`
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description" validate:"required,max=255"`
}

func (app *application) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {

	var req groupReq
	if err := readJSON(w, r, &req); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()

	group := service.Group{
		SenderID:    req.SenderID,
		ReceiverID:  req.ReceiverID,
		Name:        req.Name,
		Description: req.Description,
	}

	id, err := app.services.ChatSRVC.CreateGroupChat(ctx, &group)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			app.BadRequest(w, r, err)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, id)
	if err != nil {
		app.InternalServerError(w, r, err)
	}

}

func (app *application) GetUserChatsHandler(w http.ResponseWriter, r *http.Request) {

	userID := int64(1)

	searchTerm := r.URL.Query().Get("search")

	ctx := r.Context()

	chats, err := app.services.ChatSRVC.GetUserChats(ctx, userID, searchTerm)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, chats)
}

type group struct {
	ID          int    `json:"id"`
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"required,max=255"`
}

func (app *application) UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	var g group

	if err := readJSON(w, r, &g); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	group := service.Chatgroup{
		ChatID:      id,
		GroupName:   g.Name,
		Description: g.Description,
	}

	if err := Validate.Struct(group); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()
	_, err = app.services.ChatSRVC.Updatechat(ctx, &group)
	if err != nil {
		switch err {
		case store.SqlNotfound:
			app.NotFound(w, r)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.services.ChatSRVC.DeleteChat(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.NotFound(w, r)
		default:
			app.InternalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
