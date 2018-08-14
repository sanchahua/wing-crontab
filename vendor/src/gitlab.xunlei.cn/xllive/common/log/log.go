package log

import (
	"gitlab.xunlei.cn/xllive/common/utils"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"bufio"
)

type LogConfig struct{
	WriteConsole bool     `yaml:"write_console"`
	WriteFile bool        `yaml:"write_file"`
	OpenTrace bool        `yaml:"open_trace"`
	OpenDebug bool        `yaml:"open_debug"`
	OpenInfo bool         `yaml:"open_info"`
	OpenWarn bool         `yaml:"open_warn"`
	OpenError bool        `yaml:"open_error"`
	FilenameLayout string `yaml:"filename_layout"`
	Path string           `yaml:"path"`
	MaxRemainHours int64  `yaml:"max_remain_hours"`
}

type Log struct {
	logConfig    string
	cfgModTime   int64
	fileChan     chan string
	flushChan    chan bool
	cfgTimer     *time.Ticker
	newFileTimer *time.Timer
	bRuning      bool
	remainHours  int64

	level           LogLevel
	file            *os.File
	filename_layout string
	path            string
}

type LogLevel uint64

const (
	E_LOG_LEVEL_TRACE = 0x1
	E_LOG_LEVEL_DEBUG = 0x10
	E_LOG_LEVEL_INFO  = 0x20
	E_LOG_LEVEL_WARN  = 0x40
	E_LOG_LEVEL_ERROR = 0x80

	E_LOG_WRITE_FILE    = 0x100
	E_LOG_WRITE_CONSOLE = 0x200
)

func NewLogMgr(logConfig string) (*Log, error) {
	Infof("NewLogMgr beg, logConfig=[%s]", logConfig)
	m := new(Log)
	err := m.init(logConfig)
	Infof("NewLogMgr end")
	return m, err
}

func (l *Log) Flush() {
	l.flushChan <- true
}

func (l *Log)flush()  {
	f := bufio.NewWriter(l.file)
	f.Flush()
	f = bufio.NewWriter(os.Stdout)
	f.Flush()
}

func (l *Log) init(logConfig string) error {

	l.bRuning = true

	/// 重新加载配置文件的定时器
	var err error
	if l.logConfig, err = filepath.Abs(logConfig); err != nil {
		Errorf("init filepath.Abs fail, logConfig=[%s] err=[%v]", logConfig, err)
		return err
	}

	/// 配置文件
	if err = l.loadConfig(); err != nil {
		Errorf("init loadConfig fail, logConfig=[%s] err=[%v]", logConfig, err)
		return err
	}
	l.cfgTimer = time.NewTicker(1 * time.Minute)

	/// 创建新文件
	if err = l.newFile(); err != nil {
		Errorf("init newFile fail, logConfig=[%s] err=[%v]", logConfig, err)
		return err
	}
	l.newFileTimer = time.NewTimer(l.newFileNextDuration())

	l.fileChan = make(chan string, 10000)

	l.flushChan = make(chan bool)
	go l.run()
	return nil
}

func (l *Log) newFileNextDuration() time.Duration {
	nowTime := time.Now()
	nextTime := time.Now()
	if true == strings.Contains(l.filename_layout, "05") {
		nextTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), nowTime.Minute(), nowTime.Second()+1, 0, time.Local)
	} else if true == strings.Contains(l.filename_layout, "04") {
		nextTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), nowTime.Minute()+1, 0, 0, time.Local)
	} else if true == strings.Contains(l.filename_layout, "15") {
		nextTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour()+1, 0, 0, 0, time.Local)
	} else if true == strings.Contains(l.filename_layout, "02") {
		nextTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 0, 0, 0, 0, time.Local)
	} else if true == strings.Contains(l.filename_layout, "01") {
		nextTime = time.Date(nowTime.Year(), nowTime.Month()+1, 0, 0, 0, 0, 0, time.Local)
	} else if true == strings.Contains(l.filename_layout, "2006") {
		nextTime = time.Date(nowTime.Year()+1, nowTime.Month(), 0, 0, 0, 0, 0, time.Local)
	} else {
		nextTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour()+1, 0, 0, 0, time.Local)
	}

	return nextTime.Sub(nowTime)
}

func (l *Log) run() {
	for l.bRuning {
		select {
		case <-l.cfgTimer.C:
			if err := l.loadConfig(); err == nil {
				l.newFile()
				l.newFileTimer.Reset(l.newFileNextDuration())
			}
		case <-l.newFileTimer.C:
			l.newFileTimer.Reset(l.newFileNextDuration())
			l.newFile()
			go func() {
				l.delZips()
				l.zipLogs()
			}()
		case content := <-l.fileChan:
			l.write(content)
		case <-l.flushChan:
			l.flush()
		}
	}
}

/// 新增加 %j(json默认格式打印) %J(json格式打印所有字段)
func formatUpdate(format *string, params ...interface{}) {

	flen := len(*format)
	fsrc := []byte(*format)
	plen := len(params)
	pi := 0
	for fi := 0; fi < flen && pi < plen; fi++ {
		if fsrc[fi] != '%' {
			continue
		}
		fi++
		verb := fsrc[fi]
		switch verb {
		case 'J':
			{
				params[pi] = utils.JsonOmitMarshalString(params[pi])
				fsrc[fi] = 's'
			}
		case 'j':
			{
				params[pi] = utils.JsonMarshalString(params[pi])
				fsrc[fi] = 's'
			}
		case '%':
			fi++
			continue
		}
		pi++
		fi++
	}
	*format = string(fsrc)
}

func Format(callerSkip int, level LogLevel, format string, params ...interface{}) string {
	nowTime := time.Now()
	_, fileName, fileLine, _ := runtime.Caller(callerSkip)
	strTag := fmt.Sprintf("%s [%s] %s:%d | ",
		nowTime.Format("15:04:05.000"), LevelName(level), filepath.Base(fileName), fileLine)

	/// 预处理新增加的格式, 新增加 %j(json默认格式打印) %J(json格式打印所有字段)
	formatUpdate(&format, params...)

	/// 格式化
	strContent := fmt.Sprintf(format, params...)

	strLine := strTag + strContent + "\n"
	return strLine
}

func (l *Log) Format(callerSkip int, level LogLevel, format string, params ...interface{}) {
	if level&l.level == LogLevel(0) {
		return
	}

	content := Format(callerSkip+1, level, format, params...)
	l.fileChan <- content
}

func (l *Log) loadConfig() error {

	/// 检查配置文件是否修改过
	cfgInfo, err := os.Stat(l.logConfig)
	if err != nil {
		Errorf("loadConfig os.Stat fail, logConfig=[%s] err=[%v]", l.logConfig, err)
		return err
	}
	modTime := cfgInfo.ModTime().Unix()
	if modTime == l.cfgModTime {
		Debugf("loadConfig cfg not modify, modTime=[%d] l.cfgModTime=[%d]", modTime, l.cfgModTime)
		return nil
	}
	l.cfgModTime = modTime

	/// 配置文件json格式解析
	type Config struct {
		Log   LogConfig
	}
	var cfg *Config
	if err = utils.LoadYaml(l.logConfig, &cfg); err != nil {
		Errorf("loadConfig LoadYaml2Mapsi fail, logConfig=[%s] err=[%v]", l.logConfig, err)
		return err
	}

	logCfg := cfg.Log
	l.level = 0
	/// 配置文件字段解析
	if true == logCfg.WriteConsole {
		l.level |= E_LOG_WRITE_CONSOLE
	}
	if true == logCfg.WriteFile {
		l.level |= E_LOG_WRITE_FILE
	}
	if true == logCfg.OpenTrace {
		l.level |= E_LOG_LEVEL_TRACE
	}
	if true == logCfg.OpenDebug {
		l.level |= E_LOG_LEVEL_DEBUG
	}
	if true == logCfg.OpenInfo {
		l.level |= E_LOG_LEVEL_INFO
	}
	if true == logCfg.OpenWarn {
		l.level |= E_LOG_LEVEL_WARN
	}
	if true == logCfg.OpenError {
		l.level |= E_LOG_LEVEL_ERROR
	}

	///// 获取日志最大保留时长
	l.remainHours = logCfg.MaxRemainHours
	l.filename_layout = logCfg.FilenameLayout
	l.path = logCfg.Path

	Debugf("loadConfig success, logConfig=[%s]", l.logConfig)
	return nil
}

func (l *Log) newFile() error {
	if err := os.MkdirAll(l.path, 0777); err != nil {
		Errorf("newFile os.MkdirAll fail, path=[%s] err=[%v]", l.path, err)
		return err
	}

	/// 产生新文件
	filename := l.path + "/" + time.Now().Format(l.filename_layout) + ".log"
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
	var err error
	l.file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Errorf("newFile OpenFile fail, filename=[%s] err=[%v]", filename, err)
		return err
	}

	Infof("newFile success, filename=[%s]", filename)
	return nil
}

func (l *Log) write(content string) {

	if l.level&E_LOG_WRITE_FILE > 0 {
		l.file.WriteString(content)
	}

	if l.level&E_LOG_WRITE_CONSOLE > 0 {
		fmt.Print(content)
	}
}

func (l *Log) zipLogs() {

	nowTime := time.Now()
	filepath.Walk(l.path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		modTime := info.ModTime()
		deltaNs := nowTime.Sub(modTime)
		if deltaNs < time.Hour {
			return nil
		}

		extName := filepath.Ext(path)
		if extName != ".log" {
			return nil
		}

		zipName := path + ".zip"
		zipFile, err := os.OpenFile(zipName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			Errorf("zipLogs os.OpenFile fail, zipName=[%s] err=[%v]", zipName, err)
			return err
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		logName := filepath.Base(path)
		itemFile, err := zipWriter.Create(logName)
		if err != nil {
			Errorf("zipLogs zipWriter.Create fail, logName=[%s] err=[%v]", logName, err)
			return err
		}

		logFile, err := os.Open(path)
		if err != nil {
			Errorf("zipLogs os.Open fail, logName=[%s] err=[%v]", logName, err)
			return err
		}
		defer logFile.Close()

		_, err = io.Copy(itemFile, logFile)
		if err != nil {
			Errorf("zipLogs io.Copy fail, logName=[%s] err=[%v]", logName, err)
			return err
		}

		err = os.Remove(path)
		if err != nil {
			Errorf("zipLogs os.Remove fail, logName=[%s] err=[%v]", logName, err)
			return err
		}

		Debugf("zipLogs zip file, logName=[%s]", logName)
		return nil
	})
}

func (l *Log) delZips() {

	nowTime := time.Now()
	/// 删除旧压缩文件
	filepath.Walk(l.path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		deltaNs := nowTime.Sub(info.ModTime())
		if deltaNs < time.Duration(l.remainHours)*time.Hour {
			return nil
		}

		extName := filepath.Ext(path)
		if extName != ".zip" {
			return nil
		}

		err = os.Remove(path)
		if err != nil {
			Errorf("delZips os.Remove fail, path=[%s] err=[%v]", path, err)
			return err
		}

		Debugf("delZips os.Remove success, path=[%s]", path)
		return nil
	})
}

func (l *Log) Exit() {
	l.bRuning = false
}

func LevelName(level LogLevel) string {
	switch level {
	case E_LOG_LEVEL_TRACE:
		return "trace"
	case E_LOG_LEVEL_DEBUG:
		return "debug"
	case E_LOG_LEVEL_INFO:
		return "info"
	case E_LOG_LEVEL_WARN:
		return "warn"
	case E_LOG_LEVEL_ERROR:
		return "error"
	default:
		return ""
	}
}
