package env

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert/yaml"
)

type config struct {
	App struct {
		Name string `yaml:"name"`
		ENV  string `yaml:"env"`
	} `yaml:"app"`

	Server struct {
		Port string `yaml:""`
	} `yaml:"server"`

	Database struct {
		Addr         string `yaml:""`
		Host         string `yaml:""`
		User         string `yaml:""`
		Password     string `yaml:""`
		Name         string `yaml:""`
		MaxIdleConns int    `yaml:"MAX_IDLE_CONNS"`
		MaxOpenConns int    `yaml:"MAX_OPEN_CONNS"`
		MaxIdletime  string `yaml:"MAX_IDLE_TIME"`
	}
}

func Load() (*config, error) {

	_ = godotenv.Load()

	c := &config{}

	env := "dev"
	file := "./internal/env/config." + env + ".yaml"

	cfg, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cfg, c)
	if err != nil {
		return nil, err
	}

	c.Server.Port = getenv("PORT", "")
	c.Database.Addr = getenv("DB_PORT", "")
	c.Database.Host = getenv("DB_HOST", "")
	c.Database.User = getenv("DB_USER", "")
	c.Database.Password = getenv("DB_PASSWORD", "")
	c.Database.Name = getenv("DB_NAME", "")

	return c, nil
}

func getenv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
