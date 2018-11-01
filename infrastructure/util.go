package infrastructure

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type parsedUrlInfo struct {
	Path     string
	LastPath string
	Query    map[string][]string
}

func (p parsedUrlInfo) getLimitQueryParam() (int, error) {

	_, ok := p.Query["limit"]
	if !ok || len(p.Query["limit"]) < 1 {
		return 0, errors.New("no 'limit' query param")
	}

	limit, err := strconv.Atoi(p.Query["limit"][0])
	if err != nil {
		return 0, err
	}

	return limit, err
}

func parseUrl(url *url.URL) (*parsedUrlInfo, error) {
	lastPath := url.Path[strings.Index(url.Path[1:], "/")+2:]
	q := url.Query()

	parsedInfo := parsedUrlInfo{url.Path, lastPath, q}
	return &parsedInfo, nil
}
