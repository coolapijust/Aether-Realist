package util

import (
	"io"
	"strings"
)

// FilteredWriter is an io.Writer that suppresses specific lines based on substrings.
type FilteredWriter struct {
	Target  io.Writer
	Filters []string
}

// NewFilteredWriter creates a new FilteredWriter.
func NewFilteredWriter(target io.Writer, filters []string) *FilteredWriter {
	return &FilteredWriter{
		Target:  target,
		Filters: filters,
	}
}

func (f *FilteredWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	lowerMsg := strings.ToLower(msg)

	for _, filter := range f.Filters {
		if strings.Contains(lowerMsg, strings.ToLower(filter)) {
			// Skip this write, but return the original length to satisfy io.Writer
			return len(p), nil
		}
	}

	return f.Target.Write(p)
}
