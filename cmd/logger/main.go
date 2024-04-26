package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/isauran/logger"
)

func main() {
	logger.NewLogger(os.Stdout, logger.WithJSON(true))
	slog.Info("init", "logger", "log/slog", "format", "json")
	// {"time":"2024-03-27T15:05:29+05:00","level":"INFO","caller":"main.go:11","msg":"init","logger":"log/slog","format":"json"}
	
	gokitlogJSON := logger.NewGoKitLogger("info")
	gokitlogJSON.Log("msg", "init", "logger", "go-kit/log", "format", "json")
	// {"time":"2024-04-23T15:48:36+05:00","level":"INFO","caller":"main.go:19","msg":"init","logger":"go-kit/log","format":"json"}

	gormlogJSON := logger.NewGormLogger("info")
	gormlogJSON.Info(context.Background(), "init", "logger", "gorm.io/gorm/logger", "format", "json")
	// {"time":"2024-04-26T07:13:41+05:00","level":"INFO","caller":"main.go:21","msg":"init","logger":"gorm.io/gorm/logger","format":"json"}

	logger.NewLogger(os.Stdout)
	slog.Info("init", "logger", "log/slog", "format", "text")
	// time=2024-03-27T15:05:29+05:00 level=INFO caller=main.go:15 msg=init logger=log/slog format=text

	gokitlog := logger.NewGoKitLogger("info")
	gokitlog.Log("msg", "init", "logger", "go-kit/log", "format", "text")
	// time=2024-04-23T15:48:36+05:00 level=INFO caller=main.go:23 msg=init logger=go-kit/log format=text

	gormlog := logger.NewGormLogger("info")
	gormlog.Info(context.Background(), "init", "logger", "gorm.io/gorm/logger", "format", "text")
	// time=2024-04-26T07:13:41+05:00 level=INFO caller=main.go:33 msg=init logger=gorm.io/gorm/logger format=text
}
