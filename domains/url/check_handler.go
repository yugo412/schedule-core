package url

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type CheckResult struct {
	StatusCode int
	Status     string
	Error      string
}

func CheckURL(url string) CheckResult {
	client := resty.New().SetTimeout(5 * time.Second)

	response, err := client.R().
		Get(url)

	if err != nil {
		return CheckResult{
			Status: "warning",
			Error:  err.Error(),
		}
	}

	statusCode := response.StatusCode()

	switch {
	case statusCode == 404:
		return CheckResult{
			StatusCode: statusCode,
			Status:     "broken",
		}

	case statusCode == 410:
		return CheckResult{
			StatusCode: statusCode,
			Status:     "broken",
		}

	case statusCode >= 200 && statusCode < 400:
		return CheckResult{
			StatusCode: statusCode,
			Status:     "healthy",
		}

	default:
		return CheckResult{
			StatusCode: statusCode,
			Status:     "warning",
		}
	}
}
