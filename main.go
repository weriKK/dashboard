package main

import (
	"log/syslog"

	"github.com/weriKK/dashboard/application"
	"github.com/weriKK/dashboard/infrastructure"
	"github.com/weriKK/dashboard/util/logger"
)

func main() {

	logger.SetOutput(logger.RSyslogWriter("tcp",
		"127.0.0.1",
		514,
		syslog.LOG_WARNING|syslog.LOG_DAEMON,
		"dashboard"))

	application.Init(infrastructure.NewInMemoryFeedRepository())

	restApiServer := infrastructure.NewRestHttpServer(":8080")
	logger.Fatal(restApiServer.ListenAndServe())
}
