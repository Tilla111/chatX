package main

import (
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func getUserIDFromRequest(r *http.Request) (int64, error) {
	raw := strings.TrimSpace(r.Header.Get("X-User-ID"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("user_id"))
	}
	if raw == "" {
		return 0, errors.New("X-User-ID header is required")
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("X-User-ID must be a positive integer")
	}

	return id, nil
}

func (app *application) requireUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		app.unauthorizedError(w, r, err)
		return 0, false
	}

	return userID, true
}

// registerUserHandler godoc
// @Summary      Foydalanuvchini ro'yxatdan o'tkazish
// @Description  Yangi foydalanuvchi yaratadi va accountni aktivatsiya qilish uchun token qaytaradi.
// @Description  Frontend faqat `username`, `email`, `password` maydonlarini yuborishi kerak.
// @Description  Validation qoidalari: `username` (required, max 50), `email` (required, email format, max 72), `password` (required).
// @Description  Body'da noma'lum field bo'lsa yoki JSON noto'g'ri bo'lsa 400 qaytadi (`readJSON` unknown fieldlarni rad etadi).
// @Description  Muvaffaqiyatli javob formati: `{"data":"<activation_token>"}`. Xatolik formati: `{"error":"<message>"}`.
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        payload  body      service.RequestRegister  true  "Registration payload"
// @Success      201      {object}  map[string]string        "{"data":"2f1a2e89-5e5f-4f73-8abf-2f3dca9f7b71"}"
// @Failure      400      {object}  map[string]string        "Body/validation xatosi yoki email/username band"
// @Failure      500      {object}  map[string]string        "Ichki server xatosi"
// @Router       /users/authentication [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var req service.RequestRegister

	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	exp := app.mail.exp
	ctx := r.Context()
	token, err := app.services.UserSrvc.RegisterUser(ctx, req, exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
