package log

import (
	xlsoa_log "gitlab.xunlei.cn/xlsoa/common/log"
	"os"
	"testing"
)

func TestStdout(t *testing.T) {

	for _, level := range []xlsoa_log.Level{
		xlsoa_log.LevelAll,
		xlsoa_log.LevelDebug,
		xlsoa_log.LevelTrace,
		xlsoa_log.LevelInfo,
		xlsoa_log.LevelWarn,
		xlsoa_log.LevelError,
		xlsoa_log.LevelFatal,
	} {
		logger := New(os.Stdout)
		logger.SetLevel(level)
		logTest(logger)
	}

}

func TestFile(t *testing.T) {

	for _, level := range []xlsoa_log.Level{
		xlsoa_log.LevelAll,
		xlsoa_log.LevelDebug,
		xlsoa_log.LevelTrace,
		xlsoa_log.LevelInfo,
		xlsoa_log.LevelWarn,
		xlsoa_log.LevelError,
		xlsoa_log.LevelFatal,
	} {
		logger, err := NewWithFile("./tmp.log")
		if err != nil {
			t.Fatalf("NewWithFile error: %v", err)
		}
		logger.SetLevel(level)
		logTest(logger)
	}

}
func logTest(logger xlsoa_log.Logger) {

	logger.Debug("This is Debug")
	logger.Debugf("This is Debugf %v", 1000)

	logger.Trace("This is Trace")
	logger.Tracef("This is Tracef %v", 1000)

	logger.Info("This is Info")
	logger.Infof("This is Infof %v", 1000)

	logger.Warn("This is Warn")
	logger.Warnf("This is Warnf %v", 1000)

	logger.Error("This is Error")
	logger.Errorf("This is Errorf %v", 1000)

	logger.Fatal("This is Fatal")
	logger.Fatalf("This is Fatalf %v", 1000)
}

func TestMain(m *testing.M) {

	ret := m.Run()
	os.Remove("./tmp.log")
	os.Exit(ret)

}
