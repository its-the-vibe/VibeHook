package server

import (
	"context"
	"crypto/subtle"
	"io"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type Publisher interface {
	Publish(ctx context.Context, channel string, payload []byte) error
}

type RedisPublisher struct {
	client *redis.Client
}

func NewRedisPublisher(client *redis.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (r *RedisPublisher) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}

type BasicAuthConfig struct {
	Enabled  bool
	Username string
	Password string
}

func New(routes map[string]string, publisher Publisher, authCfg BasicAuthConfig, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	for path, channel := range routes {
		mux.Handle(path, webhookHandler(channel, publisher, logger))
	}

	if authCfg.Enabled {
		return basicAuth(authCfg, mux)
	}

	return mux
}

func webhookHandler(channel string, publisher Publisher, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		if err := publisher.Publish(r.Context(), channel, payload); err != nil {
			logger.Error("failed to publish payload", "channel", channel, "error", err)
			http.Error(w, "failed to publish payload", http.StatusBadGateway)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})
}

func basicAuth(cfg BasicAuthConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="vibehook"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Username))
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Password))
		authorized := userMatch&passwordMatch == 1
		if !authorized {
			w.Header().Set("WWW-Authenticate", `Basic realm="vibehook"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
