package comm

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func Logger(packageName string) zerolog.Logger {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	writer.FormatMessage = func(i interface{}) string {
		return "\n" + i.(string)
	}
	writer.FormatFieldName = func(i interface{}) string {
		return " " + i.(string) + ": "
	}

	return zerolog.New(writer).With().Str("package", packageName).Caller().Timestamp().Logger()

}
