package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/Chairou/toolbox/util/color"
)

var nameLogMap sync.Map
var intLogMap sync.Map

var (
	DEBUG_LEVEL int = 0
	INFO_LEVEL  int = 1
	ERROR_LEVEL int = 2
)

// logNewMu 保证同名日志池只会被创建一次，避免并发时重复打开同一文件
var logNewMu sync.Mutex

type LogIntFileName struct {
	Lock     sync.RWMutex
	orderNum int
}

var logInit LogIntFileName

type LogPool struct {
	Name         string
	FileName     string
	Level        int
	Path         string
	PrintConsole bool
	num          int      // 在 intLogMap 中的编号，Close 时用于注销
	file         *Loggerj // 实际写入的文件，由 Close 释放
	infoLogger   *Logger
	debugLogger  *Logger
	errorLogger  *Logger
}

func init() {
	logInit.init()
}

// NewLogPool 创建日志池，同名实例已存在时直接返回
func NewLogPool(name string, fileName string) (*LogPool, error) {
	logNewMu.Lock()
	defer logNewMu.Unlock()

	if inst, ok := nameLogMap.Load(name); ok {
		return inst.(*LogPool), nil
	}

	inst := &LogPool{}
	inst.Path, _ = os.Getwd()
	inst.Name = name
	inst.FileName = fileName
	inst.Level = DEBUG_LEVEL
	inst.PrintConsole = false

	dir1 := filepath.Dir(fileName)
	// 检测目录是否存在，不存在则创建（使用 MkdirAll 支持多层目录）
	if _, err := os.Stat(dir1); os.IsNotExist(err) {
		if err := os.MkdirAll(dir1, 0755); err != nil {
			return nil, fmt.Errorf("logger mkdir failed, err: %w, dir: %s", err, dir1)
		}
	}

	var logFileName string
	if dir1 == "." {
		logFileName = inst.Path + "/" + inst.FileName
	} else {
		logFileName = fileName
	}

	// 探测文件是否可写，尽早暴露权限等问题；写入由 lumberjack 按需打开文件
	probe, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	if err := probe.Close(); err != nil {
		return nil, err
	}

	lumberjackLogger := &Loggerj{
		Filename:   logFileName,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     31,    //days
		Compress:   false, // disabled by default
	}
	inst.file = lumberjackLogger

	flags := log.Ldate | log.Ltime | log.Lshortfile
	inst.infoLogger = New(lumberjackLogger, "INFO: ", flags)
	inst.debugLogger = New(lumberjackLogger, "DEBUG: ", flags)
	inst.errorLogger = New(lumberjackLogger, "ERROR: ", flags)

	nameLogMap.Store(name, inst)
	logInit.SaveIntLogMap(inst)

	return inst, nil
}

func GetLog() *LogPool {
	inst := GetLogNum(1)
	if inst == nil {
		fmt.Println("GetLogNum| get logger from logInit failed")
		return nil
	}
	return inst
}

func GetLogName(name string) *LogPool {
	inst, ok := nameLogMap.Load(name)
	if ok {
		return inst.(*LogPool)
	}
	fmt.Println("GetLogName| get logger from nameLogMap failed")
	return nil
}

func GetLogNum(logNumber int) *LogPool {
	inst, ok := intLogMap.Load(logNumber)
	if ok {
		return inst.(*LogPool)
	}
	fmt.Println("GetLogNum| get logger from logInit failed")
	return nil
}

func (c *LogIntFileName) init() {
	c.Lock.Lock()
	c.orderNum = 1
	c.Lock.Unlock()
}

func (c *LogIntFileName) SaveIntLogMap(inst *LogPool) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	intLogMap.Store(c.orderNum, inst)
	inst.num = c.orderNum
	c.orderNum++
}

func (c *LogPool) SetPrintConsole(isEnable bool) {
	c.PrintConsole = isEnable
}

// emit 写入日志文件，并在开启控制台输出时打印 console 内容。
// 调用链为 用户代码 -> 公开方法 -> emit -> Output，因此 Output 的调用深度为 3，
// 保证 Lshortfile 报告的是用户代码的文件名和行号。
func (c *LogPool) emit(l *Logger, s string, console string) {
	_ = l.Output(3, s)
	if c.PrintConsole {
		log.Println(console)
	}
}

// logEnabled 判断 level 是否达到当前日志池的输出级别
func (c *LogPool) logEnabled(level int) bool {
	return c.Level <= level
}

func (c *LogPool) DebugfTag(tag string, format string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := tag + " " + fmt.Sprintf(format, v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPool) Debugf(format string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := fmt.Sprintf(format, v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPool) DebugTag(tag string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := tag + " " + fmt.Sprintln(v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPool) Debug(v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := fmt.Sprintln(v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPool) InfofTag(tag string, format string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := tag + " " + fmt.Sprintf(format, v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPool) Infof(format string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := fmt.Sprintf(format, v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPool) InfoTag(tag string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := tag + " " + fmt.Sprintln(v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPool) Info(v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := fmt.Sprintln(v...)
		c.emit(c.infoLogger, s, s)
	}
}

// 错误日志始终输出，不受 Level 限制；控制台输出使用红色
func (c *LogPool) ErrorfTag(tag string, format string, v ...any) {
	s := tag + " " + fmt.Sprintf(format, v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPool) Errorf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPool) ErrorTag(tag string, v ...any) {
	s := tag + " " + fmt.Sprintln(v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPool) Error(v ...any) {
	s := fmt.Sprintln(v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPool) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.Level = level
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}

// Close 从注册表中注销日志池并关闭日志文件，释放文件句柄。
// 注销后同名的 NewLogPool 会创建新实例；关闭后若再次写入，lumberjack 会重新打开文件。
func (c *LogPool) Close() error {
	logNewMu.Lock()
	defer logNewMu.Unlock()

	// 只删除仍指向自身的注册项，避免误删之后用同名创建的新实例
	nameLogMap.CompareAndDelete(c.Name, c)
	intLogMap.CompareAndDelete(c.num, c)

	if c.file == nil {
		return nil
	}
	return c.file.Close()
}
