package logger

import (
	"errors"
	"fmt"
	"github.com/Chairou/toolbox/util/color"
	"github.com/natefinch/lumberjack"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var nameLogMap sync.Map
var intLogMap sync.Map

var (
	DEBUG_LEVEL int = 0
	INFO_LEVEL  int = 1
	ERROR_LEVEL int = 2
)

type LogIntFileName struct {
	Lock     sync.RWMutex
	orderNum int
}

var logIntFileName LogIntFileName

type LogPool struct {
	Fd          *os.File
	FileName    string
	Level       int
	Path        string
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
}

func init() {
	logIntFileName.init()
}

func NewLogPool(name string, fileName string) (*LogPool, error) {
	inst, ok := nameLogMap.Load(name)
	if ok {
		return inst.(*LogPool), nil
	} else {
		fd, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			return nil, err
		}

		infoLog := log.New(fd, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		debugLog := log.New(fd, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		errorLog := log.New(fd, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

		inst := &LogPool{}
		inst.Path, _ = os.Getwd()
		inst.Fd = fd
		inst.FileName = fileName
		inst.Level = DEBUG_LEVEL
		inst.Path, _ = os.Getwd()
		inst.infoLogger = infoLog
		inst.debugLogger = debugLog
		inst.errorLogger = errorLog

		dir1 := filepath.Dir(fileName)
		var logFileName string
		if dir1 == "." {
			logFileName = inst.Path + "/" + inst.FileName
		} else {
			logFileName = fileName
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   logFileName,
			MaxSize:    500, // megabytes
			MaxBackups: 10,
			MaxAge:     31,    //days
			Compress:   false, // disabled by default
		}
		// 设置日志分割
		debugLog.SetOutput(lumberjackLogger)
		infoLog.SetOutput(lumberjackLogger)
		errorLog.SetOutput(lumberjackLogger)

		nameLogMap.Store(name, inst)
		logIntFileName.SaveIntLogMap(inst)

		return inst, nil
	}
}

func GetLog() *LogPool {
	inst := GetLogNum(1)
	if inst == nil {
		_ = fmt.Errorf("GetLogNum| get logger from logIntFileName failed")
		return nil
	}
	return inst
}

func GetLogName(name string) *LogPool {
	logIntFileName.Lock.RLock()
	inst, ok := nameLogMap.Load(name)
	logIntFileName.Lock.RUnlock()
	if ok {
		return inst.(*LogPool)
	} else {
		fmt.Println("GetLogName| get logger from nameLogMap failed")
		return nil
	}
}

func GetLogNum(logNumber int) *LogPool {
	logIntFileName.Lock.RLock()
	inst, ok := intLogMap.Load(logNumber)
	logIntFileName.Lock.RUnlock()
	if ok {
		return inst.(*LogPool)
	} else {
		fmt.Println("GetLogNum| get logger from logIntFileName failed")
		return nil
	}
}

func (c *LogIntFileName) init() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.orderNum = 1
}

func (c *LogIntFileName) SaveIntLogMap(inst *LogPool) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	intLogMap.Store(c.orderNum, inst)
	c.orderNum += 1
}

func (c *LogPool) Debugf(format string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintf(format, v...)
		log.Println(s)
		c.debugLogger.Output(2, s)
	}
}

func (c *LogPool) Debugln(v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintln(v...)
		log.Println(s)
		c.debugLogger.Output(2, s)
	}
}

func (c *LogPool) Infof(format string, v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintf(format, v...)
		log.Println(s)
		c.infoLogger.Output(2, s)
	}
}

func (c *LogPool) Infoln(v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintln(v...)
		log.Println(s)
		c.infoLogger.Output(2, s)
	}
}

func (c *LogPool) Errorf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	color.SetColor(color.Red, s)
	log.Println(s)
	c.errorLogger.Output(2, s)
}

func (c *LogPool) Errorln(v ...any) {
	s := fmt.Sprintln(v...)
	color.SetColor(color.Red, s)
	log.Println(s)
	c.errorLogger.Output(2, s)
}

func (c *LogPool) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.Level = level
		nameLogMap.Store(c.FileName, c)
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}
