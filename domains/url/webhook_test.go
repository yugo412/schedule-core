package url

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendWebhookSuccess(t *testing.T) {
	var received WebhookPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	err := SendWebhook(server.URL, WebhookPayload{
		Slug:       "event",
		URL:        "https://example.com",
		Status:     "broken",
		StatusCode: http.StatusNotFound,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received.Slug != "event" {
		t.Errorf("expected slug event, got %s", received.Slug)
	}

	if received.StatusCode != http.StatusNotFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNotFound,
			received.StatusCode,
		)
	}
}

func TestSendWebhookErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	err := SendWebhook(server.URL, WebhookPayload{Slug: "event"})

	if err == nil {
		t.Fatal("expected error for non-2xx webhook response")
	}
}
