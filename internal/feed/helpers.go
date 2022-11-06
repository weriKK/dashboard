package feed

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

func getFeedFromURL(link string) ([]byte, error) {
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "dashboard-backend/1.0")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respDump, _ := httputil.DumpResponse(resp, true)
		return nil, fmt.Errorf("received error from feed server:\n%s", string(respDump))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read fetch body: %w", err)
	}

	return body, nil
}
