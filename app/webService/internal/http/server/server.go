package server

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"

	"newsWebApp/app/webService/internal/config"
)

type Server struct {
	cfg    *config.HttpServer
	log    *slog.Logger
	server *http.Server
}

func New(handler http.Handler, c *config.HttpServer, l *slog.Logger) *Server {
	srv := newHTTPServer(handler, c)

	return &Server{
		cfg:    c,
		log:    l,
		server: srv,
	}
}

func newHTTPServer(handler http.Handler, cfg *config.HttpServer) *http.Server {
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
