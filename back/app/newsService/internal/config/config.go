package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string      `yaml:"env" env-required:"true"`
	Storage     Postgres    `yaml:"psql_storage"`
	LinkStorage Redis       `yaml:"redis_storage"`
	GRPC        GRPCConfig  `yaml:"grpc_news"`
	Manager     NewsManager `yaml:"news_managment"`
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
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type NewsManager struct {
	FilterKeywords []string      `yaml:"filter_keywords"`
	FetchInterval  time.Duration `yaml:"fetch_interval"`
	ArticlesLimit  int           `yaml:"articles_limit"`
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
