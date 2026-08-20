package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakePublisher struct {
	calls []publishCall
	err   error
}

type publishCall struct {
	channel string
	payload string
}

func (f *fakePublisher) Publish(_ context.Context, channel string, payload []byte) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, publishCall{channel: channel, payload: string(payload)})
	return nil
}

func TestNewRoutesPayloadToMappedChannel(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	handler := New(
		map[string]string{"/webhook/github": "github-events"},
		publisher,
		BasicAuthConfig{Enabled: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader(`{"event":"push"}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, res.Code)
	}

	if len(publisher.calls) != 1 {
		t.Fatalf("expected one publish call, got %d", len(publisher.calls))
	}

	if publisher.calls[0].channel != "github-events" {
		t.Fatalf("expected channel github-events, got %q", publisher.calls[0].channel)
	}
	if publisher.calls[0].payload != `{"event":"push"}` {
		t.Fatalf("expected payload to match request body, got %q", publisher.calls[0].payload)
	}
}

func TestNewRequiresBasicAuthByDefaultConfig(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	handler := New(
		map[string]string{"/webhook/github": "github-events"},
		publisher,
		BasicAuthConfig{Enabled: true, Username: "user", Password: "pass"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	unauthorizedReq := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader("{}"))
	unauthorizedRes := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRes, unauthorizedReq)

	if unauthorizedRes.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status, got %d", unauthorizedRes.Code)
	}

	authorizedReq := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader("{}"))
	authorizedReq.SetBasicAuth("user", "pass")
	authorizedRes := httptest.NewRecorder()
	handler.ServeHTTP(authorizedRes, authorizedReq)

	if authorizedRes.Code != http.StatusAccepted {
		t.Fatalf("expected accepted status for authorized request, got %d", authorizedRes.Code)
	}
}

func TestPublishFailureReturnsBadGateway(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{err: errors.New("publish failed")}
	handler := New(
		map[string]string{"/webhook/github": "github-events"},
		publisher,
		BasicAuthConfig{Enabled: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	req := httptest.NewRequest(http.MethodPost, "/webhook/github", strings.NewReader("{}"))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected bad gateway status, got %d", res.Code)
	}
}

func TestUnmappedRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	handler := New(
		map[string]string{"/webhook/github": "github-events"},
		publisher,
		BasicAuthConfig{Enabled: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	req := httptest.NewRequest(http.MethodPost, "/webhook/unknown", strings.NewReader("{}"))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected not found status, got %d", res.Code)
	}

	if len(publisher.calls) != 0 {
		t.Fatalf("expected no publish calls, got %d", len(publisher.calls))
	}
}

func TestMappedRouteRejectsNonPostMethods(t *testing.T) {
	t.Parallel()

	publisher := &fakePublisher{}
	handler := New(
		map[string]string{"/webhook/github": "github-events"},
		publisher,
		BasicAuthConfig{Enabled: false},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	req := httptest.NewRequest(http.MethodGet, "/webhook/github", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed status, got %d", res.Code)
	}

	if len(publisher.calls) != 0 {
		t.Fatalf("expected no publish calls, got %d", len(publisher.calls))
	}
}
