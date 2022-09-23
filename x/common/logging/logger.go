package logging

import (
	"os"
	"runtime"

	"cloud.google.com/go/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger sets the proper logging variables
func InitLogger(version string, debug, human bool) {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if human {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	log.Logger = log.Logger.Hook(googleSeverityHook{})
	log.Logger = log.With().
		Str("version", version).
		Str("goversion", runtime.Version()).
		Logger()
}

type googleSeverityHook struct{}

// Run --> ToDo: confirm if msg is supposed to be ignored here
func (h googleSeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("severity", levelToSeverity(level).String())
}

// converts zerolog level to google's severity.
func levelToSeverity(level zerolog.Level) logging.Severity {
	switch level {
	case zerolog.DebugLevel:
		return logging.Debug
	case zerolog.WarnLevel:
		return logging.Warning
	case zerolog.ErrorLevel:
		return logging.Error
	case zerolog.FatalLevel:
		return logging.Alert
	case zerolog.PanicLevel:
		return logging.Emergency
	default:
		return logging.Info
	}
}

// FileOutput This should get deleted later, this is just a quick hack for me to get logs from a goroutine
// func FileOutput(msg string) {
// 	// create a temp file
// 	tempFile, err := os.CreateTemp(os.TempDir(), "deleteme")
// 	if err != nil {
// 		// Can we log an error before we have our logger? :)
// 		log.Error().Err(err).Msg("there was an error creating a temporary file four our log")
// 	}
// 	fileLogger := zerolog.New(tempFile).With().Logger()
// 	fileLogger.Info().Msg(msg)

// 	fmt.Printf("The log file is allocated at %s\n", tempFile.Name())
// }
