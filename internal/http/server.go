package http

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"go.sport-news/internal/config"
	v1 "go.sport-news/internal/controller/http/v1"
	"go.sport-news/internal/environment"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	logger         *zap.Logger
	config         *config.Http
	newsController v1.INewsController
	srv            *http.Server
}

func New(nC v1.INewsController, log *zap.Logger, cfg *config.Http) *Server {
	return &Server{
		logger:         log,
		config:         cfg,
		newsController: nC,
		srv: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			WriteTimeout: cfg.WriteTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Serve create and listen to http server.
func (s *Server) Serve(ctx context.Context) error {

	r := mux.NewRouter()

	r.HandleFunc("/v1/teams/{team}/news", s.newsController.GetTeamNews).Methods("GET")
	r.HandleFunc("/v1/teams/{team}/news/{id}", s.newsController.GetTeamNewsByID).Methods("GET")

	if environment.EnvFromCtx(ctx).IsLocal() {
		r.HandleFunc("/v1/cache-flush", s.newsController.ResetCache).Methods("POST")
	}

	s.srv.Handler = r

	e := make(chan error, 1)
	go func() {
		e <- s.srv.ListenAndServe()
	}()

	s.logger.Info(
		"HTTP server is running",
		zap.Int("port", s.config.Port),
	)

	select {
	case <-ctx.Done():
		if err := s.srv.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	case err := <-e:
		return err
	}
}
