package main

import (
	"chatX/internal/mailer"
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) buildActivationURL(token string) string {
	baseURL := strings.TrimSpace(app.config.apiURL)
	if baseURL == "" {
		baseURL = "localhost" + app.config.Addr
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}

	return fmt.Sprintf(
		"%s/api/v1/users/activate/%s",
		strings.TrimRight(baseURL, "/"),
		url.PathEscape(token),
	)
}

// registerUserHandler godoc
// @Summary      Foydalanuvchini ro'yxatdan o'tkazish
// @Description  Yangi foydalanuvchi yaratadi va accountni aktivatsiya qilish uchun email yuboradi.
// @Description  Frontend faqat `username`, `email`, `password` maydonlarini yuborishi kerak.
// @Description  Validation qoidalari: `username` (required, max 50), `email` (required, email format, max 72), `password` (required).
// @Description  Body'da noma'lum field bo'lsa yoki JSON noto'g'ri bo'lsa 400 qaytadi (`readJSON` unknown fieldlarni rad etadi).
// @Description  Muvaffaqiyatli javob formati: `{"data":{"message":"..."}}`. Xatolik formati: `{"error":"<message>"}`.
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        payload  body      service.RequestRegister  true  "Registration payload"
// @Success      201      {object}  map[string]any           "{"data":{"message":"registration successful"}}"
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

	exp := app.config.mail.exp
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

	activationURL := app.buildActivationURL(token)

	isSandbox := app.config.ENV != "prod"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      req.Username,
		ActivationURL: activationURL,
	}
	err = app.mailer.Send(mailer.UserWelcomeTemplate, req.Username, req.Email, vars, isSandbox)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, map[string]string{
		"message": "registration successful. Check your email to activate your account.",
	}); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type RequestCreateToken struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// createTokenHandler godoc
// @Summary      Token yaratish
// @Description  Foydalanuvchi email va parol asosida token yaratadi.
// @Description  Body'da `email` va `password` maydonlari bo'lishi kerak. Validation qoidalari: har ikkalasi ham required.
// @Description  Muvaffaqiyatli javob formati: `{"data":{"token":"..."}}`. Xatolik formati: `{"error":"<message>"}`.
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        payload  body      RequestCreateToken  true  "Token yaratish payload"
// @Success      200      {object}  map[string]any     "{"data":{"token":"..."}}"
// @Failure      400      {object}  map[string]string  "Body/validation xatosi"
// @Failure      401      {object}  map[string]string  "Noto'g'ri username yoki parol"
// @Failure      500      {object}  map[string]string  "Ichki server xatosi"
// @Router       /users/authentication/token [post]

func (app *application) CreateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RequestCreateToken
	if err := readJSON(w, r, &req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	// Get User by email
	user, err := app.services.UserSrvc.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.SqlNotfound) {
			app.unauthorizedError(w, r, errors.New("invalid email or password"))
		} else {
			app.internalServerError(w, r, err)
		}
		return
	}

	//check password
	if err := app.services.UserSrvc.CheckPassword(user, req.Password); err != nil {
		app.unauthorizedError(w, r, errors.New("invalid email or password"))
		return
	}

	//create token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": user.ID,
		"aud": app.config.app.Audience,
		"iss": app.config.app.Issuer,
		"exp": now.Add(app.config.auth.token.exp).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}
	token, err := app.auth.CreateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, map[string]string{
		"token": token,
	}); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
