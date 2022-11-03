package util

import (
	"net/url"
	"testing"
)

func is_set(err error) bool {
	return err != nil
}

func TestGetLimitQueryParam(t *testing.T) {
	NO_LIMIT_PARAM := -1
	LIMIT_NOT_A_NUMBER := -2
	tests := []struct {
		p     ParsedUrlInfo
		limit int
	}{
		{ParsedUrlInfo{"", "", map[string][]string{}}, NO_LIMIT_PARAM},
		{ParsedUrlInfo{"", "", map[string][]string{"limit": []string{}}}, NO_LIMIT_PARAM},
		{ParsedUrlInfo{"", "", map[string][]string{"limit": []string{"duck", "cow"}}}, LIMIT_NOT_A_NUMBER},
		{ParsedUrlInfo{"", "", map[string][]string{"limit": []string{"10", "20", "30"}}}, 10},
	}

	for idx, test := range tests {
		u, err := test.p.GetLimitQueryParam()

		if test.limit == NO_LIMIT_PARAM || test.limit == LIMIT_NOT_A_NUMBER {
			if u != 0 || !is_set(err) {
				t.Errorf("#%v: Expected an error, but got limit: %v, err: %v", idx, u, err)
			}

		} else {

			if is_set(err) {
				t.Errorf("#%v: Unexpected error! Wanted: '%v', got '%v' but also error: %v", idx, test.limit, u, err)
			}

			if u != test.limit {
				t.Errorf("#%v: Limit was incorrect, got '%v', wanted '%v'", idx, u, test.limit)
			}

		}
	}
}

func buildURL(u string) *url.URL {
	s, err := url.ParseRequestURI(u)
	if err != nil {
		panic(err)
	}

	return s
}

func TestParseURL(t *testing.T) {

	tests := []struct {
		url   *url.URL
		path  string
		id    string
		limit string
	}{
		{buildURL("https://example.com:8080/webfeeds/3?limit=106&lulu=10"), "/webfeeds/3", "3", "106"},
		{buildURL("/webfeeds/3?limit=10&lulu=10"), "/webfeeds/3", "3", "10"},
		{buildURL("/webfeeds/5?lulu=69&limit=15"), "/webfeeds/5", "5", "15"},
		{buildURL("/webfeeds/1?limit=12"), "/webfeeds/1", "1", "12"},
		{buildURL("/webfeeds/2"), "/webfeeds/2", "2", ""},
	}

	for _, test := range tests {
		u, err := ParseUrl(test.url)
		if err != nil {
			t.Errorf("Parsing of '%v' failed: %v", test.url, err)
		}

		if u.Path != test.path {
			t.Errorf("URL Path was incorrect, got '%v', wanted '%v'", u.Path, test.path)
		}

		if u.LastPath != test.id {
			t.Errorf("URL Id was incorrect, got '%v', wanted '%v'", u.LastPath, test.id)
		}

		if test.limit != "" {
			if _, ok := u.Query["limit"]; !ok || u.Query["limit"][0] != test.limit {
				t.Errorf("URL limit query parameter was incorrect, got '%v', wanted '%v'", u.Query["limit"], test.limit)
			}
		}
	}
}
