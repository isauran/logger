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
	// {"time":"2024-04-26T08:09:44+05:00","level":"INFO","caller":"main.go:13","msg":"init","logger":"log/slog","format":"json"}

	{
		logger := logger.NewGoKitLogger("info")
		logger.Log("msg", "init", "logger", "go-kit/log", "format", "json")
		// {"time":"2024-04-26T08:09:44+05:00","level":"INFO","caller":"main.go:18","msg":"init","logger":"go-kit/log","format":"json"}
	}

	{
		logger := logger.NewGormLogger("info")
		logger.Info(context.Background(), "init %s %s %s %s", "logger", "gorm.io/gorm/logger", "format", "json")
		// {"time":"2024-04-26T08:09:44+05:00","level":"INFO","caller":"main.go:24","msg":"init logger gorm.io/gorm/logger format json"}
	}

	logger.NewLogger(os.Stdout)
	slog.Info("init", "logger", "log/slog", "format", "text")
	// time=2024-04-26T08:09:44+05:00 level=INFO caller=main.go:29 msg=init logger=log/slog format=text

	{
		logger := logger.NewGoKitLogger("info")
		logger.Log("msg", "init", "logger", "go-kit/log", "format", "text")
		// time=2024-04-26T08:09:44+05:00 level=INFO caller=main.go:34 msg=init logger=go-kit/log format=text
	}

	{
		logger := logger.NewGormLogger("info")
		logger.Info(context.Background(), "init %s %s %s %s", "logger", "gorm.io/gorm/logger", "format", "text")
		// time=2024-04-26T08:09:44+05:00 level=INFO caller=main.go:40 msg="init logger gorm.io/gorm/logger format text"
	}
}
