package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/vinovest/sqlx"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/domains/event/models"
	"github.com/yugo412/schedule-core/domains/event/repositories"
	"github.com/yugo412/schedule-core/domains/event/services"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	sqlDB, mock, err := sqlmock.New()

	if err != nil {
		t.Fatal(err)
	}

	db := sqlx.NewDb(sqlDB, "pgx")

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectation: %v", err)
		}
		db.Close()
	})

	switch t.Name() {
	case "TestRedirectFound":
		expectSchedule(t, mock, "mantra-run-2026", "https://example.com/register")
	case "TestRedirectSlugFound":
		expectSchedule(t, mock, "mangkunegaran-run-2026", "https://register.com")
	}

	return db
}

func expectSchedule(t *testing.T, mock sqlmock.Sqlmock, slug, url string) {
	t.Helper()
	mock.ExpectExec("INSERT INTO schedules").
		WithArgs(slug, sqlmock.AnyArg(), url).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT url, title, slug FROM schedules WHERE slug = \\$1 LIMIT 1").
		WithArgs(slug).
		WillReturnRows(sqlmock.NewRows([]string{"url", "title", "slug"}).AddRow(url, "Event", slug))
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
			$1,
			$2,
			$3
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

func TestCheckURLWithoutUmami(t *testing.T) {
	handler, _, _ := setupHandler(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	request := httptest.NewRequest(http.MethodGet, "/official/event", nil)

	handler.checkURL(
		&models.Schedule{Slug: "event", Url: server.URL},
		request,
	)
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
			$1,
			$2,
			$3
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
