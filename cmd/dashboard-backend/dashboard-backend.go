package main

import (
	"dashboard/internal/config"
	"dashboard/internal/feed"
	"dashboard/internal/wrappers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	if err := config.LoadYAML("config.yaml"); err != nil {
		panic(err)
	}

	f := feed.New()

	r := mux.NewRouter()
	r.HandleFunc("/webfeedlist", f.GetFeedListHandler).Methods(http.MethodGet)
	r.HandleFunc("/webfeeds/{id:[0-9]+}", f.GetFeedHandler).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler())

	m := wrappers.NewMetrics("dashboard", "incoming")
	r.Use(m.WithMetrics)
	r.Use(wrappers.WithCORS)
	r.Use(wrappers.WithRequestLogger)

	log.Fatal(http.ListenAndServe(":8080", r))
}
