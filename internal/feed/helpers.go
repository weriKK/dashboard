package feed

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

func (f *Feed) getFeedFromURL(link string) ([]byte, error) {

	c := &http.Client{}

	// HACK: https://www.reddit.com/r/redditdev/comments/t8e8hc/getting_nothing_but_429_responses_when_using_go/
	if strings.Contains(link, "reddit.com") {
		c.Transport = &http.Transport{
			TLSNextProto: map[string]func(authority string, c *tls.Conn) http.RoundTripper{},
		}
	}

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "MyDashboard")

	resp, err := f.instrumentedDo(c.Do)(req)
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

func (f *Feed) instrumentedDo(next func(*http.Request) (*http.Response, error)) func(*http.Request) (*http.Response, error) {
	return func(r *http.Request) (*http.Response, error) {
		start := time.Now()

		resp, err := next(r)

		labelValues := []string{strconv.Itoa(resp.StatusCode), r.Method, r.Host, r.URL.Path}

		f.feedOutgoingMetrics.ReqTotal.WithLabelValues(labelValues...).Inc()
		f.feedOutgoingMetrics.ReqDurationMs.WithLabelValues(labelValues...).Observe(float64(time.Since(start).Milliseconds()))
		f.feedOutgoingMetrics.RespSizeBytes.WithLabelValues(labelValues...).Observe(float64(r.ContentLength))

		return resp, err
	}
}
