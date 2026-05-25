package umami

type Event struct {
	Name      string
	URL       string
	Hostname  string
	Language  string
	UserAgent string
	Data      map[string]any
}
