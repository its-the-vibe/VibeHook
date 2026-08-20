# VibeHook

![CI](https://github.com/its-the-vibe/VibeHook/actions/workflows/ci.yaml/badge.svg)

VibeHook is a lightweight webhook consumer service written in Go. It listens on configured paths and publishes incoming webhook payloads to Redis channels.

## Features

- Path-to-channel webhook routing via YAML config
- Redis publish using `github.com/redis/go-redis/v9`
- HTTP Basic Authentication enabled by default
- Distroless production image
- Docker Compose service with read-only filesystem

## Quick Start

1. Copy example configuration files:

   ```bash
   cp config.example.yaml config.yaml
   cp .env.example .env
   ```

2. Edit `config.yaml` with your route-to-channel mappings and non-sensitive runtime settings.
3. Set secure values in `.env` for `REDIS_PASSWORD` and `BASIC_AUTH_PASSWORD`.
4. Run locally:

   ```bash
   make run
   ```

## Configuration

### `config.yaml`

`config.yaml` contains only non-sensitive settings.

```yaml
listenAddress: ":8080"
redisAddress: "redis.example.internal:6379"
redisDB: 0
basicAuthEnabled: true
basicAuthUsername: vibehook
routes:
  /webhook/github: github-events
  /webhook/slack: slack-events
  /webhook/custom: custom-webhook
```

### `.env`

Use environment variables for sensitive configuration only.

```env
REDIS_PASSWORD=replace-with-redis-password
BASIC_AUTH_PASSWORD=replace-with-strong-password
```

## Usage

With the default config above:

- `POST /webhook/github` publishes to `github-events`
- `POST /webhook/slack` publishes to `slack-events`
- `POST /webhook/custom` publishes to `custom-webhook`

Example request:

```bash
curl -X POST \
  -u vibehook:replace-with-strong-password \
  -H 'Content-Type: application/json' \
  -d '{"event":"push"}' \
  http://localhost:8080/webhook/github
```

## Docker

Run with Docker Compose (expects external Redis):

```bash
docker compose up --build
```

The Compose service uses a read-only filesystem and mounts `config.yaml` as read-only.

## Development

```bash
make tidy
make test
make build
```
