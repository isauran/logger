package file

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/isauran/logger/core/handler"
)

func ExampleFileRotation() {
	// Use a file in the current directory for better visibility
	currentDir, _ := os.Getwd()
	logPath := filepath.Join(currentDir, "example.log")

	fileOpts := handler.FileOptions{
		Path:       logPath,
		MaxSize:    100, // MB
		MaxAge:     7,   // days
		MaxBackups: 5,   // files
		Interval:   24 * time.Hour,
	}

	// Create multi-writer handler to write to both file and stdout
	fileHandler, err := handler.NewFileHandler(fileOpts)
	if err != nil {
		slog.Error("failed to create file handler", "error", err)
		os.Exit(1)
	}
	defer fileHandler.Close()

	// Create logger with file handler
	logger := slog.New(fileHandler)

	// Log several messages to demonstrate rotation
	for i := 0; i < 5; i++ {
		logger.Info("log message to file",
			"iteration", i,
			"app", "example",
			"env", "development",
			"timestamp", time.Now().Format(time.RFC3339),
		)
		time.Sleep(100 * time.Millisecond) // Small delay between logs
	}

	// Print the contents of the log file
	fmt.Println("\nContents of the log file:")
	content, err := os.ReadFile(logPath)
	if err != nil {
		slog.Error("failed to read log file", "error", err)
		return
	}
	fmt.Println(string(content))

	// Clean up the example log file
	os.Remove(logPath)
}
