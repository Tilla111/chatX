package main

import (
	"chatX/internal/auth"
	"chatX/internal/db"
	"chatX/internal/env"
	"chatX/internal/mailer"
	"chatX/internal/store"
	service "chatX/internal/usecase"
	"chatX/internal/ws"
	"log"
	"time"

	"go.uber.org/zap"
)

const Version = "v1.0.0"

//	@title			ChatX API
//	@description	Api for gophers Chat Project !
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/api/v1
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Bearer JWT token: `Bearer <token>`
func main() {

	cfgEnv, err := env.Load()
	if err != nil {
		log.Fatalf("Error loading env: %v", err)
	}

	addr := cfgEnv.Server.Port
	if addr != "" && addr[0] != ':' {
		addr = ":" + addr
	}

	cfg := config{
		Addr: addr,
		DB: DBConfig{
			Addr:         cfgEnv.Database.Addr,
			Host:         cfgEnv.Database.Host,
			User:         cfgEnv.Database.User,
			Password:     cfgEnv.Database.Password,
			Name:         cfgEnv.Database.Name,
			MaxIdleConns: cfgEnv.Database.MaxIdleConns,
			MaxOpenConns: cfgEnv.Database.MaxOpenConns,
			MaxIdletime:  cfgEnv.Database.MaxIdletime,
		},
		ENV:     cfgEnv.App.ENV,
		Upgrade: Upgrader,
		apiURL:  cfgEnv.App.APIURL,
		mail: MailConfig{
			mailtrap: mailtrapConfig{
				host:     cfgEnv.Email.Host,
				port:     cfgEnv.Email.Port,
				username: cfgEnv.Email.Username,
				password: cfgEnv.Email.Password,
			},
			fromEmail: cfgEnv.Email.FromEmail,
			exp:       time.Hour * 24 * 3, //3 Days
		},
		auth: authConfig{
			token: tokenConfig{
				secret: cfgEnv.Auth.SecretKey,
				exp:    time.Hour * 24, //1 Day
			},
		},
		app: appConfig{
			Audience: cfgEnv.App.Audience,
			Issuer:   cfgEnv.App.Issuer,
		},
	}

	logger := *zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.NewPostgres(
		cfg.DB.Addr,
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdletime,
	)
	if err != nil {
		logger.Fatal("Error connecting to database",
			zap.Error(err),
		)
	}
	logger.Infow("database connection status",
		"status", "connected",
		"host", cfg.DB.Addr,
	)

	hub := ws.NewHub()
	go hub.Run()

	storage := store.NewStorage(db)
	services := service.NewServices(storage)
	mailer := mailer.NewMailtrap(
		cfg.mail.mailtrap.host,
		cfg.mail.mailtrap.port,
		cfg.mail.mailtrap.username,
		cfg.mail.mailtrap.password,
		cfg.mail.fromEmail,
	)

	authService := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfgEnv.App.Audience, cfgEnv.App.Issuer)

	app := &application{
		config:   cfg,
		services: *services,
		ws:       hub,
		logger:   logger,
		mailer:   mailer,
		auth:     authService,
	}

	handler := app.mount()

	logger.Infow("Starting server",
		"addr", cfg.Addr,
		"env", cfg.ENV,
	)
	err = app.run(cfg.Addr, handler)
	if err != nil {
		logger.Fatalw("Server failed to start",
			"error", err,
			"addr", cfg.Addr,
		)
	}
}
