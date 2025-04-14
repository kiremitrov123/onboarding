package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kiremitrov123/onboarding/texteditor/api"
	"github.com/kiremitrov123/onboarding/texteditor/redis"
	"github.com/kiremitrov123/onboarding/texteditor/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rdb, err := redis.NewRedisClient(ctx, "redis:6379")
	if err != nil {
		log.Fatal("could not connect to Redis", err)
	}

	svcEditor := service.NewEditorService(rdb)
	apiHandler := api.NewAPI(svcEditor)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      apiHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced to shutdown: %v\n", err)
			os.Exit(1)
		}
	}()

	log.Println("Text Editor API running on :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
