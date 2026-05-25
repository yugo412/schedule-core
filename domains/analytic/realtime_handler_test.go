package analytic

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/integrations/umami"
)

func setupHandler(
	serverURL string,
) *RealtimeHandler {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	umamiClient := umami.NewClient(
		serverURL,
		"website-id",
		"",
		"",
		logger,
	)

	umamiClient.Token = "test"

	application := &app.App{
		Config: &config.Config{},
		Logger: logger,
		Umami:  umamiClient,
	}

	return NewRealtimeHandler(application)
}

func TestRealtimeIndexReturnsAllURLs(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.Write([]byte(`
			{
				"urls": {
					"/": 10,
					"/event": 5,
					"/event/road-run": 2
				}
			}
			`))
		}),
	)

	defer server.Close()

	handler := setupHandler(server.URL)

	request := httptest.NewRequest(
		http.MethodGet,
		"/analytics/realtime",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Index(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	var body map[string]int

	err := json.NewDecoder(response.Body).
		Decode(&body)

	if err != nil {
		t.Fatal(err)
	}

	if body["/"] != 10 {
		t.Errorf(
			"expected / count 10, got %d",
			body["/"],
		)
	}

	if body["/event"] != 5 {
		t.Errorf(
			"expected /event count 5, got %d",
			body["/event"],
		)
	}
}

func TestRealtimeIndexFiltersByPrefix(
	t *testing.T,
) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.Write([]byte(`
			{
				"urls": {
					"/": 10,
					"/event": 5,
					"/event/road-run": 2,
					"/blog": 9
				}
			}
			`))
		}),
	)

	defer server.Close()

	handler := setupHandler(server.URL)

	request := httptest.NewRequest(
		http.MethodGet,
		"/analytics/realtime?path=/event",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Index(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	var body map[string]int

	err := json.NewDecoder(response.Body).
		Decode(&body)

	if err != nil {
		t.Fatal(err)
	}

	if len(body) != 2 {
		t.Errorf(
			"expected 2 results, got %d",
			len(body),
		)
	}

	if body["/event"] != 5 {
		t.Errorf(
			"expected /event count 5, got %d",
			body["/event"],
		)
	}

	if body["/event/road-run"] != 2 {
		t.Errorf(
			"expected /event/road-run count 2, got %d",
			body["/event/road-run"],
		)
	}

	if _, exists := body["/blog"]; exists {
		t.Errorf(
			"did not expect /blog in response",
		)
	}
}

func TestRealtimeIndexReturnsError(
	t *testing.T,
) {
	handler := setupHandler(
		"http://invalid-host",
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/analytics/realtime",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler.Index(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			response.StatusCode,
		)
	}
}
