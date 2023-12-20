package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"newsWebApp/app/webService/internal/config"
	"newsWebApp/app/webService/internal/http/handler"
	"newsWebApp/app/webService/internal/http/server"
	"newsWebApp/app/webService/internal/logs"
	"newsWebApp/app/webService/internal/services/authgrpc"
	"newsWebApp/app/webService/internal/services/fetcher"
	"newsWebApp/app/webService/internal/services/newsgrpc"
	"newsWebApp/app/webService/internal/storage/cache"
)

type App struct {
	cfg     *config.Config
	log     *slog.Logger
	cache   *cache.Cache
	fetcher *fetcher.NewsFetcher
	srv     *server.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.log.With("service", "Web")

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	authClient, err := authgrpc.New(ctx,
		a.log,
		a.cfg.AuthGRPC.Port,
		a.cfg.AuthGRPC.Timeout,
		a.cfg.AuthGRPC.RetriesCount,
	)
	if err != nil {
		a.log.Error("Failed to create new auth client", "err", err)
		os.Exit(1)
	}

	newsClient, err := newsgrpc.New(ctx,
		a.log,
		a.cfg.NewsGRPC.Port,
		a.cfg.NewsGRPC.Timeout,
		a.cfg.NewsGRPC.RetriesCount,
	)
	if err != nil {
		a.log.Error("Failed to create new news client", "err", err)
		os.Exit(1)
	}

	a.cache, err = cache.New(ctx,
		a.cfg.Cache.Host,
		a.cfg.Cache.Port,
		a.cfg.Manager.ArticlesLimit,
		a.cfg.Server.TemplatesPath,
	)
	if err != nil {
		a.log.Error("Failed to create new articles cache", "err", err)
		os.Exit(1)
	}

	a.fetcher = fetcher.New(newsClient, a.cache, a.cfg.Manager.RefreshInterval, a.log)

	handler := handler.New(authClient, newsClient, a.fetcher, a.log)

	a.srv = server.New(handler, &a.cfg.Server, a.log)

	return &a
}

func (a *App) MustRun() {
	a.log.Info("Starting web service", "env", a.cfg.Env, "address", a.cfg.Server.Address)

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	go func() {
		if err := a.fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				a.log.Error("Failed ower working fetcher in web service", "err store", err.Error())
				os.Exit(1)
			}
		}
	}()

	go func() {
		if err := a.srv.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.log.Error("Failed ower working web service", "err", err.Error())
				os.Exit(1)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	a.mustStop()
}

func (a *App) mustStop() {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.Timeout)
	defer cancel()

	if err := a.srv.Stop(ctx); err != nil {
		a.log.Error("Closing connection to web server", "err store", err.Error())
	}

	if err := a.cache.CloseConn(); err != nil {
		a.log.Error("Closing connection to articles cache", "err store", err.Error())
	}

	a.log.Info("Web service stoped gracefully")
}
