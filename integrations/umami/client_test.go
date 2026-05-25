package umami

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupClient(serverURL string) *Client {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	return NewClient(
		serverURL,
		"website-id",
		logger,
	)
}

func TestTrackEventSuccess(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if r.URL.Path != "/api/send" {
				t.Errorf(
					"expected path /api/send, got %s",
					r.URL.Path,
				)
			}

			if r.Method != http.MethodPost {
				t.Errorf(
					"expected POST, got %s",
					r.Method,
				)
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			w.Write([]byte(`
			{
				"cache": "token",
				"sessionId": "session-id",
				"visitId": "visit-id"
			}
			`))
		}),
	)

	defer server.Close()

	client := setupClient(server.URL)

	err := client.TrackEvent(Event{
		Name:      "registration-click",
		URL:       "/official/mantra-run-2026",
		Hostname:  "example.com",
		Language:  "id-ID",
		UserAgent: "Mozilla/5.0",
		Data: map[string]any{
			"slug": "mantra-run-2026",
		},
	})

	if err != nil {
		t.Errorf(
			"expected nil error, got %v",
			err,
		)
	}
}

func TestTrackEventRejected(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			w.Write([]byte(`
			{
				"beep": "boop"
			}
			`))
		}),
	)

	defer server.Close()

	client := setupClient(server.URL)

	err := client.TrackEvent(Event{
		Name:      "registration-click",
		URL:       "/official/mantra-run-2026",
		Hostname:  "example.com",
		Language:  "id-ID",
		UserAgent: "Go-http-client/1.1",
	})

	if err == nil {
		t.Errorf(
			"expected error, got nil",
		)
	}
}

func TestTrackEventServerError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.WriteHeader(
				http.StatusInternalServerError,
			)
		}),
	)

	defer server.Close()

	client := setupClient(server.URL)

	err := client.TrackEvent(Event{
		Name:      "registration-click",
		URL:       "/official/mantra-run-2026",
		Hostname:  "example.com",
		Language:  "id-ID",
		UserAgent: "Mozilla/5.0",
	})

	if err == nil {
		t.Errorf(
			"expected error, got nil",
		)
	}
}
