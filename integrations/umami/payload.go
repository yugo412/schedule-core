package umami

type EventPayload struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	WebsiteID string         `json:"website"`
	Hostname  string         `json:"hostname"`
	Language  string         `json:"language"`
	Url       string         `json:"url"`
	Name      string         `json:"name"`
	Data      map[string]any `json:"data,omitempty"`
}
