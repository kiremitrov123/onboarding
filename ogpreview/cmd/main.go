package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/kiremitrov123/onboarding/src/ogpreview/api"
	"github.com/kiremitrov123/onboarding/src/ogpreview/redis"
)

func initTracer(ctx context.Context, service string) (func(context.Context) error, error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(os.Getenv("JAEGER_PORT")),
		otlptracegrpc.WithInsecure(), // for local testing
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	shutdown, err := initTracer(ctx, os.Getenv("SERVICE_NAME"))
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			logger.Error("error shutting down tracer", "error", err)
		}
	}()

	rdb, err := redis.NewClient(ctx, "redis:6379")
	if err != nil {
		logger.Error("could not connect to Redis", "error", err)
		os.Exit(1)
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "OGFetch",
		Timeout:     5 * time.Second,
		MaxRequests: 5,
	})

	apiHandler := &api.API{
		Cache: rdb,
		CB:    cb,
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      apiHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received, shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server forced to shutdown", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("OG Preview API is running on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("could not start server", "error", err)
		os.Exit(1)
	}
}
