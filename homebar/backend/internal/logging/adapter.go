package logging

import (
	"log"

	"github.com/rs/zerolog"
)

// StdLogAdapter adapts zerolog to standard log interface
type StdLogAdapter struct {
	zl zerolog.Logger
}

func (a *StdLogAdapter) Write(p []byte) (n int, err error) {
	a.zl.Info().Msg(string(p))
	return len(p), nil
}

// NewStdLogAdapter creates a standard logger that writes to zerolog
func NewStdLogAdapter(logger zerolog.Logger, prefix string) *log.Logger {
	adapter := &StdLogAdapter{zl: logger.With().Str("component", prefix).Logger()}
	return log.New(adapter, "", 0) // No prefix/flags as zerolog handles it
}
