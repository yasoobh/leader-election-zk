package main

import (
	"context"
	"fmt"
)

type key int

const loggerKey key = 0

type logger interface {
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type fmtLogger struct{}

func (fl fmtLogger) Infof(format string, args ...interface{}) {
	fmt.Println("[info] " + fmt.Sprintf(format, args...))
}

func (fl fmtLogger) Debugf(format string, args ...interface{}) {
	fmt.Println("[debug] " + fmt.Sprintf(format, args...))
}

func (fl fmtLogger) Errorf(format string, args ...interface{}) {
	fmt.Println("[error] " + fmt.Sprintf(format, args...))
}

// put logger in Context
func WithLogger(ctx context.Context, l logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func GetLogger(ctx context.Context) logger {
	return ctx.Value(loggerKey).(logger)
}

// use logger from Context
