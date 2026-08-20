package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

	redisDB := 0
	if rawDB := os.Getenv("REDIS_DB"); rawDB != "" {
		parsedDB, parseErr := strconv.Atoi(rawDB)
		if parseErr != nil {
			logger.Error("invalid REDIS_DB", "error", parseErr)
			os.Exit(1)
		}
		redisDB = parsedDB
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})
	defer redisClient.Close()

	authEnabled := true
	if rawEnabled := os.Getenv("BASIC_AUTH_ENABLED"); rawEnabled != "" {
		parsedEnabled, parseErr := strconv.ParseBool(rawEnabled)
		if parseErr != nil {
			logger.Error("invalid BASIC_AUTH_ENABLED", "error", parseErr)
			os.Exit(1)
		}
		authEnabled = parsedEnabled
	}

	authCfg := server.BasicAuthConfig{
		Enabled:  authEnabled,
		Username: os.Getenv("BASIC_AUTH_USERNAME"),
		Password: os.Getenv("BASIC_AUTH_PASSWORD"),
	}

	if authCfg.Enabled && (authCfg.Username == "" || authCfg.Password == "") {
		logger.Error("basic auth is enabled but BASIC_AUTH_USERNAME or BASIC_AUTH_PASSWORD is empty")
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

	if shutdownErr := httpServer.Shutdown(ctx); shutdownErr != nil {
		logger.Error("shutdown failed", "error", shutdownErr)
		os.Exit(1)
	}
}
