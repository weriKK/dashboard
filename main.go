package main

import (
	"github.com/weriKK/dashboard/application"
	"github.com/weriKK/dashboard/infrastructure"
	"github.com/weriKK/dashboard/util/logger"
)

func main() {

	application.Init(infrastructure.NewInMemoryFeedRepository())

	restApiServer := infrastructure.NewRestHttpServer(":8080")
	logger.Fatal(restApiServer.ListenAndServe())
}
