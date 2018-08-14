package log

import (
	"fmt"
	"os"
	"bufio"
)

var (
	defaultLog *Log
)

func SetDefaultLogMgr(logMgr *Log) {
	defaultLog = logMgr
}

func FormatAll(callerSkip int, level LogLevel, format string, params...interface{}) {
	if defaultLog != nil {
		defaultLog.Format(callerSkip+1, level, format, params...)
	} else {
		content := Format(callerSkip+1, level, format, params...)
		fmt.Print(content)
	}
}

func Tracef(format string, params...interface{}) {
	FormatAll(2, E_LOG_LEVEL_TRACE, format, params...)
}

func Debugf(format string, params...interface{}) {
	FormatAll(2, E_LOG_LEVEL_DEBUG, format, params...)
}

func Infof(format string, params...interface{}) {
	FormatAll(2, E_LOG_LEVEL_INFO, format, params...)
}

func Warnf(format string, params...interface{}) {
	FormatAll(2, E_LOG_LEVEL_WARN, format, params...)
}

func Errorf(format string, params...interface{}) {
	FormatAll(2, E_LOG_LEVEL_ERROR, format, params...)
}

func Flush()  {
	if defaultLog != nil {
		defaultLog.Flush()
	} else {
		f := bufio.NewWriter(os.Stdout)
		f.Flush()
	}
}
