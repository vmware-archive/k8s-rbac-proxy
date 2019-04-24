package proxy

import (
	"fmt"
)

type Logger interface {
	Error(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
}

type OutLogger struct {
	debug bool
}

var _ Logger = OutLogger{}

func NewOutLogger(debug bool) OutLogger {
	return OutLogger{debug}
}

func (l OutLogger) Error(str string, data ...interface{}) {
	fmt.Printf("error: "+str+"\n", data...)
}

func (l OutLogger) Info(str string, data ...interface{}) {
	fmt.Printf("info: "+str+"\n", data...)
}

func (l OutLogger) Debug(str string, data ...interface{}) {
	if l.debug {
		fmt.Printf("debug: "+str+"\n", data...)
	}
}

type PrefixLogger struct {
	prefix string
	logger Logger
}

var _ Logger = PrefixLogger{}

func NewPrefixLogger(prefix string, logger Logger) PrefixLogger {
	return PrefixLogger{prefix, logger}
}

func (l PrefixLogger) Error(str string, data ...interface{}) {
	l.logger.Error(l.prefix+": "+str, data...)
}

func (l PrefixLogger) Info(str string, data ...interface{}) {
	l.logger.Info(l.prefix+": "+str, data...)
}

func (l PrefixLogger) Debug(str string, data ...interface{}) {
	l.logger.Debug(l.prefix+": "+str, data...)
}
