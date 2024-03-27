package main

import (
	"os"

	"github.com/isauran/logger"
)

func main() {
	logJSON := logger.NewLogger(os.Stdout, logger.WithJSON(true))
	logJSON.Info("init", "logger", "log/slog", "format", "json")
	// {"time":"2024-03-27T15:05:29+05:00","level":"INFO","caller":"main.go:11","msg":"init","logger":"log/slog","format":"json"}

	log := logger.NewLogger(os.Stdout)
	log.Info("init", "logger", "log/slog", "format", "text")
	// time=2024-03-27T15:05:29+05:00 level=INFO caller=main.go:15 msg=init logger=log/slog format=text

	gokitlogJSON := logger.NewGoKitLogger(os.Stdout, logger.WithJSON(true))
	gokitlogJSON.Log("msg", "init", "logger", "go-kit/log", "format", "json")
	// {"time":"2024-03-27T15:05:29+05:00","level":"INFO","caller":"gokitlog.go:12","msg":"init","logger":"go-kit/log","format":"json"}

	gokitlog := logger.NewGoKitLogger(os.Stdout)
	gokitlog.Log("msg", "init", "logger", "go-kit/log", "format", "text")
	// time=2024-03-27T15:05:29+05:00 level=INFO caller=gokitlog.go:12 msg=init logger=go-kit/log format=text
}
