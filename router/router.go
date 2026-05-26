package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/domains/analytic"
	"github.com/yugo412/schedule-core/domains/redirect"
	"github.com/yugo412/schedule-core/integrations/umami"
)

func New(app *app.App) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if app.Umami != nil {
			err := app.Umami.TrackEvent(
				umami.Event{
					Name:      "redirect-index",
					URL:       "/",
					Hostname:  r.Host,
					Language:  "id-ID",
					UserAgent: r.UserAgent(),
				},
			)
			if err != nil {
				app.Logger.Error("failed to track event", "error", err)
			}
		}

		http.Redirect(w, r, app.Config.MainUrl, http.StatusFound)
	})

	r.Get("/up", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Route("/stats", analytic.Routes(app))

	r.Route("/", redirect.Routes(app))

	return r
}
