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
	// {"time":"2024-04-26T21:11:28+05:00","level":"INFO","msg":"init","logger":"log/slog","format":"json","caller":"logger/main.go:13"}

	{
		logger := logger.NewGoKitLogger("info")
		logger.Log("msg", "init", "logger", "go-kit/log", "format", "json")
		// {"time":"2024-04-26T21:11:28+05:00","level":"INFO","msg":"init","logger":"go-kit/log","format":"json","caller":"logger/main.go:18"}
	}

	{
		logger := logger.NewGormLogger("info")
		logger.Info(context.Background(), "init %s %s %s %s", "logger", "gorm.io/gorm/logger", "format", "json")
		// {"time":"2024-04-26T21:11:28+05:00","level":"INFO","msg":"init logger gorm.io/gorm/logger format json","caller":"logger/main.go:24"}
	}

	logger.NewLogger(os.Stdout)
	slog.Info("init", "logger", "log/slog", "format", "text")
	// time=2024-04-26T21:11:28+05:00 level=INFO msg=init logger=log/slog format=text caller=logger/main.go:29

	{
		logger := logger.NewGoKitLogger("info")
		logger.Log("msg", "init", "logger", "go-kit/log", "format", "text")
		// time=2024-04-26T21:11:28+05:00 level=INFO msg=init logger=go-kit/log format=text caller=logger/main.go:34
	}

	{
		logger := logger.NewGormLogger("info")
		logger.Info(context.Background(), "init %s %s %s %s", "logger", "gorm.io/gorm/logger", "format", "text")
		// time=2024-04-26T21:11:28+05:00 level=INFO msg="init logger gorm.io/gorm/logger format text" caller=logger/main.go:40
	}
}
