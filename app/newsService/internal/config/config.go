package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string     `yaml:"env" env-required:"true"`
	UserStorage Postgres   `yaml:"psql_storage"`
	GRPC        GRPCConfig `yaml:"grpc_news"`
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

	cfg.UserStorage.Password = os.Getenv("POSTGRES_PASSWORD")
	if cfg.UserStorage.Password == "" {
		panic("postgress password is not specified in environment variables")
	}

	return cfg
}

func fetchConfigPath() (string, error) {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		if err := godotenv.Load(); err != nil {
			return "", fmt.Errorf("can't load config: %v", err)
		}
		path = os.Getenv("CONFIG_PATH")
	}

	if path == "" {
		return "", fmt.Errorf("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist")
	}

	return path, nil
}
