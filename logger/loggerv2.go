package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Chairou/toolbox/util/color"
)

var nameLogMap2 sync.Map
var intLogMap2 sync.Map

// newLogMu 保证同名日志池只会被创建一次，避免并发时重复打开同一文件
var newLogMu sync.Mutex

// LogIntFileNameV2 日志文件编号管理器，用于为日志池分配递增编号
type LogIntFileNameV2 struct {
	Lock     sync.RWMutex
	orderNum int
}

var logInitV2 LogIntFileNameV2

// LogPoolV2 日志池，封装了日志文件句柄和多级别日志记录器
type LogPoolV2 struct {
	Name         string
	LogFileName  string
	LogLevel     int      // 日志级别，可选 DEBUG_LEVEL、INFO_LEVEL、ERROR_LEVEL
	PrintConsole int      // 是否同时输出到控制台
	fd           *Loggerj // 实际写入的文件，由 Close 释放
	num          int      // 在 intLogMap2 中的编号，Close 时用于注销
	infoLogger   *Logger
	debugLogger  *Logger
	errorLogger  *Logger
}

// LogOpt 日志配置选项
type LogOpt struct {
	LogFileName  string `config:"logFileName"`  // 日志文件名或路径
	LogLevel     int    `config:"logLevel"`     // 日志级别，可选 DEBUG_LEVEL、INFO_LEVEL、ERROR_LEVEL
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
	if opt.LogFileName == "" {
		return nil, errors.New("logger file name or path is empty")
	}
	newLogMu.Lock()
	defer newLogMu.Unlock()

	if inst, ok := nameLogMap2.Load(name); ok {
		return inst.(*LogPoolV2), nil
	}
	pathFileName, err := safePath("./", opt.LogFileName)
	if err != nil {
		return nil, fmt.Errorf("logger fileName failed, err: %w", err)
	}

	level := opt.LogLevel
	if level < DEBUG_LEVEL || level > ERROR_LEVEL {
		level = DEBUG_LEVEL
	}
	inst := &LogPoolV2{
		Name:         name,
		LogFileName:  pathFileName,
		LogLevel:     level,
		PrintConsole: opt.PrintConsole,
	}

	dir2 := filepath.Dir(pathFileName)
	// 检测目录是否存在，不存在则创建（使用 MkdirAll 支持多层目录）
	if _, err := os.Stat(dir2); os.IsNotExist(err) {
		if err := os.MkdirAll(dir2, 0755); err != nil {
			return nil, fmt.Errorf("logger mkdir failed, err: %w, dir: %s", err, dir2)
		}
	}

	// 探测文件是否可写，尽早暴露权限等问题；写入由 lumberjack 按需打开文件
	probe, err := os.OpenFile(pathFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return nil, err
	}
	if err := probe.Close(); err != nil {
		return nil, err
	}

	lumberjackLogger := &Loggerj{
		Filename:   pathFileName,
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
	inst.fd = lumberjackLogger

	flags := log.Ldate | log.Ltime | log.Lshortfile
	inst.infoLogger = New(lumberjackLogger, "INFO: ", flags)
	inst.debugLogger = New(lumberjackLogger, "DEBUG: ", flags)
	inst.errorLogger = New(lumberjackLogger, "ERROR: ", flags)

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
	inst.num = c.orderNum
	c.orderNum++
}

func (c *LogPoolV2) SetPrintConsole(isEnable int) {
	c.PrintConsole = isEnable
}

// emit 写入日志文件，并在开启控制台输出时打印 console 内容。
// 调用链为 用户代码 -> 公开方法 -> emit -> Output，因此 Output 的调用深度为 3，
// 保证 Lshortfile 报告的是用户代码的文件名和行号。
func (c *LogPoolV2) emit(l *Logger, s string, console string) {
	_ = l.Output(3, s)
	if c.PrintConsole == 1 {
		log.Println(console)
	}
}

// logEnabled 判断 level 是否达到当前日志池的输出级别
func (c *LogPoolV2) logEnabled(level int) bool {
	return c.LogLevel <= level
}

func (c *LogPoolV2) DebugfTag(tag string, format string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := tag + " " + fmt.Sprintf(format, v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPoolV2) Debugf(format string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := fmt.Sprintf(format, v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPoolV2) DebugTag(tag string, v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := tag + " " + fmt.Sprintln(v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPoolV2) Debug(v ...any) {
	if c.logEnabled(DEBUG_LEVEL) {
		s := fmt.Sprintln(v...)
		c.emit(c.debugLogger, s, s)
	}
}

func (c *LogPoolV2) InfofTag(tag string, format string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := tag + " " + fmt.Sprintf(format, v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPoolV2) Infof(format string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := fmt.Sprintf(format, v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPoolV2) InfoTag(tag string, v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := tag + " " + fmt.Sprintln(v...)
		c.emit(c.infoLogger, s, s)
	}
}

func (c *LogPoolV2) Info(v ...any) {
	if c.logEnabled(INFO_LEVEL) {
		s := fmt.Sprintln(v...)
		c.emit(c.infoLogger, s, s)
	}
}

// ErrorfTag 错误日志始终输出，不受 LogLevel 限制；控制台输出使用红色
func (c *LogPoolV2) ErrorfTag(tag string, format string, v ...any) {
	s := tag + " " + fmt.Sprintf(format, v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPoolV2) Errorf(format string, v ...any) {
	s := fmt.Sprintf(format, v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPoolV2) ErrorTag(tag string, v ...any) {
	s := tag + " " + fmt.Sprintln(v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPoolV2) Error(v ...any) {
	s := fmt.Sprintln(v...)
	c.emit(c.errorLogger, s, color.SetColor(color.Red, s))
}

func (c *LogPoolV2) SetLevel(level int) error {
	if level >= DEBUG_LEVEL && level <= ERROR_LEVEL {
		c.LogLevel = level
		return nil
	}
	return errors.New("level must be DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL")
}

// poolWriter 写入日志文件，并根据日志池当前的 PrintConsole 决定是否同时输出到控制台
type poolWriter struct {
	pool *LogPoolV2
}

func (w poolWriter) Write(p []byte) (int, error) {
	n, err := w.pool.fd.Write(p)
	if w.pool.PrintConsole == 1 {
		_, _ = os.Stderr.Write(p)
	}
	return n, err
}

// Writer 暴露 io.Writer，供其他日志模块（如 log、gin、第三方 logger）写入同一个日志文件。
// 该 Writer 并发安全，并与本日志池共用同一套转存策略（MaxSizeMB/MaxBackups/MaxAgeDay/Compress），
// 写入超过大小限制时会自动转存。
// 是否同时打印到屏幕（stderr）由日志池的 PrintConsole 实时决定，可随时通过 SetPrintConsole 切换。
// 调用方不应关闭该 Writer，请通过 LogPoolV2.Close 统一释放。
func (c *LogPoolV2) Writer() io.Writer {
	return poolWriter{pool: c}
}

// Rotate 手动触发日志转存：关闭当前文件并重命名为带时间戳的备份，之后的写入进入新文件。
func (c *LogPoolV2) Rotate() error {
	if c.fd == nil {
		return errors.New("logger file is not initialized")
	}
	return c.fd.Rotate()
}

// Close 从注册表中注销日志池并关闭日志文件，释放文件句柄。
// 注销后同名的 NewLogOpt 会创建新实例；已持有的旧实例若再次写入，
// lumberjack 会重新打开文件，但它不再能通过 GetLogNameV2/GetLogNumV2 获取。
func (c *LogPoolV2) Close() error {
	newLogMu.Lock()
	defer newLogMu.Unlock()

	// 只删除仍指向自身的注册项，避免误删之后用同名创建的新实例
	nameLogMap2.CompareAndDelete(c.Name, c)
	intLogMap2.CompareAndDelete(c.num, c)

	if c.fd == nil {
		return nil
	}
	return c.fd.Close()
}

// safePath 将用户路径限制在基准目录内，返回解析符号链接后的绝对路径
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
	// 只拒绝真正的上级目录（".." 或 "../..."），不误伤 "..app.log" 这类文件名
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes base directory: %s", userPath)
	}

	return resolved, nil
}
