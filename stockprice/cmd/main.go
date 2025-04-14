package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/api"
	"github.com/kiremitrov123/onboarding/stockprice/cache"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Set up local and Redis caches
	localCache := cache.NewLocalCache(5*time.Minute, 10*time.Minute)
	redisCache, err := cache.NewRedisCache(ctx, "redis:6379")
	if err != nil {
		log.Fatal("could not connect to Redis", err)
	}

	// invalidate local cache when new values arrive
	err = redisCache.SubscribeInvalidation(ctx, func(key string) {
		localCache.Invalidate(key)
	})
	if err != nil {
		log.Fatalf("failed to subscribe to invalidation: %v", err)
	}

	api := &api.API{
		Local: localCache,
		Redis: redisCache,
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      api,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutdown signal received, shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("server forced to shutdown: %v", err)
		}
	}()

	log.Println("StockPrice API is running on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("could not start server: %v", err)
	}
}
