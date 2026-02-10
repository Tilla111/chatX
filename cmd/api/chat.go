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

// CreateGroupChatHandler godoc
// @Summary      Guruh suhbatini yaratish
// @Description  Yangi guruh yaratadi va foydalanuvchilarni unga qo'shadi
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        payload body groupReq true "Guruh ma'lumotlari"
// @Success      201  {object}  map[string]int64 "Guruh IDsi qaytadi"
// @Failure      400  {object}  error "Noto'g'ri so'rov yuborilgan"
// @Failure      500  {object}  error "Server xatoligi"
// @Router       /chats [post]
func (app *application) CreatechatHandler(w http.ResponseWriter, r *http.Request) {

	var req ChatReq
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	id, err := app.services.ChatSRVC.CreatePrivateChat(ctx, req.SenderID, req.ReceiverID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, id)
	if err != nil {
		app.internalServerError(w, r, err)
	}

}

type groupReq struct {
	SenderID    int64   `json:"sender_id" validate:"required,max=255"`
	ReceiverID  []int64 `json:"receiver1_id" validate:"required,max=255"`
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description" validate:"required,max=255"`
}

// CreateGroupHandler godoc
// @Summary      Guruh suhbatini yaratish
// @Description  Yangi guruh chatini yaratadi va a'zolarni biriktiradi
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        payload body groupReq true "Guruh yaratish ma'lumotlari"
// @Success      201  {object}  map[string]interface{} "Guruh muvaffaqiyatli yaratildi"
// @Failure      400  {object}  map[string]string      "Noto'g'ri so'rov yoki validatsiya xatosi"
// @Failure      500  {object}  map[string]string      "Server ichki xatosi"
// @Router       /groups [post]
func (app *application) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {

	var req groupReq
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
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
			app.badRequestError(w, r, err)
		default:
			app.badRequestError(w, r, err)
		}
		return
	}

	err = writeJSON(w, http.StatusCreated, id)
	if err != nil {
		app.internalServerError(w, r, err)
	}

}

// GetUserChatsHandler godoc
// @Summary      Foydalanuvchi chatlarini ro'yxatini olish
// @Description  Tizimga kirgan foydalanuvchining barcha shaxsiy va guruh suhbatlarini qaytaradi.
// @Tags         chats
// @Accept       json
// @Produce      json
// @Param        search  query     string  false  "Chat nomi yoki oxirgi xabar bo'yicha qidirish"
// @Success      200     {object}   map[string][]service.ChatInfo  "Chatlar ro'yxati muvaffaqiyatli qaytarildi"
// @Failure      401     {object}  map[string]string "Avtorizatsiyadan o'tilmagan"
// @Failure      500     {object}  map[string]string "Serverning ichki xatosi"
// @Router       /chats [get]
func (app *application) GetUserChatsHandler(w http.ResponseWriter, r *http.Request) {

	userID := int64(1)

	searchTerm := r.URL.Query().Get("search")

	ctx := r.Context()

	chats, err := app.services.ChatSRVC.GetUserChats(ctx, userID, searchTerm)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, chats)
}

type group struct {
	ID          int    `json:"id"`
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"required,max=255"`
}

// UpdateChatHandler godoc
// @Summary      Guruh ma'lumotlarini yangilash
// @Description  Mavjud guruhning nomi va tavsifini o'zgartiradi. Chat ID path orqali yuboriladi.
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        chat_id      path      int     true  "Yangilanadigan chat (guruh) IDsi"
// @Param        request_body  body      group   true  "Yangi guruh ma'lumotlari"
// @Success      200           {string}  string  "Muvaffaqiyatli yangilandi"
// @Failure      400           {object}  map[string]string "Noto'g'ri ID yoki validatsiya xatosi"
// @Failure      404           {object}  map[string]string "Guruh topilmadi"
// @Failure      500           {object}  map[string]string "Server xatosi"
// @Router       /groups/{chat_id} [put]
func (app *application) UpdateChatHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	var g group

	if err := readJSON(w, r, &g); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	group := service.Chatgroup{
		ChatID:      id,
		GroupName:   g.Name,
		Description: g.Description,
	}

	if err := Validate.Struct(group); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()
	_, err = app.services.ChatSRVC.Updatechat(ctx, &group)
	if err != nil {
		switch err {
		case store.SqlNotfound:
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteChatHandler godoc
// @Summary      Chatni o'chirish
// @Description  Berilgan ID bo'yicha chatni (shaxsiy yoki guruh) butunlay o'chirib tashlaydi.
// @Tags         chats
// @Param        chat_id  path      int  true  "O'chirilishi kerak bo'lgan chat IDsi"
// @Success      204      "Muvaffaqiyatli o'chirildi (kontent qaytarilmaydi)"
// @Failure      400      {object}  map[string]string "Noto'g'ri ID formati"
// @Failure      404      {object}  map[string]string "Chat topilmadi"
// @Failure      500      {object}  map[string]string "Serverning ichki xatosi"
// @Router       /chats/{chat_id} [delete]
func (app *application) DeleteChatHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.services.ChatSRVC.DeleteChat(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.SqlNotfound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
