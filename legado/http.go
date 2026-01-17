package legado

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MaxResponseSize is the maximum response body size (10MB)
const MaxResponseSize = 10 * 1024 * 1024

// InsecureSkipVerify controls whether TLS certificate verification is skipped.
// Should only be set to true for testing or specific sites that require it.
var InsecureSkipVerify = false

// postBytes performs an HTTP POST request and returns the response body
func postBytes(url string, headers http.Header, body []byte, timeout time.Duration, retryCount int) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: InsecureSkipVerify},
		},
	}

	var lastErr error
	for i := 0; i <= retryCount; i++ {
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// Read body and close immediately (not defer) to avoid resource leak in retry loop
		// Limit response size to prevent DoS
		limitedReader := io.LimitReader(resp.Body, MaxResponseSize+1)
		data, err := io.ReadAll(limitedReader)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if len(data) > MaxResponseSize {
			return nil, fmt.Errorf("response too large (max %d bytes)", MaxResponseSize)
		}

		return data, nil
	}

	return nil, lastErr
}
