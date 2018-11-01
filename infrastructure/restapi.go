package infrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/cors"
	"github.com/weriKK/dashboard/application"
)

type jsonFeedList struct {
	Count int
	Feeds []jsonFeedListItem
}

type jsonFeedListItem struct {
	Name      string
	Url       string
	Resource  string
	Column    int
	ItemLimit int
}

type jsonFeed struct {
	Id    int
	Name  string
	Url   string
	Items []jsonFeedItem
}

type jsonFeedItem struct {
	Title       string
	Url         string
	Description string
}

func webFeedListHandler(w http.ResponseWriter, r *http.Request) {

	feedList, _ := application.GetFeedIdList()

	payload := jsonFeedList{len(feedList), []jsonFeedListItem{}}
	for _, v := range feedList {
		payload.Feeds = append(payload.Feeds, jsonFeedListItem{v.Name, v.Url, fmt.Sprintf("http://%s%s/%d", r.Host, r.URL, v.Id), v.Column, v.ItemLimit})
	}
	writeJSONPayload(w, payload)
}

func webFeedContentHandler(w http.ResponseWriter, r *http.Request) {

	parsed, err := parseUrl(r.URL)
	if err != nil {
		panic(err)
	}

	id, err := strconv.Atoi(parsed.LastPath)
	if err != nil {
		panic(err)
	}

	// if error, limit is set to 0
	limit, err := parsed.getLimitQueryParam()

	feedContent, _ := application.GetFeedContent(id, limit)

	payload := jsonFeed{feedContent.Id, feedContent.Name, feedContent.Url, []jsonFeedItem{}}
	for _, v := range feedContent.Items {
		payload.Items = append(payload.Items, jsonFeedItem{v.Title, v.Url, v.Description})
	}
	writeJSONPayload(w, payload)
}

func writeJSONPayload(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func NewRestHttpServer(addr string) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/webfeeds", webFeedListHandler)
	mux.HandleFunc("/webfeeds/", webFeedContentHandler) // Trailing '/' is a different subtree, /webfeeds/{id} will comehere

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	})

	s := &http.Server{
		Addr:    addr,
		Handler: c.Handler(mux),
	}

	return s
}
