package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================
// 测试 NewLogPool：基本创建和重复创建返回同一实例
// ============================================================

func TestNewLogPool_Basic(t *testing.T) {
	// 清理测试文件
	defer os.Remove("test_basic.log")

	lp, err := NewLogPool("basic", "test_basic.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	if lp.Name != "basic" {
		t.Errorf("期望 Name='basic', 实际='%s'", lp.Name)
	}
	if lp.FileName != "test_basic.log" {
		t.Errorf("期望 FileName='test_basic.log', 实际='%s'", lp.FileName)
	}
	if lp.Level != DEBUG_LEVEL {
		t.Errorf("期望 Level=DEBUG_LEVEL(%d), 实际=%d", DEBUG_LEVEL, lp.Level)
	}
	if lp.PrintConsole != false {
		t.Errorf("期望 PrintConsole=false, 实际=%v", lp.PrintConsole)
	}
	if lp.Fd == nil {
		t.Error("期望 Fd 不为 nil")
	}
	if lp.infoLogger == nil || lp.debugLogger == nil || lp.errorLogger == nil {
		t.Error("期望所有 logger 实例不为 nil")
	}
}

func TestNewLogPool_DuplicateName(t *testing.T) {
	defer os.Remove("test_dup.log")

	lp1, err := NewLogPool("dup_test", "test_dup.log")
	if err != nil {
		t.Fatalf("第一次 NewLogPool 创建失败: %v", err)
	}
	defer lp1.Close()

	// 重复创建同名日志池，应返回同一实例
	lp2, err := NewLogPool("dup_test", "test_dup.log")
	if err != nil {
		t.Fatalf("第二次 NewLogPool 创建失败: %v", err)
	}

	if lp1 != lp2 {
		t.Error("重复创建同名日志池应返回同一实例")
	}
}

// ============================================================
// 测试 NewLogPool：MkdirAll 支持多层目录创建（修复 #3）
// ============================================================

func TestNewLogPool_MkdirAll(t *testing.T) {
	nestedDir := "./test_nested_dir/sub1/sub2"
	logFile := nestedDir + "/test_mkdir.log"
	defer os.RemoveAll("./test_nested_dir")

	lp, err := NewLogPool("mkdir_test", logFile)
	if err != nil {
		t.Fatalf("NewLogPool 创建多层目录失败: %v", err)
	}
	defer lp.Close()

	// 验证多层目录已被创建
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Errorf("多层目录 '%s' 应该已被创建", nestedDir)
	}
}

// ============================================================
// 测试 NewLogPool：目录创建失败时返回 error 而非 os.Exit（修复 #1）
// ============================================================

func TestNewLogPool_MkdirError(t *testing.T) {
	// 使用无效路径触发目录创建失败（在只读路径下创建目录）
	// 在 /dev/null 下创建目录应该失败
	invalidPath := "/dev/null/impossible/path/test.log"
	_, err := NewLogPool("mkdir_err_test", invalidPath)
	if err == nil {
		t.Error("在无效路径下创建日志池应返回 error，而非 nil")
	}
}

// ============================================================
// 测试 NewLogPool：logFileName 路径一致性（修复 #6）
// ============================================================

func TestNewLogPool_LogFileNameConsistency(t *testing.T) {
	// 当 fileName 不含目录前缀时（dir == "."），应使用绝对路径
	defer os.Remove("test_path.log")

	lp, err := NewLogPool("path_test", "test_path.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	cwd, _ := os.Getwd()
	expectedPath := cwd + "/test_path.log"

	// 验证文件确实在当前工作目录下被创建
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("日志文件应在 '%s' 被创建", expectedPath)
	}
}

func TestNewLogPool_LogFileNameWithDir(t *testing.T) {
	// 当 fileName 包含目录前缀时，应直接使用 fileName
	dir := "./test_logdir"
	logFile := dir + "/test_withdir.log"
	defer os.RemoveAll(dir)

	lp, err := NewLogPool("withdir_test", logFile)
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("日志文件应在 '%s' 被创建", logFile)
	}
}

// ============================================================
// 测试 GetLogName：移除了多余的 sync.Map 外层锁（修复 #5）
// ============================================================

func TestGetLogName_Found(t *testing.T) {
	defer os.Remove("test_getname.log")

	lp, err := NewLogPool("getname_test", "test_getname.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	result := GetLogName("getname_test")
	if result == nil {
		t.Fatal("GetLogName 应返回已创建的日志池实例")
	}
	if result != lp {
		t.Error("GetLogName 返回的实例应与创建时的实例相同")
	}
	if result.Name != "getname_test" {
		t.Errorf("期望 Name='getname_test', 实际='%s'", result.Name)
	}
}

func TestGetLogName_NotFound(t *testing.T) {
	result := GetLogName("nonexistent_name_xyz")
	if result != nil {
		t.Error("GetLogName 对不存在的名称应返回 nil")
	}
}

// ============================================================
// 测试 GetLogNum：移除了多余的 sync.Map 外层锁（修复 #5）
// ============================================================

func TestGetLogNum_Found(t *testing.T) {
	defer os.Remove("test_getnum.log")

	lp, err := NewLogPool("getnum_test", "test_getnum.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// logInit.orderNum 从 1 开始递增，获取最新的编号
	// 由于其他测试可能已经创建了日志池，我们通过遍历查找
	found := false
	for i := 1; i <= 100; i++ {
		result := GetLogNum(i)
		if result == lp {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetLogNum 应能通过编号找到已创建的日志池实例")
	}
}

func TestGetLogNum_NotFound(t *testing.T) {
	result := GetLogNum(99999)
	if result != nil {
		t.Error("GetLogNum 对不存在的编号应返回 nil")
	}
}

// ============================================================
// 测试 GetLog
// ============================================================

func TestGetLog(t *testing.T) {
	defer os.Remove("test_getlog.log")

	_, err := NewLogPool("getlog_test", "test_getlog.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}

	result := GetLog()
	if result == nil {
		t.Error("GetLog 应返回第一个创建的日志池实例")
	}
}

// ============================================================
// 测试 SetLevel：修复了 nameLogMap key 不一致问题（修复 #7）
// ============================================================

func TestSetLevel_Valid(t *testing.T) {
	defer os.Remove("test_setlevel.log")

	lp, err := NewLogPool("setlevel_test", "test_setlevel.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 测试设置各个有效级别
	for _, level := range []int{DEBUG_LEVEL, INFO_LEVEL, ERROR_LEVEL} {
		err := lp.SetLevel(level)
		if err != nil {
			t.Errorf("SetLevel(%d) 不应返回错误: %v", level, err)
		}
		if lp.Level != level {
			t.Errorf("SetLevel(%d) 后 Level 应为 %d, 实际=%d", level, level, lp.Level)
		}
	}
}

func TestSetLevel_Invalid(t *testing.T) {
	defer os.Remove("test_setlevel_inv.log")

	lp, err := NewLogPool("setlevel_inv_test", "test_setlevel_inv.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 测试无效级别
	err = lp.SetLevel(-1)
	if err == nil {
		t.Error("SetLevel(-1) 应返回错误")
	}

	err = lp.SetLevel(3)
	if err == nil {
		t.Error("SetLevel(3) 应返回错误")
	}
}

func TestSetLevel_NameLogMapKeyConsistency(t *testing.T) {
	defer os.Remove("test_setlevel_key.log")

	lp, err := NewLogPool("key_test", "test_setlevel_key.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 设置级别后，应能通过 Name 而非 FileName 获取到实例
	err = lp.SetLevel(ERROR_LEVEL)
	if err != nil {
		t.Fatalf("SetLevel 失败: %v", err)
	}

	// 通过 Name 获取应成功
	result := GetLogName("key_test")
	if result == nil {
		t.Fatal("SetLevel 后通过 Name 获取日志池应成功")
	}
	if result.Level != ERROR_LEVEL {
		t.Errorf("期望 Level=ERROR_LEVEL(%d), 实际=%d", ERROR_LEVEL, result.Level)
	}
}

// ============================================================
// 测试 Error 系列方法：color.SetColor 返回值修复（修复 #4）
// ============================================================

func TestError_ColorOutput(t *testing.T) {
	defer os.Remove("test_error_color.log")

	lp, err := NewLogPool("error_color_test", "test_error_color.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	lp.SetPrintConsole(true)

	// 测试所有 Error 系列方法不会 panic
	lp.Error("test error message")
	lp.Errorf("test error format: %s", "formatted")
	lp.ErrorTag("TAG", "test error with tag")
	lp.ErrorfTag("TAG", "test error format with tag: %s", "formatted")

	// 验证日志文件中写入了内容
	content, err := os.ReadFile("test_error_color.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	logContent := string(content)

	// 验证日志文件中不包含 ANSI 颜色转义码（颜色只应在控制台输出）
	if strings.Contains(logContent, "\033[") {
		t.Error("日志文件中不应包含 ANSI 颜色转义码")
	}

	// 验证日志文件中包含错误消息
	if !strings.Contains(logContent, "test error message") {
		t.Error("日志文件中应包含 'test error message'")
	}
	if !strings.Contains(logContent, "test error format: formatted") {
		t.Error("日志文件中应包含 'test error format: formatted'")
	}
	if !strings.Contains(logContent, "TAG") {
		t.Error("日志文件中应包含 'TAG'")
	}
}

func TestErrorf_Output(t *testing.T) {
	defer os.Remove("test_errorf.log")

	lp, err := NewLogPool("errorf_test", "test_errorf.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	lp.Errorf("error code: %d, message: %s", 404, "not found")

	content, err := os.ReadFile("test_errorf.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if !strings.Contains(string(content), "error code: 404, message: not found") {
		t.Error("Errorf 应正确格式化输出")
	}
}

func TestErrorTag_Output(t *testing.T) {
	defer os.Remove("test_errortag.log")

	lp, err := NewLogPool("errortag_test", "test_errortag.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	lp.ErrorTag("MODULE", "something went wrong")

	content, err := os.ReadFile("test_errortag.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "MODULE") {
		t.Error("ErrorTag 应包含 tag 'MODULE'")
	}
	if !strings.Contains(logContent, "something went wrong") {
		t.Error("ErrorTag 应包含错误消息")
	}
}

func TestErrorfTag_Output(t *testing.T) {
	defer os.Remove("test_errorftag.log")

	lp, err := NewLogPool("errorftag_test", "test_errorftag.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	lp.ErrorfTag("SVC", "request failed: %s (code=%d)", "timeout", 504)

	content, err := os.ReadFile("test_errorftag.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "SVC") {
		t.Error("ErrorfTag 应包含 tag 'SVC'")
	}
	if !strings.Contains(logContent, "request failed: timeout (code=504)") {
		t.Error("ErrorfTag 应正确格式化输出")
	}
}

// ============================================================
// 测试 Close 方法（新增功能 #10）
// ============================================================

func TestClose(t *testing.T) {
	defer os.Remove("test_close.log")

	lp, err := NewLogPool("close_test", "test_close.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}

	if lp.Fd == nil {
		t.Fatal("关闭前 Fd 不应为 nil")
	}

	err = lp.Close()
	if err != nil {
		t.Errorf("Close 不应返回错误: %v", err)
	}

	// 关闭后再次写入应该失败（文件描述符已关闭）
	_, writeErr := lp.Fd.Write([]byte("test"))
	if writeErr == nil {
		t.Error("关闭后写入文件描述符应返回错误")
	}
}

func TestClose_NilFd(t *testing.T) {
	lp := &LogPool{Fd: nil}
	err := lp.Close()
	if err != nil {
		t.Errorf("Fd 为 nil 时 Close 不应返回错误: %v", err)
	}
}

// ============================================================
// 测试 LogPool.Name 字段（新增字段 #7）
// ============================================================

func TestLogPool_NameField(t *testing.T) {
	defer os.Remove("test_namefield.log")

	lp, err := NewLogPool("my_service", "test_namefield.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	if lp.Name != "my_service" {
		t.Errorf("期望 Name='my_service', 实际='%s'", lp.Name)
	}

	// 验证 Name 和 FileName 是独立的
	if lp.Name == lp.FileName {
		t.Error("Name 和 FileName 应该是不同的值")
	}
}

// ============================================================
// 测试 SetPrintConsole
// ============================================================

func TestSetPrintConsole(t *testing.T) {
	defer os.Remove("test_console.log")

	lp, err := NewLogPool("console_test", "test_console.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	if lp.PrintConsole != false {
		t.Error("默认 PrintConsole 应为 false")
	}

	lp.SetPrintConsole(true)
	if lp.PrintConsole != true {
		t.Error("SetPrintConsole(true) 后应为 true")
	}

	lp.SetPrintConsole(false)
	if lp.PrintConsole != false {
		t.Error("SetPrintConsole(false) 后应为 false")
	}
}

// ============================================================
// 测试 Debug/Info 系列方法在不同 Level 下的行为
// ============================================================

func TestDebug_LevelFilter(t *testing.T) {
	defer os.Remove("test_debug_level.log")

	lp, err := NewLogPool("debug_level_test", "test_debug_level.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 设置为 INFO 级别，Debug 不应写入
	lp.SetLevel(INFO_LEVEL)
	lp.Debug("should not appear")
	lp.Debugf("should not appear: %s", "formatted")
	lp.DebugTag("TAG", "should not appear")
	lp.DebugfTag("TAG", "should not appear: %s", "formatted")

	content, err := os.ReadFile("test_debug_level.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if strings.Contains(string(content), "should not appear") {
		t.Error("INFO 级别下 Debug 消息不应被写入日志文件")
	}
}

func TestInfo_LevelFilter(t *testing.T) {
	defer os.Remove("test_info_level.log")

	lp, err := NewLogPool("info_level_test", "test_info_level.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 设置为 ERROR 级别，Info 不应写入
	lp.SetLevel(ERROR_LEVEL)
	lp.Info("should not appear")
	lp.Infof("should not appear: %s", "formatted")
	lp.InfoTag("TAG", "should not appear")
	lp.InfofTag("TAG", "should not appear: %s", "formatted")

	content, err := os.ReadFile("test_info_level.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if strings.Contains(string(content), "should not appear") {
		t.Error("ERROR 级别下 Info 消息不应被写入日志文件")
	}
}

func TestError_AlwaysWritten(t *testing.T) {
	defer os.Remove("test_error_always.log")

	lp, err := NewLogPool("error_always_test", "test_error_always.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp.Close()

	// 即使在 ERROR 级别，Error 也应写入
	lp.SetLevel(ERROR_LEVEL)
	lp.Error("error always written")

	content, err := os.ReadFile("test_error_always.log")
	if err != nil {
		t.Fatalf("读取日志文件失败: %v", err)
	}

	if !strings.Contains(string(content), "error always written") {
		t.Error("Error 消息在任何级别下都应被写入")
	}
}

// ============================================================
// 测试 SaveIntLogMap 的递增编号
// ============================================================

func TestSaveIntLogMap_Ordering(t *testing.T) {
	defer os.Remove("test_order1.log")
	defer os.Remove("test_order2.log")

	lp1, err := NewLogPool("order_test1", "test_order1.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp1.Close()

	lp2, err := NewLogPool("order_test2", "test_order2.log")
	if err != nil {
		t.Fatalf("NewLogPool 创建失败: %v", err)
	}
	defer lp2.Close()

	// 验证两个日志池被分配了不同的编号
	found1 := false
	found2 := false
	for i := 1; i <= 100; i++ {
		result := GetLogNum(i)
		if result == lp1 {
			found1 = true
		}
		if result == lp2 {
			found2 = true
		}
	}

	if !found1 {
		t.Error("lp1 应能通过 GetLogNum 找到")
	}
	if !found2 {
		t.Error("lp2 应能通过 GetLogNum 找到")
	}
}

// ============================================================
// 测试绝对路径下的日志文件创建
// ============================================================

func TestNewLogPool_AbsolutePath(t *testing.T) {
	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "toolbox_test", "abs_test.log")
	defer os.RemoveAll(filepath.Join(tmpDir, "toolbox_test"))

	lp, err := NewLogPool("abs_path_test", logFile)
	if err != nil {
		t.Fatalf("NewLogPool 绝对路径创建失败: %v", err)
	}
	defer lp.Close()

	lp.Info("absolute path test")

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("日志文件应在绝对路径 '%s' 被创建", logFile)
	}
}
