package main

import (
	"dashboard/internal/config"
	"dashboard/internal/feed"
	"dashboard/internal/wrappers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	if err := config.LoadYAML("config.yaml"); err != nil {
		panic(err)
	}

	f := feed.New()

	r := mux.NewRouter()
	r.HandleFunc("/webfeedlist", f.GetFeedListHandler).Methods(http.MethodGet)
	r.HandleFunc("/webfeeds/{id:[0-9]+}", f.GetFeedHandler).Methods(http.MethodGet)

	c := wrappers.WithCORS(r)
	l := wrappers.WithRequestLogger(c)

	log.Fatal(http.ListenAndServe(":8080", l))
}
