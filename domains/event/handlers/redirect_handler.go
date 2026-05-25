package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/domains/event/services"
)

type RedirectHandler struct {
	app             *app.App
	scheduleService *services.ScheduleService
}

func NewRedirectHandler(app *app.App, scheduleService *services.ScheduleService) *RedirectHandler {
	return &RedirectHandler{
		app:             app,
		scheduleService: scheduleService,
	}
}

func (h *RedirectHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	schedule, err := h.scheduleService.FindBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			h.app.Logger.Warn("Schedule not found", "slug", slug)
			http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

			return
		}

		h.app.Logger.Error("Error finding schedule", "slug", slug, "error", err)

		http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

		return
	}

	http.Redirect(w, r, schedule.Url, http.StatusFound)
}
