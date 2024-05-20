package main

import (
	"context"
	"github.com/chapsuk/grace"
	"github.com/patrickmn/go-cache"
	"go.sport-news/internal/config"
	v1 "go.sport-news/internal/controller/http/v1"
	"go.sport-news/internal/database"
	"go.sport-news/internal/environment"
	"go.sport-news/internal/http"
	ll "go.sport-news/internal/logger"
	"go.sport-news/internal/repository"
	"go.sport-news/internal/scheduler"
	"go.uber.org/zap"
	"time"
)

//nolint:gochecknoglobals
const (
	version = "unknown"
)

func main() {
	cfg := config.MustLoad()

	logger := ll.MustLoad(version, cfg.Env, cfg.Logger.Level)
	defer logger.Sync() //nolint:errcheck

	ctx := grace.ShutdownContext(context.Background())
	ctx = environment.CtxWithEnv(ctx, cfg.Env)

	db := database.MustLoad(ctx, logger, cfg.Mongo)
	defer db.Disconnect(ctx)

	if cfg.Parser.Enable == 1 {
		scheduler.New(logger, cfg.Parser, db)
	}

	httpServer := http.New(
		v1.NewNewsController(
			repository.NewNewsRepository(db),
			cache.New(1*time.Minute, 5*time.Minute),
			logger,
		),
		logger,
		&cfg.HTTP,
	)
	if err := httpServer.Serve(ctx); err != nil {
		logger.Fatal("http server fatal", zap.Error(err))
	}

	grace.ShutdownContext(ctx)
}
