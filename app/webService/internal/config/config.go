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
	Env      string         `yaml:"env" env-required:"true"`
	Server   HttpServer     `yaml:"http_server"`
	AuthGRPC GRPCConfig     `yaml:"grpc_auth"`
	NewsGRPC GRPCConfig     `yaml:"grpc_news"`
	Cache    Redis          `yaml:"redis_storage"`
	Manager  ArticleManager `yaml:"news_managment"`
}

type HttpServer struct {
	Address       string        `yaml:"address" env-default:"localhost:8080"`
	Timeout       time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout   time.Duration `yaml:"idle_timeout" env-default:"90s"`
	TemplatesPath string        `yaml:"templates_path"`
}

type GRPCConfig struct {
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
