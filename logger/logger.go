package logger

import (
	"errors"
	"fmt"
	"github.com/natefinch/lumberjack"
	"log"
	"os"
	"sync"
)

var logMap sync.Map

var (
	DEBUG_LEVEL int = 0
	INFO_LEVEL  int = 1
	ERROR_LEVEL int = 2
)

type LogIntFileName struct {
	Lock        sync.RWMutex
	orderNum    int
	IntFileName map[int]string
}

var logIntFileName LogIntFileName

type logPool struct {
	Fd          *os.File
	FileName    string
	Level       int
	Path        string
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
}

func init() {
	logIntFileName.Lock.Lock()
	defer logIntFileName.Lock.Unlock()
	logIntFileName.orderNum = 1
	logIntFileName.IntFileName = make(map[int]string)
}

func NewLogPool(fileName string) (*logPool, error) {
	inst, ok := logMap.Load(fileName)
	if ok {
		return inst.(*logPool), nil
	} else {
		SaveLogNameToInt(fileName)
		fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return nil, err
		}

		infoLog := log.New(fd, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		debugLog := log.New(fd, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		errorLog := log.New(fd, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

		inst := &logPool{}
		inst.Path, _ = os.Getwd()
		inst.Fd = fd
		inst.FileName = fileName
		inst.Level = DEBUG_LEVEL
		inst.Path, _ = os.Getwd()
		inst.infoLogger = infoLog
		inst.debugLogger = debugLog
		inst.errorLogger = errorLog

		lumberjackLogger := &lumberjack.Logger{
			Filename:   inst.Path + "/" + inst.FileName,
			MaxSize:    500, // megabytes
			MaxBackups: 10,
			MaxAge:     31,    //days
			Compress:   false, // disabled by default
		}
		// 设置日志分割
		debugLog.SetOutput(lumberjackLogger)
		infoLog.SetOutput(lumberjackLogger)
		errorLog.SetOutput(lumberjackLogger)

		logMap.Store(fileName, inst)
		return inst, nil
	}
}

func GetLogPool(fileName string) (*logPool, error) {
	inst, ok := logMap.Load(fileName)
	if ok {
		return inst.(*logPool), nil
	} else {
		return nil, errors.New("get logger from logMap failed")
	}
}

func SaveLogNameToInt(fileName string) {
	logIntFileName.Lock.Lock()
	logIntFileName.IntFileName[logIntFileName.orderNum] = fileName
	logIntFileName.orderNum += 1
	logIntFileName.Lock.Unlock()
}

func GetLogNum(logNumber int) (*logPool, error) {
	logIntFileName.Lock.RLock()
	fileName, ok := logIntFileName.IntFileName[logNumber]
	logIntFileName.Lock.RUnlock()
	if ok {
		return GetLogPool(fileName)
	} else {
		return nil, errors.New("GetLogNum| get logger from logIntFileName failed")
	}
}

func (c *logPool) Debugf(format string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintf(format, v...)
		c.debugLogger.Output(2, s)
	}
}

func (c *logPool) Debugln(v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintln(v...)
		c.debugLogger.Output(2, s)
	}
}

func (c *logPool) Infof(format string, v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintf(format, v...)
		c.infoLogger.Output(2, s)
	}
}

func (c *logPool) Infoln(v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintln(v...)
		c.infoLogger.Output(2, s)
	}
}

func (c *logPool) Errorf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	c.errorLogger.Output(2, s)
}

func (c *logPool) Errorln(v ...any) {
	s := fmt.Sprintln(v...)
	c.errorLogger.Output(2, s)
}

func (c *logPool) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.Level = level
		logMap.Store(c.FileName, c)
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}
