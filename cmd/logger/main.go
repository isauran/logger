package main

import (
	"os"
	"github.com/isauran/slogger"
)

func main() {
	logJSON := slogger.NewLogger(os.Stdout, slogger.WithJSON(true))
	logJSON.Info("init", "logger", "log/slog", "format", "json")
	// {"time":"2024-03-17T02:01:14+05:00","level":"INFO","caller":"main.go:17","msg":"init","logger":"log/slog","format":"json"}

	log := slogger.NewLogger(os.Stdout)
	log.Info("init", "logger", "log/slog", "format", "text")
	// time=2024-03-17T02:01:14+05:00 level=INFO caller=main.go:27 msg=init logger=log/slog format=text
}
