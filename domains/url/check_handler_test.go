package url

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckURLHealthy(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	defer server.Close()

	result := CheckURL(server.URL)

	if result.Status != "healthy" {
		t.Errorf("expected healthy, got %s", result.Status)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusOK,
			result.StatusCode,
		)
	}
}

func TestCheckURLBroken404(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	defer server.Close()

	result := CheckURL(server.URL)

	if result.Status != "broken" {
		t.Errorf("expected broken, got %s", result.Status)
	}

	if result.StatusCode != http.StatusNotFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNotFound,
			result.StatusCode,
		)
	}
}

func TestCheckURLBroken410(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusGone)
		}),
	)

	defer server.Close()

	result := CheckURL(server.URL)

	if result.Status != "broken" {
		t.Errorf("expected broken, got %s", result.Status)
	}

	if result.StatusCode != http.StatusGone {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusGone,
			result.StatusCode,
		)
	}
}

func TestCheckURLWarning500(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	)

	defer server.Close()

	result := CheckURL(server.URL)

	if result.Status != "warning" {
		t.Errorf("expected warning, got %s", result.Status)
	}

	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			result.StatusCode,
		)
	}
}

func TestCheckURLNetworkError(t *testing.T) {
	result := CheckURL("http://invalid.invalid")

	if result.Status != "warning" {
		t.Errorf("expected warning, got %s", result.Status)
	}

	if result.Error == "" {
		t.Error("expected error message")
	}
}
