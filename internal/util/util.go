package util

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/mmcdole/gofeed"
)

type ParsedUrlInfo struct {
	Path     string
	LastPath string
	Query    map[string][]string
}

func (p ParsedUrlInfo) GetLimitQueryParam() (int, error) {

	_, ok := p.Query["limit"]
	if !ok || len(p.Query["limit"]) < 1 {
		return 0, errors.New("no 'limit' query param")
	}

	limit, err := strconv.Atoi(p.Query["limit"][0])
	if err != nil {
		return 0, err
	}

	return limit, nil
}

func ParseUrl(url *url.URL) (*ParsedUrlInfo, error) {
	lastPath := url.Path[strings.Index(url.Path[1:], "/")+2:]
	q := url.Query()

	parsedInfo := ParsedUrlInfo{url.Path, lastPath, q}
	return &parsedInfo, nil
}

func NewFeedParser() *gofeed.Parser {
	// TODO: this is a workaround to avoid reddit.com blocking the requests with 403 Forbidden
	//       it has something to do with TLS or http2, and it was not an issue with earlier go versions (v1.11)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	p := gofeed.NewParser()
	p.Client = client

	return p
}
