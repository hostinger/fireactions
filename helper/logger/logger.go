package logger

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

// New returns a new logger with the given configuration.
func New(level string) (*zerolog.Logger, error) {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}

		file = short
		return file + ":" + strconv.Itoa(line)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &logger, nil
}
