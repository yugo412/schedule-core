package url

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type WebhookPayload struct {
	Slug       string `json:"slug"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
}

func SendWebhook(
	webhookURL string,
	payload WebhookPayload,
) error {
	client := resty.New().SetTimeout(5 * time.Second)

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(webhookURL)

	if err != nil {
		return err
	}

	if response.IsError() {
		return fmt.Errorf(
			"webhook returned status %d",
			response.StatusCode(),
		)
	}

	return nil
}
