package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string       `yaml:"env" env-required:"true"`
	Storage        Postgres     `yaml:"psql_storage"`
	SessionStorage Redis        `yaml:"redis_storage"`
	RegistrStorage Memcached    `yaml:"memcached_storage"`
	Manager        TokenManager `yaml:"token_managment"`
	GRPC           GRPCConfig   `yaml:"grpc_auth"`
}

type Postgres struct {
	Driver   string `yaml:"db_driver"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type Redis struct {
	Host   string        `yaml:"host"`
	Port   string        `yaml:"port"`
	Expire time.Duration `yaml:"expire"`
}

type Memcached struct {
	Host    string        `yaml:"host"`
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type TokenManager struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	SecretKey       string        `env:"SECRET_KEY"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	cfg := new(Config)

	path, err := fetchConfigPath()
	if err != nil {
		panic(err)
	}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	cfg.Storage.Password = os.Getenv("POSTGRES_PASSWORD")
	if cfg.Storage.Password == "" {
		panic("postgress password is not specified in environment variables")
	}

	cfg.Manager.SecretKey = os.Getenv("SECRET_KEY")
	if cfg.Manager.SecretKey == "" {
		panic("secret key is not specified in environment variables")
	}

	return cfg
}

func fetchConfigPath() (string, error) {
	if err := godotenv.Load(); err != nil {
		return "", fmt.Errorf("can't load config: %v", err)
	}

	path := os.Getenv("CONFIG_PATH")

	if path == "" {
		return "", fmt.Errorf("config path is empty")
	}

	return path, nil
}
