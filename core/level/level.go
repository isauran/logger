package level

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// CustomLevel represents a user-defined logging level
type CustomLevel struct {
	Level    slog.Level
	Name     string
	Color    string
	Icon     string
	Metadata map[string]interface{}
}

// LevelRegistry manages custom logging levels
type LevelRegistry struct {
	mu      sync.RWMutex
	levels  map[slog.Level]*CustomLevel
	aliases map[string]slog.Level
}

var (
	defaultRegistry = NewRegistry()
	defaultLevels   = map[slog.Level]*CustomLevel{
		slog.LevelDebug: {
			Level: slog.LevelDebug,
			Name:  "DEBUG",
			Color: "\033[36m", // Cyan
			Icon:  "🔍",
		},
		slog.LevelInfo: {
			Level: slog.LevelInfo,
			Name:  "INFO",
			Color: "\033[32m", // Green
			Icon:  "ℹ️",
		},
		slog.LevelWarn: {
			Level: slog.LevelWarn,
			Name:  "WARN",
			Color: "\033[33m", // Yellow
			Icon:  "⚠️",
		},
		slog.LevelError: {
			Level: slog.LevelError,
			Name:  "ERROR",
			Color: "\033[31m", // Red
			Icon:  "❌",
		},
	}
)

// Validation constants
const (
	maxLevelValue = 1000
	minLevelValue = -1000
)

// NewRegistry creates a new level registry
func NewRegistry() *LevelRegistry {
	r := &LevelRegistry{
		levels:  make(map[slog.Level]*CustomLevel),
		aliases: make(map[string]slog.Level),
	}

	// Register default levels with validation
	for _, level := range defaultLevels {
		if err := validateLevel(level); err != nil {
			panic(fmt.Sprintf("invalid default level: %v", err))
		}
		r.RegisterLevel(level)
	}

	return r
}

// validateLevel checks if a custom level is valid
func validateLevel(level *CustomLevel) error {
	if level == nil {
		return fmt.Errorf("level cannot be nil")
	}
	if level.Name == "" {
		return fmt.Errorf("level name cannot be empty")
	}
	if level.Level < slog.Level(minLevelValue) || level.Level > slog.Level(maxLevelValue) {
		return fmt.Errorf("level value %d out of valid range [%d, %d]", level.Level, minLevelValue, maxLevelValue)
	}
	return nil
}

// RegisterLevel adds a custom level to the registry with validation
func (r *LevelRegistry) RegisterLevel(level *CustomLevel) error {
	if err := validateLevel(level); err != nil {
		return fmt.Errorf("invalid level: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for existing level with same value
	if existing, ok := r.levels[level.Level]; ok {
		return fmt.Errorf("level value %d already registered as %q", level.Level, existing.Name)
	}

	// Check for existing name
	upperName := strings.ToUpper(level.Name)
	for _, existing := range r.levels {
		if strings.ToUpper(existing.Name) == upperName {
			return fmt.Errorf("level name %q already registered", level.Name)
		}
	}

	r.levels[level.Level] = level
	r.aliases[upperName] = level.Level
	return nil
}

// GetLevel retrieves level information by level value with read lock
func (r *LevelRegistry) GetLevel(level slog.Level) *CustomLevel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if l, ok := r.levels[level]; ok {
		return l
	}
	return &CustomLevel{Level: level, Name: level.String()}
}

// ParseLevel converts a level string using the registry with read lock
func (r *LevelRegistry) ParseLevel(levelStr string) (slog.Level, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	upperStr := strings.ToUpper(levelStr)
	if level, ok := r.aliases[upperStr]; ok {
		return level, nil
	}

	// Try parsing as a numeric value
	var l slog.Level
	if err := l.UnmarshalText([]byte(levelStr)); err != nil {
		return 0, fmt.Errorf("unknown level %q", levelStr)
	}
	return l, nil
}
