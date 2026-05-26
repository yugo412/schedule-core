package router

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vinovest/sqlx"
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
) (
	http.Handler,
	*config.Config,
	*sqlx.DB,
	*httptest.Server,
) {
	cfg := &config.Config{
		MainUrl: "https://example.com",
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	db := setupTestDB(t)

	umamiServer := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch r.URL.Path {

			case "/api/auth/login":
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.Write([]byte(`
				{
					"token": "secret-token"
				}
				`))

			case "/api/realtime/local":
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.Write([]byte(`
				{
					"urls": {
						"/event": 10
					}
				}
				`))
			}
		}),
	)

	umamiClient := umami.NewClient(
		umamiServer.URL,
		"local",
		"username",
		"password",
		logger,
	)

	application := &app.App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Umami:  umamiClient,
	}

	return New(application), cfg, db, umamiServer
}

func TestRootRedirect(t *testing.T) {
	router, cfg, _, _ := setupRouter(t)

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
	router, _, _, _ := setupRouter(t)

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
	router, _, db, _ := setupRouter(t)

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
	router, cfg, _, _ := setupRouter(t)

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

func TestStatsRealtimeRoute(
	t *testing.T,
) {
	router, _, _, _ := setupRouter(t)

	request := httptest.NewRequest(
		http.MethodGet,
		"/stats/realtime",
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
}

func TestOfficialSlugRoute(
	t *testing.T,
) {
	router, _, db, server := setupRouter(t)

	defer server.Close()

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
		"https://register.com",
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

	router.ServeHTTP(
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

	if location != "https://register.com" {
		t.Errorf(
			"expected location %s, got %s",
			"https://register.com",
			location,
		)
	}
}

func TestRootSlugRoute(
	t *testing.T,
) {
	router, cfg, db, server := setupRouter(t)

	defer server.Close()

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

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

func TestRootRouteTracksEvent(
	t *testing.T,
) {
	tracked := false

	umamiServer := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			switch r.URL.Path {

			case "/api/send":
				tracked = true

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.Write([]byte(`
				{
					"sessionId": "session",
					"visitId": "visit"
				}
				`))
			}
		}),
	)

	defer umamiServer.Close()

	cfg := &config.Config{
		MainUrl: "https://example.com",
	}

	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	db := setupTestDB(t)

	umamiClient := umami.NewClient(
		umamiServer.URL,
		"website-id",
		"",
		"",
		logger,
	)

	application := &app.App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Umami:  umamiClient,
	}

	router := New(application)

	request := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
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

	if !tracked {
		t.Errorf(
			"expected umami track event to be called",
		)
	}
}
