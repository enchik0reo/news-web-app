package migrator

import (
	"errors"
	"fmt"
	"os"
	"time"

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

func Up() error {
	const op = "migrator.Up"

	migrationsPath, configPath, err := loadEnv()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	db, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

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

	m, err := newMigraor(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func newMigraor(sourceURL string, databaseURL string) (*migrate.Migrate, error) {
	var err error
	var m *migrate.Migrate

	for i := 1; i <= 5; i++ {
		m, err = migrate.New(sourceURL, databaseURL)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return m, nil
}

func loadEnv() (string, string, error) {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	migrationsPath := os.Getenv("MIGRATION_PATH")

	if migrationsPath == "" {
		return "", "", fmt.Errorf("migration path is empty")
	}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		return "", "", fmt.Errorf("config path is empty")
	}

	return migrationsPath, configPath, nil
}

func loadConfig(path string) (*DB, error) {
	cfg := Config{}

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w ", err)
	}

	cfg.DB.Password = os.Getenv("POSTGRES_PASSWORD")

	return &cfg.DB, nil
}
