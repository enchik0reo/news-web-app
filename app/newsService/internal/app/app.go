package app

// TODO
/* type NewsService interface {
	SaveArticle(ctx context.Context)          // Чтоб сохранять
	GetPostedArticles(ctx context.Context)    // Чтоб отображать на главной странице опубликованные новости
	AllNotPostedArticles(ctx context.Context) // Чтоб смотреть не опубликованные новости
	MarkArticlePosted(ctx context.Context)    // Чтоб пометить новость как опубликованную
} */

type App struct {
}

func New() *App {
	a := App{}

	return &a
}

func (a *App) MustRun() {

}
