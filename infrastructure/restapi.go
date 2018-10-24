package infrastructure

import (
	"encoding/json"
	"fmt"
	"github.com/weriKK/dashboard/application"
	"net/http"
)

type jsonFeedList struct {
	Count int
	Feeds []jsonFeedItem
}

type jsonFeedItem struct {
	Name string
	Url  string
}

func webFeedsHandler(w http.ResponseWriter, r *http.Request) {

	feedList, _ := application.GetFeedIdList()

	payload := jsonFeedList{len(feedList), []jsonFeedItem{}}
	for _, v := range feedList {
		payload.Feeds = append(payload.Feeds, jsonFeedItem{v.Name, fmt.Sprintf("http://%s%s/%d", r.Host, r.URL, v.Id)})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)

}

func NewRestHttpServer(addr string) *http.Server {

	mux := http.NewServeMux()
	mux.HandleFunc("/webfeeds", webFeedsHandler)

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
