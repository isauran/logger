package context

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
)

type sourceKey struct{}

// WithSource adds source location information to context
func WithSource(ctx context.Context, skip int) context.Context {
	return context.WithValue(ctx, sourceKey{}, getSource(skip+1))
}

// GetSource retrieves source information from context
func GetSource(ctx context.Context) *slog.Source {
	if src, ok := ctx.Value(sourceKey{}).(*slog.Source); ok {
		return src
	}
	return nil
}

// getSource returns the caller's source location
func getSource(skip int) *slog.Source {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}
	return &slog.Source{
		File: file,
		Line: line,
	}
}

// FormatSource formats source location in a standardized way
func FormatSource(s *slog.Source) string {
	if s == nil {
		return ""
	}
	return filepath.Join(filepath.Base(filepath.Dir(s.File)), filepath.Base(s.File)) + ":" + fmt.Sprint(s.Line)
}
