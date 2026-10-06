package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/config"
	"github.com/yugo412/schedule-core/domains/event/models"
	"github.com/yugo412/schedule-core/domains/event/repositories"
	"github.com/yugo412/schedule-core/domains/event/services"
	"github.com/yugo412/schedule-core/domains/url"
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
		url TEXT,
		started_at TEXT
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
		MainUrl:   "https://example.com",
		UTMSource: "jadwallari.com",
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
		"https://example.com/register?ref=campaign",
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

	expected := "https://example.com/register?ref=campaign&utm_source=jadwallari.com"

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

func TestRedirectWithoutUrl(t *testing.T) {
	handler, cfg, db := setupHandler(t)

	_, err := db.Exec(`
		INSERT INTO schedules (
			slug,
			title,
			url
		) VALUES (
			?,
			?,
			NULL
		)
	`,
		"null-url-run-2026",
		"Null Url Run 2026",
	)

	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodGet,
		"/official/null-url-run-2026",
		nil,
	)

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(
		"slug",
		"null-url-run-2026",
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

func TestCheckURLWithoutWebhook(t *testing.T) {
	handler, _, _ := setupHandler(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	handler.checkURL(
		&models.Schedule{Slug: "event", Url: server.URL},
	)
}

func TestCheckURLCallsWebhookOnBrokenLink(t *testing.T) {
	handler, cfg, _ := setupHandler(t)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer target.Close()

	var (
		mutex    sync.Mutex
		called   bool
		received url.WebhookPayload
	)

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		called = true

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode webhook payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer webhook.Close()

	cfg.LinkCheckWebhookURL = webhook.URL

	handler.checkURL(
		&models.Schedule{
			Slug:  "event",
			Title: "Event",
			Url:   target.URL,
		},
	)

	mutex.Lock()
	defer mutex.Unlock()

	if !called {
		t.Fatal("expected webhook to be called")
	}

	if received.Slug != "event" {
		t.Errorf("expected slug event, got %s", received.Slug)
	}

	if received.Status != "warning" {
		t.Errorf("expected status warning, got %s", received.Status)
	}

	if received.StatusCode != http.StatusInternalServerError {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			received.StatusCode,
		)
	}
}

func TestCheckURLSkipsWebhookWhenHealthy(t *testing.T) {
	handler, cfg, _ := setupHandler(t)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer target.Close()

	var (
		mutex  sync.Mutex
		called bool
	)

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		called = true

		w.WriteHeader(http.StatusOK)
	}))

	defer webhook.Close()

	cfg.LinkCheckWebhookURL = webhook.URL

	handler.checkURL(
		&models.Schedule{Slug: "event", Url: target.URL},
	)

	mutex.Lock()
	defer mutex.Unlock()

	if called {
		t.Error("expected webhook not to be called for healthy link")
	}
}

func redirectRequest(slug string) *http.Request {
	request := httptest.NewRequest(
		http.MethodGet,
		"/official/"+slug,
		nil,
	)

	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("slug", slug)

	return request.WithContext(
		context.WithValue(
			request.Context(),
			chi.RouteCtxKey,
			routeContext,
		),
	)
}

func TestRedirectSkipsLinkCheckWhenScheduleStarted(t *testing.T) {
	handler, cfg, db := setupHandler(t)

	called := make(chan struct{}, 1)

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case called <- struct{}{}:
		default:
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer webhook.Close()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer target.Close()

	cfg.LinkCheckWebhookURL = webhook.URL

	_, err := db.Exec(
		"INSERT INTO schedules (slug, title, url, started_at) VALUES (?, ?, ?, ?)",
		"past-run-2026",
		"Past Run 2026",
		target.URL,
		time.Now().UTC().Add(-time.Hour).Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		t.Fatal(err)
	}

	handler.Redirect(
		httptest.NewRecorder(),
		redirectRequest("past-run-2026"),
	)

	select {
	case <-called:
		t.Fatal("expected the link check to be skipped for a schedule that already started")
	case <-time.After(500 * time.Millisecond):
	}
}

func TestRedirectChecksLinkWhenScheduleHasNotStarted(t *testing.T) {
	handler, cfg, db := setupHandler(t)

	called := make(chan struct{}, 1)

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case called <- struct{}{}:
		default:
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer webhook.Close()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	defer target.Close()

	cfg.LinkCheckWebhookURL = webhook.URL

	_, err := db.Exec(
		"INSERT INTO schedules (slug, title, url, started_at) VALUES (?, ?, ?, ?)",
		"upcoming-run-2026",
		"Upcoming Run 2026",
		target.URL,
		time.Now().UTC().Add(time.Hour).Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		t.Fatal(err)
	}

	handler.Redirect(
		httptest.NewRecorder(),
		redirectRequest("upcoming-run-2026"),
	)

	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Fatal("expected the link check to run for an upcoming schedule")
	}
}
