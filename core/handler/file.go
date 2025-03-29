package handler

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileOptions configures file handler behavior
type FileOptions struct {
	Path       string
	MaxSize    int64         // maximum size in megabytes
	MaxAge     int           // maximum age in days
	MaxBackups int           // maximum number of old log files to retain
	Interval   time.Duration // interval between rotations
}

// FileHandler manages log file output with rotation
type FileHandler struct {
	handler    slog.Handler
	opts       FileOptions
	mu         sync.Mutex
	file       *os.File
	size       int64
	lastRotate time.Time
	stopChan   chan struct{}
	doneChan   chan struct{}
}

func NewFileHandler(opts FileOptions) (*FileHandler, error) {
	if err := validateFileOptions(&opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	f, err := openFile(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("stat file: %w", err)
	}

	h := &FileHandler{
		opts:       opts,
		file:       f,
		size:       info.Size(),
		lastRotate: info.ModTime(),
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
	}

	// Create base handler for the file
	h.handler = New(f, &Options{
		JSON:       true, // Default to JSON for files
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
		AddSource:  true,
	})

	// Start rotation goroutine if interval is specified
	if opts.Interval > 0 {
		go h.rotationWorker()
	}

	return h, nil
}

func (h *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *FileHandler) Close() error {
	close(h.stopChan)
	<-h.doneChan // Wait for rotation worker to finish

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.file != nil {
		if err := h.file.Sync(); err != nil {
			return fmt.Errorf("sync file: %w", err)
		}
		if err := h.file.Close(); err != nil {
			return fmt.Errorf("close file: %w", err)
		}
		h.file = nil
	}
	return nil
}

func (h *FileHandler) rotate() error {
	if h.file == nil {
		return nil
	}

	// Sync and close current file
	if err := h.file.Sync(); err != nil {
		return fmt.Errorf("sync current file: %w", err)
	}
	if err := h.file.Close(); err != nil {
		return fmt.Errorf("close current file: %w", err)
	}
	h.file = nil

	// Rotate files
	if err := h.rotateFiles(); err != nil {
		return fmt.Errorf("rotate files: %w", err)
	}

	// Open new file
	file, err := openFile(h.opts.Path)
	if err != nil {
		return fmt.Errorf("open new file: %w", err)
	}

	// Update handler state
	h.file = file
	h.size = 0
	h.lastRotate = time.Now()
	return nil
}

func (h *FileHandler) rotateFiles() error {
	// Check if rotation is needed
	if !h.shouldRotate() {
		return nil
	}

	// Remove old backups first
	if err := h.removeOldBackups(); err != nil {
		return fmt.Errorf("remove old backups: %w", err)
	}

	// Shift existing backups
	for i := h.opts.MaxBackups - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", h.opts.Path, i)
		newPath := fmt.Sprintf("%s.%d", h.opts.Path, i+1)

		// Ignore errors for missing files
		if err := os.Rename(oldPath, newPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rename %s to %s: %w", oldPath, newPath, err)
		}
	}

	// Move current file to .1
	backupPath := h.opts.Path + ".1"
	if err := os.Rename(h.opts.Path, backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rename current file: %w", err)
	}

	return nil
}

func (h *FileHandler) removeOldBackups() error {
	for i := h.opts.MaxBackups + 1; ; i++ {
		path := fmt.Sprintf("%s.%d", h.opts.Path, i)
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				break
			}
			return fmt.Errorf("remove old backup %s: %w", path, err)
		}
	}
	return nil
}

func (h *FileHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ensure file is open
	if h.file == nil {
		if err := h.rotate(); err != nil {
			return fmt.Errorf("rotate on handle: %w", err)
		}
	}

	// Write to file
	data := h.formatRecord(r)
	n, err := h.file.Write(data)
	if err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	// Update size and check rotation
	h.size += int64(n)
	if h.shouldRotate() {
		if err := h.rotate(); err != nil {
			return fmt.Errorf("rotate after write: %w", err)
		}
	}

	return nil
}

func (h *FileHandler) shouldRotate() bool {
	if h.opts.MaxSize > 0 && h.size >= h.opts.MaxSize {
		return true
	}
	if h.opts.Interval > 0 && time.Since(h.lastRotate) >= h.opts.Interval {
		return true
	}
	return false
}

func (h *FileHandler) rotationWorker() {
	ticker := time.NewTicker(h.opts.Interval)
	defer ticker.Stop()
	defer close(h.doneChan)

	for {
		select {
		case <-ticker.C:
			h.mu.Lock()
			if err := h.rotate(); err != nil {
				slog.Error("rotate log file",
					"error", err,
					"path", h.opts.Path,
				)
			}
			h.mu.Unlock()
		case <-h.stopChan:
			return
		}
	}
}

func validateFileOptions(opts *FileOptions) error {
	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	if opts.Path == "" {
		return fmt.Errorf("file path is required")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(opts.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Convert maxSize to bytes for internal use
	if opts.MaxSize > 0 {
		opts.MaxSize *= 1024 * 1024 // Convert MB to bytes
	} else {
		opts.MaxSize = 100 * 1024 * 1024 // 100MB default
	}

	if opts.MaxAge <= 0 {
		opts.MaxAge = 7 // 7 days default
	}

	if opts.MaxBackups <= 0 {
		opts.MaxBackups = 5 // 5 backups default
	}

	return nil
}

func openFile(path string) (*os.File, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// Open file with appropriate permissions
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

func (h *FileHandler) formatRecord(r slog.Record) []byte {
	// Create a memory buffer and a handler to write into it
	var buf []byte
	bufWriter := &bufferWriter{&buf}
	baseHandler, ok := h.handler.(*BaseHandler)
	timeFormat := time.RFC3339
	addSource := true
	if ok {
		timeFormat = baseHandler.opts.TimeFormat
		addSource = baseHandler.opts.AddSource
	}

	// Create a custom handler that ensures time field is first
	memHandler := New(bufWriter, &Options{
		JSON:       true,
		Level:      slog.LevelDebug, // Always log all records
		TimeFormat: timeFormat,
		AddSource:  addSource,
	})

	// Try to format using the memory handler
	if err := memHandler.Handle(context.Background(), r); err != nil {
		// If JSON formatting fails, fallback to simple format with time first
		buf = []byte(fmt.Sprintf("[%s] %s: %s\n",
			r.Time.Format(time.RFC3339),
			r.Level,
			r.Message,
		))
	}

	// Ensure newline at the end
	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}

	return buf
}

// bufferWriter is an io.Writer that writes to a byte slice
type bufferWriter struct {
	buf *[]byte
}

func (w *bufferWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// WithAttrs returns a new FileHandler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new FileHandler with the same options
	newHandler := &FileHandler{
		handler:    h.handler.WithAttrs(attrs),
		opts:       h.opts,
		mu:         sync.Mutex{},
		file:       h.file, // Share the file handle
		size:       h.size,
		lastRotate: h.lastRotate,
		stopChan:   h.stopChan, // Share the stop channel
		doneChan:   h.doneChan, // Share the done channel
	}

	return newHandler
}

// WithGroup returns a new FileHandler with the given group appended to
// the receiver's existing groups.
func (h *FileHandler) WithGroup(name string) slog.Handler {
	// Create a new FileHandler with the same options
	newHandler := &FileHandler{
		handler:    h.handler.WithGroup(name),
		opts:       h.opts,
		mu:         sync.Mutex{},
		file:       h.file, // Share the file handle
		size:       h.size,
		lastRotate: h.lastRotate,
		stopChan:   h.stopChan, // Share the stop channel
		doneChan:   h.doneChan, // Share the done channel
	}

	return newHandler
}
