package main

import (
	"context"
	"fmt"
	"github.com/chapsuk/grace"
	"github.com/gorilla/mux"
	"go.sport-news/internal/config"
	ll "go.sport-news/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoadFromYAML("./e2e/e2e.yml")
	logger := ll.MustLoad("", cfg.Env, cfg.Logger.Level)
	ctx := grace.ShutdownContext(context.Background())

	defer logger.Sync() //nolint:errcheck

	r := mux.NewRouter()

	r.HandleFunc("/getnewsarticleinformation", func(w http.ResponseWriter, r *http.Request) {
		serveXMLFile(w, "./e2e/testdata/article.xml")
	}).Methods("GET")

	r.HandleFunc("/getnewlistinformation", func(w http.ResponseWriter, r *http.Request) {
		serveXMLFile(w, "./e2e/testdata/listArticles.xml")
	}).Methods("GET")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.ExternalPort),
		Handler: r,
	}

	e := make(chan error, 1)
	go func() {
		e <- srv.ListenAndServe()
	}()

	logger.Info(
		"HTTP server is running",
		zap.Int("port", cfg.HTTP.ExternalPort),
	)

	select {
	case <-ctx.Done():
		if err := srv.Shutdown(ctx); err != nil {
			logger.Fatal("shutdown err", zap.Error(err))
			grace.ShutdownContext(ctx)
		}
	case err := <-e:
		logger.Fatal("err", zap.Error(err))
		grace.ShutdownContext(ctx)
	}
}

func serveXMLFile(w http.ResponseWriter, filename string) {
	xmlContent, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/xml")

	if _, err := w.Write(xmlContent); err != nil {
		http.Error(w, "Unable to write XML content", http.StatusInternalServerError)
		return
	}
}
