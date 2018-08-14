package log

type Level int

func (l Level) String() string {
	return LevelName[l]
}

const (
	LevelAll   Level = 0
	LevelDebug Level = 1
	LevelTrace Level = 2
	LevelInfo  Level = 3
	LevelWarn  Level = 4
	LevelError Level = 5
	LevelFatal Level = 6
)

var LevelName = map[Level]string{
	LevelAll:   "ALL",
	LevelDebug: "DEBUG",
	LevelTrace: "TRACE",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

var NameToLevel = map[string]Level{
	"ALL":   LevelAll,
	"DEBUG": LevelDebug,
	"TRACE": LevelTrace,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
	"FATAL": LevelFatal,
}

type Logger interface {
	SetLevel(level Level)

	Debug(args ...interface{})
	Debugf(fmt string, args ...interface{})

	Trace(args ...interface{})
	Tracef(fmt string, args ...interface{})

	Info(args ...interface{})
	Infof(fmt string, args ...interface{})

	Warn(args ...interface{})
	Warnf(fmt string, args ...interface{})

	Error(args ...interface{})
	Errorf(fmt string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(fmt string, args ...interface{})
}
