package server

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

	"newsWebApp/app/apiService/internal/config"
)

type Server struct {
	cfg    *config.ApiServer
	log    *slog.Logger
	server *http.Server
}

func New(handler http.Handler, c *config.ApiServer, l *slog.Logger) *Server {
	srv := setupServer(handler, c)

	return &Server{
		cfg:    c,
		log:    l,
		server: srv,
	}
}

func setupServer(handler http.Handler, cfg *config.ApiServer) *http.Server {
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	return &http.Server{
		Addr:           cfg.Address,
		Handler:        handler,
		TLSConfig:      tlsConfig,
		ReadTimeout:    cfg.Timeout,
		WriteTimeout:   cfg.Timeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: 524288,
	}
}

func (s *Server) Start() error {
	s.log.Info("Web server is running", "address", s.cfg.Address)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
