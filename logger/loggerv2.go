package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Chairou/toolbox/util/color"
)

var nameLogMap2 sync.Map
var intLogMap2 sync.Map

// LogIntFileNameV2 日志文件编号管理器，用于为日志池分配递增编号
type LogIntFileNameV2 struct {
	Lock     sync.RWMutex
	orderNum int
}

var logInitV2 LogIntFileNameV2

// LogPoolV2 日志池，封装了日志文件描述符和多级别日志记录器
type LogPoolV2 struct {
	Fd           *os.File
	Name         string
	FileName     string
	Level        int
	PrintConsole int
	infoLogger   *Logger
	debugLogger  *Logger
	errorLogger  *Logger
}

// LogOpt 日志配置选项
type LogOpt struct {
	FileName     string `config:"fileName"`     // 日志文件名或路径
	Level        int    `config:"level"`        // 日志级别，可选 DEBUG_LEVEL、INFO_LEVEL、ERROR_LEVEL
	MaxSizeMB    int    `config:"maxSizeMB"`    // 单个日志文件最大大小（MB）
	MaxBackups   int    `config:"maxBackups"`   // 最大保留的旧日志文件数量
	MaxAgeDay    int    `config:"maxAgeDay"`    // 旧日志文件最大保留天数
	Compress     int    `config:"compress"`     // 是否压缩旧日志文件，默认不压缩
	PrintConsole int    `config:"printConsole"` // 是否同时输出到控制台
}

func init() {
	logInitV2.init()
}

// NewLogOpt 根据配置选项创建日志池实例，如果同名实例已存在则直接返回
func NewLogOpt(name string, opt *LogOpt) (*LogPoolV2, error) {
	if opt.FileName == "" {
		return nil, errors.New("logger file name or path is empty")
	}
	if inst, ok := nameLogMap2.Load(name); ok {
		return inst.(*LogPoolV2), nil
	}
	pathFileName, err := safePath("./", opt.FileName)
	if err != nil {
		return nil, fmt.Errorf("logger fileName failed, err: %w", err)
	}

	level := opt.Level
	if level < DEBUG_LEVEL || level > ERROR_LEVEL {
		level = DEBUG_LEVEL
	}
	inst := &LogPoolV2{
		Name:         name,
		FileName:     pathFileName,
		Level:        level,
		PrintConsole: opt.PrintConsole,
	}

	dir2 := filepath.Dir(pathFileName)
	// 检测目录是否存在，不存在则创建（使用 MkdirAll 支持多层目录）
	if _, err := os.Stat(dir2); os.IsNotExist(err) {
		if err := os.MkdirAll(dir2, 0755); err != nil {
			return nil, fmt.Errorf("logger mkdir failed, err: %w, dir: %s", err, dir2)
		}
	}
	var logFileName = pathFileName

	fd, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	infoLog := New(fd, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLog := New(fd, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := New(fd, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	inst.Fd = fd
	inst.infoLogger = infoLog
	inst.debugLogger = debugLog
	inst.errorLogger = errorLog

	lumberjackLogger := &Loggerj{
		Filename:   logFileName,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     31,    //days
		Compress:   false, // disabled by default
	}
	if opt.MaxSizeMB != 0 {
		lumberjackLogger.MaxSize = opt.MaxSizeMB
	}
	if opt.MaxBackups != 0 {
		lumberjackLogger.MaxBackups = opt.MaxBackups
	}
	if opt.MaxAgeDay != 0 {
		lumberjackLogger.MaxAge = opt.MaxAgeDay
	}
	if opt.Compress == 1 {
		lumberjackLogger.Compress = true
	}

	// 设置日志分割
	debugLog.SetOutput(lumberjackLogger)
	infoLog.SetOutput(lumberjackLogger)
	errorLog.SetOutput(lumberjackLogger)

	nameLogMap2.Store(name, inst)
	logInitV2.SaveIntLogMap(inst)

	return inst, nil
}

// GetLogV2 获取第一个创建的 v2 日志池实例
func GetLogV2() *LogPoolV2 {
	inst := GetLogNumV2(1)
	if inst == nil {
		fmt.Println("GetLogNum| get logger from logInit failed")
		return nil
	}
	return inst
}

// GetLogNameV2 根据名称获取 v2 日志池实例
func GetLogNameV2(name string) *LogPoolV2 {
	inst, ok := nameLogMap2.Load(name)
	if ok {
		return inst.(*LogPoolV2)
	}
	fmt.Println("GetLogName| get logger from nameLogMap failed")
	return nil
}

// GetLogNumV2 根据编号获取 v2 日志池实例
func GetLogNumV2(logNumber int) *LogPoolV2 {
	inst, ok := intLogMap2.Load(logNumber)
	if ok {
		return inst.(*LogPoolV2)
	}
	fmt.Println("GetLogNum| get logger from logInit failed")
	return nil
}

func (c *LogIntFileNameV2) init() {
	c.Lock.Lock()
	c.orderNum = 1
	c.Lock.Unlock()
}

func (c *LogIntFileNameV2) SaveIntLogMap(inst *LogPoolV2) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	intLogMap2.Store(c.orderNum, inst)
	c.orderNum++
}

func (c *LogPoolV2) SetPrintConsole(isEnable int) {
	c.PrintConsole = isEnable
}

func (c *LogPoolV2) DebugfTag(tag string, format string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := tag + " " + fmt.Sprintf(format, v...)
		_ = c.debugLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}
func (c *LogPoolV2) Debugf(format string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintf(format, v...)
		_ = c.debugLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}

func (c *LogPoolV2) DebugTag(tag string, v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := tag + " " + fmt.Sprintln(v...)
		_ = c.debugLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}
func (c *LogPoolV2) Debug(v ...any) {
	if c.Level <= DEBUG_LEVEL {
		s := fmt.Sprintln(v...)
		_ = c.debugLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}

func (c *LogPoolV2) InfofTag(tag string, format string, v ...any) {
	if c.Level <= INFO_LEVEL {
		s := tag + " " + fmt.Sprintf(format, v...)
		_ = c.infoLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}
func (c *LogPoolV2) Infof(format string, v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintf(format, v...)
		_ = c.infoLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}

func (c *LogPoolV2) InfoTag(tag string, v ...any) {
	if c.Level <= INFO_LEVEL {
		s := tag + " " + fmt.Sprintln(v...)
		_ = c.infoLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}

func (c *LogPoolV2) Info(v ...any) {
	if c.Level <= INFO_LEVEL {
		s := fmt.Sprintln(v...)
		_ = c.infoLogger.Output(3, s)
		if c.PrintConsole == 1 {
			log.Println(s)
		}
	}
}

func (c *LogPoolV2) ErrorfTag(tag string, format string, v ...any) {
	s := tag + " " + fmt.Sprintf(format, v...)
	coloredStr := color.SetColor(color.Red, s)
	_ = c.errorLogger.Output(3, s)
	if c.PrintConsole == 1 {
		log.Println(coloredStr)
	}
}

func (c *LogPoolV2) Errorf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	coloredStr := color.SetColor(color.Red, s)
	_ = c.errorLogger.Output(3, s)
	if c.PrintConsole == 1 {
		log.Println(coloredStr)
	}
}

func (c *LogPoolV2) ErrorTag(tag string, v ...any) {
	s := tag + " " + fmt.Sprintln(v...)
	coloredStr := color.SetColor(color.Red, s)
	_ = c.errorLogger.Output(3, s)
	if c.PrintConsole == 1 {
		log.Println(coloredStr)
	}
}

func (c *LogPoolV2) Error(v ...any) {
	s := fmt.Sprintln(v...)
	coloredStr := color.SetColor(color.Red, s)
	_ = c.errorLogger.Output(3, s)
	if c.PrintConsole == 1 {
		log.Println(coloredStr)
	}
}

func (c *LogPoolV2) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.Level = level
		nameLogMap2.Store(c.Name, c)
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}

// Close 关闭日志池，释放文件描述符
func (c *LogPoolV2) Close() error {
	if c.Fd != nil {
		return c.Fd.Close()
	}
	return nil
}

// 检查日志文件路径是否安全
func safePath(basePath, userPath string) (string, error) {
	// 清理路径中的 ../ ./ 等
	cleaned := filepath.Clean(userPath)

	// 如果是相对路径，拼接到基准目录下
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(basePath, cleaned)
	}

	// 将路径转换为绝对路径
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// 尝试解析符号链接，获取真实路径
	resolved, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		// 目录可能还不存在，直接使用绝对路径
		resolved = absPath
	} else {
		resolved = filepath.Join(resolved, filepath.Base(absPath))
	}

	// 确保最终路径在基准目录下
	baseAbs, _ := filepath.Abs(basePath)
	// 解析基准目录的符号链接，保持与 resolved 路径一致
	if realBase, err := filepath.EvalSymlinks(baseAbs); err == nil {
		baseAbs = realBase
	}
	rel, err := filepath.Rel(baseAbs, resolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes base directory: %s", userPath)
	}

	return resolved, nil
}
