package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DB DB `yaml:"psql_storage"`
}

type DB struct {
	Driver   string `yaml:"db_driver"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func main() {
	var action string

	flag.StringVar(&action, "action", "", "up or down")
	flag.Parse()

	if !(action == "up" || action == "down") {
		panic("invalid action: " + action)
	}

	migrationsPath, configPath := mustLoadEnv()

	db := mustLoadConfig(configPath)

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	databaseURL := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		db.Driver,
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.DBName,
		db.SSLMode,
	)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(fmt.Sprintf("can't get new migrations: %v", err))
	}

	switch action {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Print("no migrations to apply")
				return
			}

			panic(fmt.Sprintf("can't up new migrations: %v", err))
		}
	case "down":
		if err := m.Down(); err != nil {
			panic(fmt.Sprintf("can't down migrations: %v", err))
		}
	}

	log.Print("migrations applied successfully")
}

func mustLoadEnv() (string, string) {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	migrationsPath := os.Getenv("MIGRATION_PATH")

	if migrationsPath == "" {
		panic("migration path is empty")
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		panic("config path is empty")
	}

	return migrationsPath, configPath
}

func mustLoadConfig(path string) *DB {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist")
	}

	cfg := Config{}

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		panic("failed to read config")
	}

	cfg.DB.Password = os.Getenv("POSTGRES_PASSWORD")

	return &cfg.DB
}
