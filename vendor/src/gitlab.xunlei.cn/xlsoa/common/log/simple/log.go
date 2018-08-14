package log

import (
	"fmt"
	xlsoa_log "gitlab.xunlei.cn/xlsoa/common/log"
	"io"
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
	level  xlsoa_log.Level
}

func New(w io.Writer) *Logger {
	return &Logger{
		logger: log.New(w, "", log.LstdFlags|log.Lshortfile),
		level:  xlsoa_log.LevelInfo,
	}
}

func (l *Logger) SetLevel(level xlsoa_log.Level) {
	l.level = level
}

func (l *Logger) Debug(args ...interface{}) {
	l.log(xlsoa_log.LevelDebug, args...)
}

func (l *Logger) Debugf(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelDebug, fmt, args...)
}

func (l *Logger) Trace(args ...interface{}) {
	l.log(xlsoa_log.LevelTrace, args...)
}

func (l *Logger) Tracef(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelTrace, fmt, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.log(xlsoa_log.LevelInfo, args...)
}

func (l *Logger) Infof(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelInfo, fmt, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.log(xlsoa_log.LevelWarn, args...)
}

func (l *Logger) Warnf(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelWarn, fmt, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.log(xlsoa_log.LevelError, args...)
}

func (l *Logger) Errorf(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelError, fmt, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log(xlsoa_log.LevelFatal, args...)
}

func (l *Logger) Fatalf(fmt string, args ...interface{}) {
	l.logf(xlsoa_log.LevelFatal, fmt, args...)
}

func (l *Logger) log(level xlsoa_log.Level, args ...interface{}) {
	if level < l.level {
		return
	}
	args1 := []interface{}{
		fmt.Sprintf("[%v] ", level),
	}
	args1 = append(args1, args...)
	l.logger.Output(3, fmt.Sprint(args1...))
}
func (l *Logger) logf(level xlsoa_log.Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	format1 := fmt.Sprintf("[%v] ", level)
	format1 += format
	l.logger.Output(3, fmt.Sprintf(format1, args...))
}

// Default logger
var defaultLogger = New(os.Stdout)

func SetOutput(w io.Writer) {
	defaultLogger.logger.SetOutput(w)
}

func SetLevel(level xlsoa_log.Level) {
	defaultLogger.SetLevel(level)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Debugf(fmt string, args ...interface{}) {
	defaultLogger.Debugf(fmt, args...)
}

func Trace(args ...interface{}) {
	defaultLogger.Trace(args...)
}

func Tracef(fmt string, args ...interface{}) {
	defaultLogger.Tracef(fmt, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Infof(fmt string, args ...interface{}) {
	defaultLogger.Infof(fmt, args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Warnf(fmt string, args ...interface{}) {
	defaultLogger.Warnf(fmt, args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Errorf(fmt string, args ...interface{}) {
	defaultLogger.Errorf(fmt, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Fatalf(fmt string, args ...interface{}) {
	defaultLogger.Fatalf(fmt, args...)
}
