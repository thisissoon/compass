package logger

import (
	"io"
	"io/ioutil"
	"os"

	"compass/config"

	"github.com/rs/zerolog"
)

// New returns a new logger setup from configuration
func New() zerolog.Logger {
	f := map[string]interface{}{
		"version": config.Version(),
		"commit":  config.Commit(),
	}
	var w io.Writer = os.Stdout
	switch LogFormat() {
	case "console", "terminal":
		w = zerolog.ConsoleWriter{
			Out: os.Stdout,
		}
	case "discard":
		w = ioutil.Discard
	}
	return zerolog.New(w).
		With().
		Timestamp().
		Fields(f).
		Logger()
}
