package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// App logs startup events, auth failures, and database errors → logs/app.log
// Docker logs all Docker SDK and SSH operations               → logs/docker.log
var (
	App    *slog.Logger
	Docker *slog.Logger
)

var requestOutput io.Writer

// RequestOutput returns the writer to pass to Fiber's logger middleware Output field.
func RequestOutput() io.Writer { return requestOutput }

type ctxKey struct{}

// ContextWithRequestID stores the request ID in ctx for use in the service layer.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// RequestIDFromContext retrieves the request ID stored by ContextWithRequestID.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKey{}).(string)
	return v
}

// Init initializes all loggers. Must be called once from main before any logger is used.
// Each logger writes to stdout AND its dedicated file (rotation via lumberjack).
func Init(level slog.Level) {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		panic("logger: cannot create logs/ directory: " + err.Error())
	}

	opts := &slog.HandlerOptions{Level: level}

	App = slog.New(slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, newFile("logs/app.log")), opts,
	))

	Docker = slog.New(slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, newFile("logs/docker.log")), opts,
	))

	requestOutput = io.MultiWriter(os.Stdout, newFile("logs/requests.log"))

	slog.SetDefault(App)
}

func newFile(path string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename: path,
		MaxSize:  100,
		MaxAge:   30,
		Compress: true,
	}
}
