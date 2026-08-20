package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/its-the-vibe/VibeHook/internal/config"
	"github.com/its-the-vibe/VibeHook/internal/server"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddress,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	authCfg := server.BasicAuthConfig{
		Enabled:  cfg.BasicAuthEnabled,
		Username: cfg.BasicAuthUsername,
		Password: os.Getenv("BASIC_AUTH_PASSWORD"),
	}

	if authCfg.Enabled && authCfg.Password == "" {
		logger.Error("basic auth is enabled but BASIC_AUTH_PASSWORD is empty")
		os.Exit(1)
	}

	handler := server.New(cfg.Routes, server.NewRedisPublisher(redisClient), authCfg, logger)

	httpServer := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", cfg.ListenAddress)
		if serveErr := httpServer.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			logger.Error("server exited", "error", serveErr)
			os.Exit(1)
		}
	}()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)
	<-shutdownSignal

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	if shutdownErr := httpServer.Shutdown(ctx); shutdownErr != nil {
		logger.Error("shutdown failed", "error", shutdownErr)
		os.Exit(1)
	}
}
