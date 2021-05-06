package channelx

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debugf(str string, args ...interface{})
	Infof(str string, args ...interface{})
	Warnf(str string, args ...interface{})
	Errorf(str string, args ...interface{})
}

type consoleLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewConsoleLogger() *consoleLogger {
	return &consoleLogger{
		debug: log.New(os.Stdout, "[debug] ", log.LstdFlags),
		info:  log.New(os.Stdout, "[info] ", log.LstdFlags),
		warn:  log.New(os.Stdout, "[warn] ", log.LstdFlags),
		error: log.New(os.Stdout, "[error] ", log.LstdFlags),
	}
}

func (cl *consoleLogger) Debugf(format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.debug.Println(message)
}

func (cl *consoleLogger) Infof(format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.info.Println(message)
}

func (cl *consoleLogger) Warnf(format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.warn.Println(message)
}

func (cl *consoleLogger) Errorf(format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.error.Println(message)
}
