package logger

import (
	"errors"
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

type logPool struct {
	Fd          *os.File
	FileName    string
	Level       int
	Path        string
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
}

func NewLogPool(fileName string) (*logPool, error) {
	inst, ok := logMap.Load(fileName)
	if ok {
		return inst.(*logPool), nil
	} else {
		fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return nil, err
		}

		infoLog := log.New(fd, "INFO: ", log.Ldate|log.Ltime|log.Llongfile)
		debugLog := log.New(fd, "DEBUG: ", log.Ldate|log.Ltime|log.Llongfile)
		errorLog := log.New(fd, "ERROR: ", log.Ldate|log.Ltime|log.Llongfile)

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

func (c *logPool) Debugf(format string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		c.debugLogger.Printf(format, v...)
	}
}

func (c *logPool) Debugln(v ...any) {
	if c.Level <= DEBUG_LEVEL {
		c.debugLogger.Println(v...)
	}
}

func (c *logPool) Infof(format string, v ...any) {
	if c.Level <= INFO_LEVEL {
		c.infoLogger.Printf(format, v...)
	}
}

func (c *logPool) Infoln(v ...any) {
	if c.Level <= INFO_LEVEL {
		c.infoLogger.Println(v...)
	}
}

func (c *logPool) Errorf(format string, v ...any) {
	c.errorLogger.Printf(format, v...)
}

func (c *logPool) Errorln(v ...any) {
	c.errorLogger.Println(v...)
}

func (c *logPool) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.Level = level
		logMap.Store(c.FileName, c)
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}
