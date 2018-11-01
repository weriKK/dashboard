package infrastructure

import (
	"encoding/json"
	"fmt"
	"github.com/weriKK/dashboard/application"
	"net/http"
	"strconv"
	"strings"
	"github.com/rs/cors"
)

type jsonFeedList struct {
	Count int
	Feeds []jsonFeedListItem
}

type jsonFeedListItem struct {
	Name   		string
	Url    		string
	Resource 	string
	Column 		int
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
		payload.Feeds = append(payload.Feeds, jsonFeedListItem{v.Name, v.Url, fmt.Sprintf("http://%s%s/%d", r.Host, r.URL, v.Id), v.Column})
	}
	writeJSONPayload(w, payload)
}

func webFeedContentHandler(w http.ResponseWriter, r *http.Request) {

	// Todo: this is a huge injection vulnerability issue here
	id, err := strconv.Atoi(r.URL.Path[strings.Index(r.URL.Path[1:], "/")+2:])
	if err != nil {
		panic(err)
	}
	feedContent, _ := application.GetFeedContent(id)

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
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	})

	s := &http.Server{
		Addr:    addr,
		Handler: c.Handler(mux),
	}

	return s
}
