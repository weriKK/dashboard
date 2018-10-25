package infrastructure

import (
	"encoding/json"
	"fmt"
	"github.com/weriKK/dashboard/application"
	"net/http"
	"strconv"
	"strings"
)

type jsonFeedList struct {
	Count int
	Feeds []jsonFeedListItem
}

type jsonFeedListItem struct {
	Name string
	Url  string
}

type jsonFeed struct {
	Id    int
	Name  string
	Url   string
	Items []jsonFeedItem
}

type jsonFeedItem struct {
	Title   string
	Url     string
	Content string
}

func webFeedListHandler(w http.ResponseWriter, r *http.Request) {

	feedList, _ := application.GetFeedIdList()

	payload := jsonFeedList{len(feedList), []jsonFeedListItem{}}
	for _, v := range feedList {
		payload.Feeds = append(payload.Feeds, jsonFeedListItem{v.Name, fmt.Sprintf("http://%s%s/%d", r.Host, r.URL, v.Id)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)

}

func webFeedContentHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.URL.Path[strings.Index(r.URL.Path[1:], "/")+2:])
	if err != nil {
		panic(err)
	}
	feedContent, _ := application.GetFeedContent(id)

	payload := jsonFeed{feedContent.Id, feedContent.Name, feedContent.Url, []jsonFeedItem{}}
	for _, v := range feedContent.Items {
		payload.Items = append(payload.Items, jsonFeedItem{v.Title, v.Url, v.Content})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func NewRestHttpServer(addr string) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/webfeeds", webFeedListHandler)
	mux.HandleFunc("/webfeeds/", webFeedContentHandler) // Trailing '/' is a different subtree, /webfeeds/{id} will comehere

	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

/*
[
	{
		"name": "MMO Champion"
		"url": "http://localhost:8080/webfeeds/mmochampion"
	},
	{
		"name": "GiantBomb"
		"url": "http://localhost:8080/webfeeds/giantbomb"
	}
]
*/
