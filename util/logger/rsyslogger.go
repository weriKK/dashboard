package logger

import (
	"fmt"
	"log"
	"log/syslog"
)

func RSyslogWriter(protocol string, host string, port int, flags syslog.Priority, tag string) *syslog.Writer {
	rsyslog, err := syslog.Dial(protocol, fmt.Sprintf("%s:%v", host, port), flags, tag)
	if err != nil {
		log.Fatal(err)
	}

	return rsyslog
}
