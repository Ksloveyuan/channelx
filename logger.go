package channelx

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Debugc(category, format string, arg ...interface{})
	Infoc(category, format string, arg ...interface{})
	Warnc(category string, err error,format string, arg ...interface{})
	Errorc(category string, err error, format string, arg ...interface{})
}

type consoleLogger struct {
	debug *log.Logger
	info *log.Logger
	warn *log.Logger
	error *log.Logger
}

func newConsoleLogger() *consoleLogger  {
	return &consoleLogger{
		debug:log.New(os.Stdout, "[debug] ", log.LstdFlags),
		info:log.New(os.Stdout, "[info] ", log.LstdFlags),
		warn:log.New(os.Stdout, "[warn] ", log.LstdFlags),
		error:log.New(os.Stdout, "[error] ", log.LstdFlags),
	}
}

func(cl *consoleLogger) Debugc(category, format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.debug.Println( category, message)
}

func(cl *consoleLogger) Infoc(category, format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.info.Println( category, message)
}

func(cl *consoleLogger) Warnc(category string, err error, format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	if err == nil{
		cl.warn.Println(category, message)
	} else {
		cl.warn.Println(category, err.Error(), message)
	}
}

func(cl *consoleLogger) Errorc(category string, err error, format string, arg ...interface{}) {
	message := fmt.Sprintf(format, arg...)
	cl.error.Println(category, err.Error(), message)
}
