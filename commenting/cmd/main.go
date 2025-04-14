package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiremitrov123/onboarding/commenting/api"
	"github.com/kiremitrov123/onboarding/commenting/db"
	"github.com/kiremitrov123/onboarding/commenting/redis"
	"github.com/kiremitrov123/onboarding/commenting/service"
)

type Config struct {
	DBURL     string
	RedisAddr string
	HTTPAddr  string
}

func loadConfig() Config {
	return Config{
		DBURL:     getEnv("DATABASE_URL", "postgresql://root@localhost:26257/commenting?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR", "redis:6379"),
		HTTPAddr:  getEnv("HTTP_ADDR", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := loadConfig()

	pg, err := db.NewPostgres(ctx, cfg.DBURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	redisCache, err := redis.NewCache(ctx, cfg.RedisAddr)
	if err != nil {
		logger.Error("failed to connect to Redis", slog.Any("error", err))
		os.Exit(1)
	}

	repo := db.NewRepo(pg.DB())
	svc := service.NewCommentService(repo, redisCache)
	apiHandler := api.NewAPI(svc, logger)

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apiHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down gracefully")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("forced shutdown", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	logger.Info("commenting API running", slog.String("addr", cfg.HTTPAddr))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("could not start server", slog.Any("error", err))
		os.Exit(1)
	}
}
