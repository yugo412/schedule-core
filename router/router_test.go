package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/integrations/umami"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect(
		"sqlite",
		":memory:",
	)

	if err != nil {
		t.Fatal(err)
	}

	schema := `
	CREATE TABLE schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL,
		title TEXT NOT NULL,
		url TEXT NOT NULL
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func setupRouter(
	t *testing.T,
) (http.Handler, *config.Config, *sqlx.DB) {
	cfg := &config.Config{
		MainUrl: "https://example.com",
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	db := setupTestDB(t)

	umamiClient := umami.NewClient("https://localhost", "local", logger)

	application := &app.App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Umami:  umamiClient,
	}

	return New(application), cfg, db
}

func TestRootRedirect(t *testing.T) {
	router, cfg, _ := setupRouter(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusFound,
			response.StatusCode,
		)
	}

	location := response.Header.Get("Location")

	if location != cfg.MainUrl {
		t.Errorf(
			"expected location %s, got %s",
			cfg.MainUrl,
			location,
		)
	}
}

func TestUpRoute(t *testing.T) {
	router, _, _ := setupRouter(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/up",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusOK,
			response.StatusCode,
		)
	}

	body := recorder.Body.String()

	if body != "OK" {
		t.Errorf(
			"expected body %q, got %q",
			"OK",
			body,
		)
	}
}

func TestOfficialRedirect(t *testing.T) {
	router, _, db := setupRouter(t)

	_, err := db.Exec(`
		INSERT INTO schedules (
			slug,
			title,
			url
		) VALUES (
			?,
			?,
			?
		)
	`,
		"mantra-run-2026",
		"Mantra Run 2026",
		"https://example.com/register",
	)

	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/official/mantra-run-2026",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusFound,
			response.StatusCode,
		)
	}

	location := response.Header.Get("Location")

	expected := "https://example.com/register"

	if location != expected {
		t.Errorf(
			"expected location %s, got %s",
			expected,
			location,
		)
	}
}

func TestOfficialRedirectNotFound(t *testing.T) {
	router, cfg, _ := setupRouter(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/official/not-found",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	if response.StatusCode != http.StatusFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusFound,
			response.StatusCode,
		)
	}

	location := response.Header.Get("Location")

	if location != cfg.MainUrl {
		t.Errorf(
			"expected location %s, got %s",
			cfg.MainUrl,
			location,
		)
	}
}
