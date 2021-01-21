package log

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/pkgerrors"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var logger *logWrapper
var once sync.Once

// logWrapper wrap the zerologger for logging in different destinations
type logWrapper struct {
	logCore *zerolog.Logger
	appName string
	host    string
}

func getLogger() *logWrapper {
	once.Do(func() {
		if logger == nil {
			debug.PrintStack()
			log.Fatal("please initialize the logger once before using it with log.Init (). This can be done in main.go")
		}
	})
	return logger
}

func (l *logWrapper) standart(e *zerolog.Event, message interface{}, args ...interface{}) {
	switch mess := message.(type) {
	case error:
		//errMess := fmt.Sprintf("%+v", mess)
		e.Str("reason", "error").Stack().Err(mess).Msg("")
	case string:
		e.Str("reason", "info").Stack().Msgf(mess, args...) //todo: may be in debug mode set request id
	default:
		e.Str("reason", "error").Stack().Msgf(fmt.Sprintf("message %v has unknown type %v", message, mess), args...) //todo: may be in debug mode set request id
	}
}

func (l *logWrapper) debug(message interface{}, args ...interface{}) {
	l.standart(l.logCore.Debug(),message,args...)
	/*switch mess := message.(type) {
	case error:
		//errMess := fmt.Sprintf("%+v", mess)
		l.logCore.Debug().Str("reason", "error").Stack().Err(mess).Msg("")
	case string:
		l.logCore.Debug().Str("reason", "info").Stack().Msgf(mess, args...) //todo: may be in debug mode set request id
	default:
		l.logCore.Debug().Str("reason", "error").Stack().Msgf(fmt.Sprintf("message %v has unknown type %v", message, mess), args...) //todo: may be in debug mode set request id
	}*/
}

func (l *logWrapper) info(message interface{}, args ...interface{}) {
	switch mess := message.(type) {
	case error:
		//errMess := fmt.Sprintf("%+v", mess)
		l.logCore.Info().Str("reason", "error").Stack().Err(mess).Msg("")
	case string:
		l.logCore.Info().Str("reason", "info").Msgf(mess, args...)
	default:
		l.logCore.Info().Str("reason", "error").Msgf(fmt.Sprintf("message %v has unknown type %v", message, mess), args...)
	}
}

func (l *logWrapper) error(message interface{}, args ...interface{}) {
	switch mess := message.(type) {
	case error:
		//errMess := fmt.Sprintf("%+v", mess)
		l.logCore.Error().Str("reason", "error").Stack().Err(mess).Msg("")
	case string:
		l.logCore.Error().Str("reason", "info").Msgf(mess, args...)
	default:
		l.logCore.Error().Str("reason", "error").Msgf(fmt.Sprintf("message %v has unknown type %v", message, mess), args...)
	}
}

// InitWithStdout initialize logger with stdout writer and make beauty output.
// Use rtb/log in  your imports.
func InitWithStdout(logLevel, appName,env string) error {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    runtime.GOOS == "windows",
		TimeFormat: time.RFC3339,
	}
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("***%s****", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%s", i))
	}
	return Init(output, logLevel, appName,env)
}

// InitWithFile initialize logger with file writer.
// Use rtb/log in  your imports.
func InitWithFile(logfile, logLevel, appName,env string) error {
	wr, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "can't create/open log file")
	}
	return Init(wr, logLevel, appName,env)
}

// Init initialize logger with generic writer.
// Use rtb/log in  your imports.
func Init(logWriter io.Writer, logLevel, appName,env string) error {
	globalLL := zerolog.InfoLevel
	wr := diode.NewWriter(logWriter, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger Dropped %d messages", missed)
	})
	logl := zerolog.New(wr)

	switch logLevel {
	case "disable":
		globalLL = zerolog.Disabled
	case "error":
		globalLL = zerolog.ErrorLevel
	case "info":
		globalLL = zerolog.InfoLevel
	case "debug":
		globalLL = zerolog.DebugLevel
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	default:
		log.Printf("unknown log level %v, I know levels: disable,error,info,debug. Setting default level:info\n", logLevel)
	}

	zerolog.MessageFieldName = "short_message"
	zerolog.ErrorFieldName = "short_message"
	zerolog.ErrorStackFieldName = "full_message"
	zerolog.LevelFieldName = "loglevel"
	zerolog.SetGlobalLevel(globalLL)
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	logl = logl.With().Timestamp().Float64("version", 1.1).
		Str("host", host).
		Str("_app", appName).Str("_env", env) /*.Int64("timestamp", time.Now().Unix())*/ .Logger().Level(globalLL)

	logger = &logWrapper{
		logCore: &logl,
	}
	return nil
}
// Debug using debugging level for verbose the messages
// for stacktrace use error package github.com/pkg/errors in your imports.
// Use rtb/log in  your imports. Initialize logger before first using.
func Debug(message interface{}, args ...interface{}) {
	getLogger().debug(message, args...)
}
// Info using Info level for logging messages
// Use rtb/log in  your imports. Initialize logger before first using.
func Info(message interface{}, args ...interface{}) {
	getLogger().info(message, args...)
}

// Error using Error level for logging error messages.
// Use rtb/log in  your imports. Initialize logger before first using.
func Error(message interface{}, args ...interface{}) {
	getLogger().error(message, args...)
}
// Println using Info level for logging messages
// Use rtb/log in  your imports. Initialize logger before first using.
func Println(message interface{}, args ...interface{}) {
	getLogger().info(message, args...)
}
// Print using Info level for logging messages
// Use rtb/log in  your imports. Initialize logger before first using.
func Print(message interface{}, args ...interface{}) {
	getLogger().info(message, args...)
}
// Fatal using Error level for logging fatal error messages and exit from app.
// Use rtb/log in  your imports. Initialize logger before first using.
func Fatal(err error) {
	go getLogger().logCore.Error().Str("reason", "fatal error").Stack().Err(err).Msg("")
	time.Sleep(time.Second)
	os.Exit(1)
}

