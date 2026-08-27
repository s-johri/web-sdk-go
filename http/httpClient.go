package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// requestTimeout bounds a single PayU call. Without it a stalled endpoint
// blocks the calling goroutine forever.
const requestTimeout = 30 * time.Second

// maxErrorBodyBytes limits how much of an error response is quoted back, so a
// large HTML error page does not end up inside an error string.
const maxErrorBodyBytes = 512

// defaultClient is shared across calls so connections are reused.
var defaultClient = &http.Client{Timeout: requestTimeout}

// PostForm posts an x-www-form-urlencoded payload to a PayU endpoint and
// decodes the JSON response.
func PostForm(endpoint string, payload url.Values) (map[string]interface{}, error) {
	request, err := http.NewRequest("POST", endpoint, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	// Set the content type header to x-www-form-urlencoded
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := defaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		// PayU serves an HTML error page on failure, which would otherwise
		// surface as a confusing JSON decode error.
		body, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
		return nil, fmt.Errorf("payu returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var responseBody map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		return nil, err
	}
	return responseBody, nil
}
