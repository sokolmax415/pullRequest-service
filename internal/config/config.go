package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Server struct {
		Port string `env:"SERVER_PORT" env-default:"8080"`
	} `yaml:"server"`

	Database struct {
		Host     string `env:"DB_HOST" env-default:"localhost"`
		Port     string `env:"DB_PORT" env-default:"5432"`
		Name     string `env:"DB_NAME" env-default:"pr_service"`
		User     string `env:"DB_USER" env-default:"postgres"`
		Password string `env:"DB_PASSWORD" env-default:"password"`
	} `yaml:"database"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
