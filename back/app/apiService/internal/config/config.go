package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env          string         `yaml:"env" env-required:"true"`
	Server       ApiServer      `yaml:"api_server"`
	AuthGRPC     GRPCConfig     `yaml:"grpc_auth"`
	NewsGRPC     GRPCConfig     `yaml:"grpc_news"`
	Cache        Redis          `yaml:"redis_storage"`
	Manager      ArticleManager `yaml:"news_managment"`
	TokenManager TokenManager   `yaml:"token_managment"`
}

type ApiServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"90s"`
}

type GRPCConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retries_count"`
}

type ArticleManager struct {
	ArticlesLimit   int           `yaml:"articles_limit"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type TokenManager struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
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
