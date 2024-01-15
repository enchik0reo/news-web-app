package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"newsWebApp/app/apiService/internal/config"
	"newsWebApp/app/apiService/internal/logs"
	"newsWebApp/app/apiService/internal/server/handler"
	"newsWebApp/app/apiService/internal/server/server"
	"newsWebApp/app/apiService/internal/services/authgrpc"
	"newsWebApp/app/apiService/internal/services/fetcher"
	"newsWebApp/app/apiService/internal/services/newsgrpc"
	"newsWebApp/app/apiService/internal/storage/cache"
)

type App struct {
	cfg     *config.Config
	log     *slog.Logger
	cache   *cache.Cache
	fetcher *fetcher.NewsFetcher
	handler http.Handler
	srv     *server.Server
}

func New() *App {
	a := App{}
	var err error

	a.cfg = config.MustLoad()

	a.log = logs.Setup(a.cfg.Env)

	a.log.With("service", "Api")

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	authClient, err := authgrpc.New(ctx,
		a.log,
		a.cfg.AuthGRPC.Port,
		a.cfg.AuthGRPC.Timeout,
		a.cfg.AuthGRPC.RetriesCount,
	)
	if err != nil {
		a.log.Error("Failed to create new auth client", "err", err.Error())
		os.Exit(1)
	}

	newsClient, err := newsgrpc.New(ctx,
		a.log,
		a.cfg.NewsGRPC.Port,
		a.cfg.NewsGRPC.Timeout,
		a.cfg.NewsGRPC.RetriesCount,
	)
	if err != nil {
		a.log.Error("Failed to create new news client", "err", err.Error())
		os.Exit(1)
	}

	a.cache, err = connectToCache(ctx,
		a.cfg.Cache.Host,
		a.cfg.Cache.Port,
		a.cfg.Manager.ArticlesLimit,
	)
	if err != nil {
		a.log.Error("Failed to create new articles cache", "err", err.Error())
		os.Exit(1)
	}

	a.fetcher = fetcher.New(newsClient, a.cache, a.cfg.Manager.RefreshInterval, a.log)

	a.handler, err = handler.New(authClient, newsClient, a.fetcher, a.cfg.TokenManager.RefreshTokenTTL, a.log)
	if err != nil {
		a.log.Error("Failed to create new handler", "err", err.Error())
		os.Exit(1)
	}

	a.srv = server.New(a.handler, &a.cfg.Server, a.log)

	return &a
}

func (a *App) MustRun() {
	a.log.Info("Starting api service", "env", a.cfg.Env, "address", a.cfg.Server.Address)

	ctx, cacnel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cacnel()

	go func() {
		if err := a.fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				a.log.Error("Failed ower working fetcher in api service", "err store", err.Error())
			}
		}
	}()

	go func() {
		if err := a.srv.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.log.Error("Failed ower working api service", "err", err.Error())
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
		a.log.Error("Closing connection to api server", "err store", err.Error())
	}

	if err := a.cache.CloseConn(); err != nil {
		a.log.Error("Closing connection to articles cache", "err store", err.Error())
	}

	a.log.Info("Api service stoped gracefully")
}

func connectToCache(ctx context.Context, host string, port string, limit int) (*cache.Cache, error) {
	var err error
	var c *cache.Cache

	for i := 1; i <= 5; i++ {
		c, err = cache.New(ctx, host, port, limit)
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return c, nil
}
