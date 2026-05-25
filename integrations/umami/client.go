package umami

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL   string
	WebsiteID string
	http      *resty.Client
	logger    *slog.Logger
}

func NewClient(baseURL, websiteID string, logger *slog.Logger) *Client {
	return &Client{
		BaseURL:   baseURL,
		WebsiteID: websiteID,
		http:      resty.New(),
		logger:    logger,
	}
}

func (c *Client) TrackEvent(event Event) error {
	payload := EventPayload{
		Type: "event",
		Payload: Payload{
			Hostname:  event.Hostname,
			Language:  event.Language,
			Url:       event.URL,
			WebsiteID: c.WebsiteID,
			Name:      event.Name,
			Data:      event.Data,
		},
	}

	_, err := c.http.R().SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", event.UserAgent).
		SetBody(payload).
		Post(c.BaseURL + "/api/send")

	if err != nil {
		return err
	}

	return nil

}
