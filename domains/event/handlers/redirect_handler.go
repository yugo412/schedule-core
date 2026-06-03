package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yugo412/schedule-core/app"
	"github.com/yugo412/schedule-core/domains/event/services"
	"github.com/yugo412/schedule-core/domains/url"
	"github.com/yugo412/schedule-core/integrations/umami"
)

type RedirectHandler struct {
	app             *app.App
	scheduleService *services.ScheduleService
}

func NewRedirectHandler(
	app *app.App,
	scheduleService *services.ScheduleService,
) *RedirectHandler {
	return &RedirectHandler{
		app:             app,
		scheduleService: scheduleService,
	}
}

func (h *RedirectHandler) Redirect(
	w http.ResponseWriter,
	r *http.Request,
) {
	slug := chi.URLParam(r, "slug")

	schedule, err := h.scheduleService.FindBySlug(
		r.Context(),
		slug,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			h.app.Logger.Warn("schedule not found", "slug", slug)

			http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

			return
		}

		h.app.Logger.Error(
			"failed to find schedule",
			"slug", slug,
			"error", err,
		)

		http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

		return
	}

	if h.app.Umami != nil {
		go func() {
			err := h.app.Umami.TrackEvent(
				umami.Event{
					Name:      "registration-click",
					URL:       r.URL.Path,
					Hostname:  r.Host,
					Language:  "id-ID",
					UserAgent: r.UserAgent(),
					Data: map[string]any{
						"slug":  schedule.Slug,
						"url":   schedule.Url,
						"title": schedule.Title,
					},
				},
			)

			if err != nil {
				h.app.Logger.Error(
					"failed to track umami event",
					"slug", slug,
					"error", err,
				)
			}
		}()
	}

	go func() {
		result := url.CheckURL(schedule.Url)

		if result.Status != "healthy" {
			h.app.Umami.TrackEvent(
				umami.Event{
					Name:      "link-check",
					URL:       r.URL.Path,
					Hostname:  r.Host,
					Language:  "id-ID",
					UserAgent: r.UserAgent(),
					Data: map[string]any{
						"slug":        schedule.Slug,
						"url":         schedule.Url,
						"title":       schedule.Title,
						"error":       result.Error,
						"status_code": result.StatusCode,
					},
				},
			)
		}
	}()

	http.Redirect(
		w,
		r,
		schedule.Url,
		http.StatusFound,
	)
}

func (h *RedirectHandler) RedirectSlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	schedule, err := h.scheduleService.FindBySlug(r.Context(), slug)
	if err != nil {
		if err == sql.ErrNoRows {
			h.app.Logger.Warn("schedule not found", "slug", slug)

			http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

			return
		}

		h.app.Logger.Warn("schedule not found", "slug", slug)

		http.Redirect(w, r, h.app.Config.MainUrl, http.StatusFound)

		return
	}

	if h.app.Umami != nil {
		err := h.app.Umami.TrackEvent(
			umami.Event{
				Name:      "redirect-slug",
				URL:       r.URL.Path,
				Hostname:  r.Host,
				Language:  "id-ID",
				UserAgent: r.UserAgent(),
				Data: map[string]any{
					"slug":  schedule.Slug,
					"url":   schedule.Url,
					"title": schedule.Title,
				},
			},
		)

		if err != nil {
			h.app.Logger.Error(
				"failed to track umami event",
				"slug", slug,
				"error", err,
			)
		}
	}

	http.Redirect(w, r, h.app.Config.MainUrl+"/event/"+slug, http.StatusFound)
}
