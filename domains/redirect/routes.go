package redirect

import (
	"github.com/go-chi/chi/v5"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/domains/event/handlers"
	"github.com/yugo412/schedule-core/domains/event/repositories"
	"github.com/yugo412/schedule-core/domains/event/services"
)

func Routes(app *app.App) func(r chi.Router) {
	repository := repositories.NewScheduleRepository(app.DB)
	scheduleService := services.NewScheduleService(repository)
	handler := handlers.NewRedirectHandler(app, scheduleService)

	return func(r chi.Router) {
		r.Get("/{slug}", handler.Redirect)
	}
}
