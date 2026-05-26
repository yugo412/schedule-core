package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/domains/event/repositories"
	"github.com/yugo412/schedule-core/domains/event/services"
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

func setupHandler(
	t *testing.T,
) (*RedirectHandler, *config.Config, *sqlx.DB) {
	cfg := &config.Config{
		MainUrl: "https://example.com",
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	db := setupTestDB(t)

	application := &app.App{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}

	repository := repositories.NewScheduleRepository(
		db,
	)

	service := services.NewScheduleService(
		repository,
	)

	handler := NewRedirectHandler(
		application,
		service,
	)

	return handler, cfg, db
}

func TestRedirectFound(t *testing.T) {
	handler, _, db := setupHandler(t)

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

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(
		"slug",
		"mantra-run-2026",
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	recorder := httptest.NewRecorder()

	handler.Redirect(recorder, request)

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

func TestRedirectNotFound(t *testing.T) {
	handler, cfg, _ := setupHandler(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/official/not-found",
		nil,
	)

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(
		"slug",
		"not-found",
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	recorder := httptest.NewRecorder()

	handler.Redirect(recorder, request)

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

func TestRedirectSlugFound(
	t *testing.T,
) {
	handler, cfg, db := setupHandler(t)

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
		"mangkunegaran-run-2026",
		"Mangkunegaran Run 2026",
		"https://register.com",
	)

	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/mangkunegaran-run-2026",
		nil,
	)

	routeContext := chi.NewRouteContext()

	routeContext.URLParams.Add(
		"slug",
		"mangkunegaran-run-2026",
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	recorder := httptest.NewRecorder()

	handler.RedirectSlug(
		recorder,
		request,
	)

	response := recorder.Result()

	if response.StatusCode != http.StatusFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusFound,
			response.StatusCode,
		)
	}

	expected := cfg.MainUrl +
		"/event/mangkunegaran-run-2026"

	location := response.Header.Get(
		"Location",
	)

	if location != expected {
		t.Errorf(
			"expected location %s, got %s",
			expected,
			location,
		)
	}
}

func TestRedirectSlugNotFound(
	t *testing.T,
) {
	handler, cfg, _ := setupHandler(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/unknown-event",
		nil,
	)

	routeContext := chi.NewRouteContext()

	routeContext.URLParams.Add(
		"slug",
		"unknown-event",
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)

	recorder := httptest.NewRecorder()

	handler.RedirectSlug(
		recorder,
		request,
	)

	response := recorder.Result()

	if response.StatusCode != http.StatusFound {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusFound,
			response.StatusCode,
		)
	}

	location := response.Header.Get(
		"Location",
	)

	if location != cfg.MainUrl {
		t.Errorf(
			"expected location %s, got %s",
			cfg.MainUrl,
			location,
		)
	}
}
