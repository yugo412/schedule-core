package analytic

import (
	"github.com/go-chi/chi/v5"
	"github.com/yugo412/schedule-core/app"
)

func Routes(app *app.App) func(r chi.Router) {
	handler := NewRealtimeHandler(app)

	return func(r chi.Router) {
		r.Get("/realtime", handler.Index)
	}
}
