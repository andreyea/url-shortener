package main

import (
	"net/http"
	"os"

	"log/slog"

	"github.com/andreyea/url-shortener/internal/config"
	"github.com/andreyea/url-shortener/internal/http-server/handlers/url/save"
	mwLogger "github.com/andreyea/url-shortener/internal/http-server/middleware/logger"
	"github.com/andreyea/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/andreyea/url-shortener/internal/lib/logger/sl"
	"github.com/andreyea/url-shortener/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting url-shortener", "env", cfg.Env)

	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))

	router.Use(middleware.Recoverer)

	router.Use(middleware.URLFormat)

	_ = storage

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("urk-shortener", map[string]string{cfg.HttpServer.User: cfg.HttpServer.Password}))
		r.Post("/", save.New(log, storage))
	})

	//router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.HttpServer.Address))

	srv := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug},
		))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
