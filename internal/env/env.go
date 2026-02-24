package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Name           string   `yaml:"name"`
		ENV            string   `yaml:"env"`
		APIURL         string   `yaml:"api_url"`
		AllowedOrigins []string `yaml:"allowed_origins"`
		Audience       string   `yaml:"audience"`
		Issuer         string   `yaml:"issuer"`
	} `yaml:"app"`

	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Addr         string `yaml:"addr"`
		Host         string `yaml:"host"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Name         string `yaml:"name"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdletime  string `yaml:"max_idle_time"`
	}
	Email struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		FromEmail string `yaml:"fromEmail"`
	}
	Auth struct {
		SecretKey string `yaml:"secret_key"`
	}
}

func Load() (*Config, error) {

	_ = godotenv.Load()

	c := &Config{}

	env := getenv("APP_ENV", "dev")
	file := "./internal/env/config." + env + ".yaml"

	cfg, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cfg, c)
	if err != nil {
		return nil, err
	}

	c.Server.Port = getenv("PORT", c.Server.Port)
	c.Database.Addr = getenv("DB_PORT", c.Database.Addr)
	c.Database.Host = getenv("DB_HOST", c.Database.Host)
	c.Database.User = getenv("DB_USER", c.Database.User)
	c.Database.Password = getenv("DB_PASSWORD", c.Database.Password)
	c.Database.Name = getenv("DB_NAME", c.Database.Name)
	c.App.APIURL = getenv("API_URL", c.App.APIURL)

	if v := getenv("DB_MAX_IDLE_CONNS", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.Database.MaxIdleConns = parsed
		}
	}
	if v := getenv("DB_MAX_OPEN_CONNS", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.Database.MaxOpenConns = parsed
		}
	}
	c.Database.MaxIdletime = getenv("DB_MAX_IDLE_TIME", c.Database.MaxIdletime)
	c.Email.Host = getenv("MAILTRAP_HOST", c.Email.Host)
	if v := getenv("MAILTRAP_PORT", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			c.Email.Port = parsed
		}
	}
	c.Email.Username = getenv("MAILTRAP_USERNAME", c.Email.Username)
	c.Email.Password = getenv("MAILTRAP_PASSWORD", c.Email.Password)
	c.Email.FromEmail = getenv("FROM_EMAIL", c.Email.FromEmail)
	c.App.Audience = getenv("JWT_AUDIENCE", c.App.Audience)
	c.App.Issuer = getenv("JWT_ISSUER", c.App.Issuer)
	c.Auth.SecretKey = getenv("JWT_SECRET_KEY", c.Auth.SecretKey)

	return c, nil
}

func getenv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
