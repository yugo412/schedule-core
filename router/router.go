package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/domains/redirect"
)

func New(app *app.App) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, app.Config.MainUrl, http.StatusFound)
	})

	r.Get("/up", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/official", redirect.Routes(app))

	return r
}
