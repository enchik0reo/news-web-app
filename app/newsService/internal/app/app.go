package app

import (
	"log/slog"

	"newsWebApp/app/newsService/internal/config"
	"newsWebApp/app/newsService/internal/logs"
)

// TODO
/* type NewsService interface {
	SaveArticle(ctx context.Context)          // Чтоб сохранять
	GetPostedArticles(ctx context.Context)    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles(ctx context.Context) // Чтоб смотреть не опубликованные новости
	MarkArticlePosted(ctx context.Context)    // Чтоб пометить новость как опубликованную
} */

type App struct {
	cfg *config.Config
	log *slog.Logger
}

func New() *App {
	a := App{}

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	return &a
}

func (a *App) MustRun() {

}
