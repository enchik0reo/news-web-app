package psql

import (
	"database/sql"
	"fmt"

	"newsWebApp/app/newsService/internal/config"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.Postgres) (*Storage, error) {
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

	s := Storage{db: db}

	return &s, nil
}

func (s *Storage) CloseConn() error {
	return s.db.Close()
}

// TODO
/* type NewsService interface {
	SaveArticle(ctx context.Context)          // Чтоб сохранять
	GetPostedArticles(ctx context.Context)    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles(ctx context.Context) // Чтоб смотреть не опубликованные новости
	MarkArticlePosted(ctx context.Context)    // Чтоб пометить новость как опубликованную
} */

func (s *Storage) A() {

}
