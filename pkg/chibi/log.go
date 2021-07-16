package chibi

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initialize the global logger of github.com/rs/zerolog/log.
func InitLogger(level string, pretty bool) error {
	var l zerolog.Level

	if level != "" {
		lv, err := zerolog.ParseLevel(level)
		if err != nil {
			return err
		}

		l = lv
	} else {
		l = zerolog.InfoLevel
	}

	var w io.Writer
	if pretty {
		w = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		w = os.Stdout
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(l)
	log.Logger = zerolog.New(w).With().Timestamp().Caller().Logger().Hook(severityHook{})

	return nil
}

type severityHook struct{}

func (h severityHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	if l != zerolog.NoLevel {
		var s string
		switch l {
		case zerolog.TraceLevel:
			s = "DEFAULT"
		case zerolog.DebugLevel:
			s = "DEBUG"
		case zerolog.InfoLevel:
			s = "INFO"
		case zerolog.WarnLevel:
			s = "WARNING"
		case zerolog.ErrorLevel:
			s = "ERROR"
		case zerolog.FatalLevel:
			s = "CRITICAL"
		case zerolog.PanicLevel:
			s = "EMERGENCY"
		}
		e.Str("severity", s)
	}
}
