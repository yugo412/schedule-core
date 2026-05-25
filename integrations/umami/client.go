package umami

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL   string
	WebsiteID string
	http      *resty.Client
	logger    *slog.Logger
}

type TrackEventResponse struct {
	Cache     string `json:"cache"`
	SessionID string `json:"sessionId"`
	VisitID   string `json:"visitId"`
}

type ErrorResponse struct {
	Beep string `json:"beep"`
}

func NewClient(
	baseURL,
	websiteID string,
	logger *slog.Logger,
) *Client {
	return &Client{
		BaseURL:   baseURL,
		WebsiteID: websiteID,
		http: resty.New().
			SetTimeout(5 * time.Second),
		logger: logger,
	}
}

func (c *Client) TrackEvent(
	event Event,
) error {
	payload := EventPayload{
		Type: "event",
		Payload: Payload{
			Hostname:  event.Hostname,
			Language:  event.Language,
			URL:       event.URL,
			WebsiteID: c.WebsiteID,
			Name:      event.Name,
			Data:      event.Data,
		},
	}

	var successResponse TrackEventResponse

	response, err := c.http.R().
		SetHeader(
			"Content-Type",
			"application/json",
		).
		SetHeader(
			"User-Agent",
			event.UserAgent,
		).
		SetBody(payload).
		SetResult(&successResponse).
		Post(c.BaseURL + "/api/send")

	if err != nil {
		c.logger.Error(
			"failed to send umami event",
			"event", event.Name,
			"error", err,
		)

		return err
	}

	if response.IsError() {
		c.logger.Error(
			"umami returned error response",
			"event", event.Name,
			"status", response.StatusCode(),
		)

		return errors.New("umami request failed")
	}

	var errorResponse ErrorResponse

	err = json.Unmarshal(
		response.Body(),
		&errorResponse,
	)

	if err == nil && errorResponse.Beep == "boop" {
		c.logger.Warn(
			"umami rejected event as bot",
			"event", event.Name,
			"url", event.URL,
		)

		return errors.New("umami rejected event")
	}

	c.logger.Info(
		"umami event tracked",
		"event", event.Name,
		"status", response.StatusCode(),
		"session_id", successResponse.SessionID,
		"visit_id", successResponse.VisitID,
	)

	return nil
}
