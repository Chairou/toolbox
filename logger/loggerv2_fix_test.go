package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// 测试使用工作目录下的相对路径，因为 NewLogOpt 只允许写入工作目录内
const v2TestDir = "testdata_loggerv2"

func newV2TestPool(t *testing.T, name string) *LogPoolV2 {
	t.Helper()
	t.Cleanup(func() { _ = os.RemoveAll(v2TestDir) })
	inst, err := NewLogOpt(name, &LogOpt{FileName: filepath.Join(v2TestDir, name+".log")})
	if err != nil {
		t.Fatalf("NewLogOpt failed: %v", err)
	}
	t.Cleanup(func() { _ = inst.Close() })
	return inst
}

// 日志中的 file:line 必须指向调用 Infof 的测试代码，而不是 logger 包内部
func TestLoggerV2CallerLine(t *testing.T) {
	inst := newV2TestPool(t, "caller_line")

	_, _, line, _ := runtime.Caller(0)
	inst.Infof("caller check")
	_ = inst.Close()

	data, err := os.ReadFile(filepath.Join(v2TestDir, "caller_line.log"))
	if err != nil {
		t.Fatalf("read log failed: %v", err)
	}
	want := fmt.Sprintf("loggerv2_fix_test.go:%d: ", line+1)
	if !strings.Contains(string(data), want) {
		t.Fatalf("expected %q in log, got: %s", want, data)
	}
}

// 以 ".." 开头的合法文件名不应被路径检查拒绝
func TestSafePathDotDotPrefixName(t *testing.T) {
	if _, err := safePath("./", "..app.log"); err != nil {
		t.Fatalf("safePath rejected a legal file name: %v", err)
	}
	if _, err := safePath("./", "../outside.log"); err == nil {
		t.Fatal("safePath accepted a path escaping the base directory")
	}
}

// 并发创建同名日志池只能得到同一个实例
func TestNewLogOptConcurrentSameName(t *testing.T) {
	const name = "concurrent_create"
	t.Cleanup(func() { _ = os.RemoveAll(v2TestDir) })

	var wg sync.WaitGroup
	results := make([]*LogPoolV2, 16)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			inst, err := NewLogOpt(name, &LogOpt{FileName: filepath.Join(v2TestDir, name+".log")})
			if err != nil {
				t.Errorf("NewLogOpt failed: %v", err)
				return
			}
			results[i] = inst
		}(i)
	}
	wg.Wait()
	t.Cleanup(func() {
		if results[0] != nil {
			_ = results[0].Close()
		}
	})

	for i, inst := range results {
		if inst != results[0] {
			t.Fatalf("result %d is a different instance", i)
		}
	}
}

// Close 之后实例从注册表中移除，同名创建得到新实例，旧实例无法再通过名称或编号获取
func TestCloseUnregisters(t *testing.T) {
	const name = "close_unregister"
	t.Cleanup(func() { _ = os.RemoveAll(v2TestDir) })
	opt := &LogOpt{FileName: filepath.Join(v2TestDir, name+".log")}

	old, err := NewLogOpt(name, opt)
	if err != nil {
		t.Fatalf("NewLogOpt failed: %v", err)
	}
	if GetLogNameV2(name) != old {
		t.Fatal("instance not registered by name")
	}
	if err := old.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if GetLogNameV2(name) != nil {
		t.Fatal("instance still registered by name after Close")
	}
	if GetLogNumV2(old.num) != nil {
		t.Fatal("instance still registered by number after Close")
	}

	fresh, err := NewLogOpt(name, opt)
	if err != nil {
		t.Fatalf("NewLogOpt after Close failed: %v", err)
	}
	t.Cleanup(func() { _ = fresh.Close() })
	if fresh == old {
		t.Fatal("NewLogOpt returned the closed instance")
	}
	if GetLogNameV2(name) != fresh {
		t.Fatal("new instance not registered by name")
	}
}

// Close 之后可以安全再次写入，文件会被重新打开
func TestCloseThenWrite(t *testing.T) {
	inst := newV2TestPool(t, "close_write")
	if err := inst.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	inst.Infof("after close")
	_ = inst.Close()

	data, err := os.ReadFile(filepath.Join(v2TestDir, "close_write.log"))
	if err != nil {
		t.Fatalf("read log failed: %v", err)
	}
	if !strings.Contains(string(data), "after close") {
		t.Fatalf("write after Close was lost: %s", data)
	}
}

// v1 的 LogPool 同样必须报告调用方的 file:line
func TestLogPoolCallerLine(t *testing.T) {
	const name = "v1_caller_line"
	dir := filepath.Join(v2TestDir, "v1")
	t.Cleanup(func() { _ = os.RemoveAll(v2TestDir) })

	inst, err := NewLogPool(name, filepath.Join(dir, "caller.log"))
	if err != nil {
		t.Fatalf("NewLogPool failed: %v", err)
	}
	_, _, line, _ := runtime.Caller(0)
	inst.Infof("v1 caller check")
	_ = inst.Close()

	data, err := os.ReadFile(filepath.Join(dir, "caller.log"))
	if err != nil {
		t.Fatalf("read log failed: %v", err)
	}
	want := fmt.Sprintf("loggerv2_fix_test.go:%d: ", line+1)
	if !strings.Contains(string(data), want) {
		t.Fatalf("expected %q in log, got: %s", want, data)
	}
}
