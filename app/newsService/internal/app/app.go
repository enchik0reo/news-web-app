package app

import "newsWebApp/app/newsService/internal/config"

// TODO
/* type NewsService interface {
	SaveArticle(ctx context.Context)          // Чтоб сохранять
	GetPostedArticles(ctx context.Context)    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles(ctx context.Context) // Чтоб смотреть не опубликованные новости
	MarkArticlePosted(ctx context.Context)    // Чтоб пометить новость как опубликованную
} */

type App struct {
	cfg *config.Config
}

func New() *App {
	a := App{}

	a.cfg = config.MustLoad()

	return &a
}

func (a *App) MustRun() {

}
