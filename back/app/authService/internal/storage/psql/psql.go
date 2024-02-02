package psql

import (
	"database/sql"
	"fmt"

	"newsWebApp/app/authService/internal/config"

	_ "github.com/lib/pq"
)

func New(cfg config.Postgres) (*sql.DB, error) {
	// dbdriver://username:password@host:port/dbname?param1=true&param2=false
	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Driver, cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
