package main

import (
	"log"

	"github.com/weriKK/dashboard/application"
	"github.com/weriKK/dashboard/infrastructure"
)

func main() {
	application.Init(infrastructure.NewInMemoryFeedRepository())

	restApiServer := infrastructure.NewRestHttpServer(":8080")
	log.Fatal(restApiServer.ListenAndServe())
}