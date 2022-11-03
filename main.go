package main

import (
	"dashboard/internal/application"
	"dashboard/internal/infrastructure"
	"dashboard/internal/util/logger"
)

func main() {

	application.Init(infrastructure.NewInMemoryFeedRepository())

	restApiServer := infrastructure.NewRestHttpServer(":8080")
	logger.Fatal(restApiServer.ListenAndServe())
}
