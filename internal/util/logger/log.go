package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{log.New(out, prefix, flag)}
}

var g = New(os.Stdout, "", 0)

func SetOutput(out io.Writer) {
	g.SetOutput(out)
}

func Info(v ...interface{}) {
	g.Printf("INFO %v", v...)
}

func Infof(format string, v ...interface{}) {
	g.Printf(fmt.Sprintf("INFO %s", format), v...)
}

func Debug(v ...interface{}) {
	g.Printf("DEBUG %v", v...)
}

func Debugf(format string, v ...interface{}) {
	g.Printf(fmt.Sprintf("DEBUG %s", format), v...)
}

func Error(v ...interface{}) {
	g.Printf("INFO %v", v...)
}

func Errorf(format string, v ...interface{}) {
	g.Printf(fmt.Sprintf("ERROR %s", format), v...)
}

func Fatal(v ...interface{}) {
	g.Fatalf("FATAL %v", v...)
}

func Fatalf(format string, v ...interface{}) {
	g.Fatalf(fmt.Sprintf("FATAL %s", format), v...)
}

func Panic(v ...interface{}) {
	g.Panicf("PANIC %v", v...)
}

func Panicf(format string, v ...interface{}) {
	g.Panicf(fmt.Sprintf("PANIC %s", format), v...)
}
