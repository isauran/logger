package fields

import (
	"fmt"
	"runtime"
	"strings"
)

// formatStack formats a stack trace starting from the given skip level
func formatStack(skip int) string {
	const depth = 32
	var pcs [depth]uintptr

	// +1 for this function, +1 for runtime.Callers
	n := runtime.Callers(skip+2, pcs[:])
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var builder strings.Builder

	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// Skip runtime functions
		if strings.Contains(frame.Function, "runtime.") {
			continue
		}

		fmt.Fprintf(&builder, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
	}

	return builder.String()
}
