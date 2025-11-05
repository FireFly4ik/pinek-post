package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strings"
)

func Setup(productionType string) *zerolog.Logger {
	if productionType == "prod" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	zerolog.TimeFieldFormat = "15:04:05 02.01.2006"

	zerolog.CallerMarshalFunc = func(pc uintptr, post string, line int) string {
		parts := strings.Split(post, "/")
		if len(parts) > 2 {
			post = strings.Join(parts[len(parts)-2:], "/")
		}
		return fmt.Sprintf("%s:%d", post, line)
	}

	var writer io.Writer = os.Stdout

	loggerContext := zerolog.New(writer).
		With().
		Caller().
		Timestamp().
		Logger()

	log.Info().Msg("logger setup complete")
	return &loggerContext
}
