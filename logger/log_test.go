package logger

import "testing"

// go test -v log_test.go logger.go

func TestLogger(t *testing.T) {
	type Sample struct {
		Sex  string
		Age  int64
		Name string
	}
	sample := &Sample{Sex: "man", Age: 45, Name: "Roy"}
	log, err := NewLogPool("test1", "test1.log")
	if err != nil {
		t.Error("NewLogPool err:", err)
	}
	log.Infoln("a", 1)
	log.Infof("%#v", sample)

	sample.Sex = "woman"
	sample.Age = 20
	sample.Name = "Jessica"
	log2 := GetLogName("test1")
	t.Log(log2.Path, log2.FileName)
	if err != nil {
		t.Error("GetLogPool err:", err)
	}
	log2.Debugln("b", 2)
	log2.Debugf("%#v", sample)
	log3 := GetLogNum(1)
	if err != nil {
		t.Error("GetLogPool err:", err)
	}
	log3.Debugln("log3", 3)

}

func TestLogPool_SetLevel(t *testing.T) {
	log, err := NewLogPool("test1", "test1.log")
	if err != nil {
		t.Error("NewLogPool err:", err)
	}
	log.SetLevel(ERROR_LEVEL)
	log.Infoln("write INFO in ERROR_LEVEL")

	log = GetLogName("test1.log")
	if log != nil {
		t.Error("NewLogPool err:", err)
	}
	log.Infoln("write INFO in ERROR_LEVEL")
	log.SetLevel(INFO_LEVEL)
	log.Infoln("write INFO in INFO_LEVEL")
	GetLog().Infoln("QQQQQQ")

}

func TestSetPrefix(t *testing.T) {
	log, err := NewLogPool("test1", "test1.log")
	if err != nil {
		t.Error("NewLogPool err:", err)
	}
	log.infoLogger.Printf("[%s] %s: %s\n", "WARN", "2021-10-01 10:01:00", "This is a warning message.")
}
