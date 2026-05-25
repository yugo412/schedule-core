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
	Username  string
	Password  string
	Token     string

	http   *resty.Client
	logger *slog.Logger
}

type LoginResponse struct {
	Token string `json:"token"`
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
	baseURL string,
	websiteID string,
	username string,
	password string,
	logger *slog.Logger,
) *Client {
	return &Client{
		BaseURL:   baseURL,
		WebsiteID: websiteID,
		Username:  username,
		Password:  password,

		http:   resty.New().SetTimeout(5 * time.Second),
		logger: logger,
	}
}

func (c *Client) ensureAuthenticated() error {
	if c.Token != "" {
		return nil
	}

	return c.Authenticate()
}

func (c *Client) Authenticate() error {
	var response LoginResponse

	_, err := c.http.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"username": c.Username,
			"password": c.Password,
		}).
		SetResult(&response).
		Post(c.BaseURL + "/api/auth/login")

	if err != nil {
		return err
	}

	if response.Token == "" {
		return errors.New("empty umami token")
	}

	c.Token = response.Token

	return nil
}

func (c *Client) Realtime() (*RealtimeResponse, error) {
	var response RealtimeResponse

	err := c.ensureAuthenticated()
	if err != nil {
		c.logger.Error("invalid authentication", "error", err)

		return nil, err
	}

	request, err := c.http.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(c.Token).
		SetResult(&response).
		Get(c.BaseURL + "/api/realtime/" + c.WebsiteID)

	if err != nil {
		c.logger.Warn("Failed to get realtime analytics", "error", err)

		return nil, err
	}

	if request.IsError() {
		c.logger.Warn("Failed to request realtime analytics", "error", request.Error())

		return nil, errors.New("failed to request umami realtime anaylytics")
	}

	return &response, nil
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
