package analytic

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yugo412/schedule-core/app"
)

type RealtimeHandler struct {
	app *app.App
}

func NewRealtimeHandler(app *app.App) *RealtimeHandler {
	return &RealtimeHandler{app: app}
}

func (h *RealtimeHandler) Index(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	realtime, err := h.app.Umami.Realtime()
	if err != nil {
		h.app.Logger.Error("failed to fetch realtime analytic", "error", err)

		http.Error(w, "Failed to fetch realtime analytics", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	filtered := make(map[string]int)

	for realtimePath, count := range realtime.URLs {
		if path == "" {
			filtered[realtimePath] = count

			continue
		}

		if strings.HasPrefix(
			realtimePath,
			path,
		) {
			filtered[realtimePath] = count
		}
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	err = json.NewEncoder(w).Encode(filtered)
	if err != nil {
		h.app.Logger.Error("failed to encode realtime analytic", "error", err)

		http.Error(w, "Failed to encode realtime analytics", http.StatusInternalServerError)

		return
	}

}
