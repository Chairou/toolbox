package conf

import (
	"os"
	"path/filepath"
	"testing"
)

// 与 dev.yaml 内容一致，用于验证 yaml.v3 下每个字段都被正确解析
const sampleYAML = `env: dev
version: 123
redis_name : redis1
redis_host : 127.0.0.1:6379
redis_auth : XXXX
mysql_name : mysql1
mysql_host : 127.0.0.1:3306
mysql_user : root
mysql_pass : XXXX
mysql_db : dbSoftCenter
mysql_charset : utf8mb4
log_file_name : srv.log
`

func sampleConfig() Config {
	return Config{
		Env:          "dev",
		Version:      123,
		RedisName:    "redis1",
		RedisHost:    "127.0.0.1:6379",
		RedisAuth:    "XXXX",
		MysqlName:    "mysql1",
		MysqlHost:    "127.0.0.1:3306",
		MysqlUser:    "root",
		MysqlPass:    "XXXX",
		MysqlDb:      "dbSoftCenter",
		MysqlCharSet: "utf8mb4",
		LogFileName:  "srv.log",
	}
}

func assertConfig(t *testing.T, got, want Config) {
	t.Helper()
	if got != want {
		t.Errorf("config mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

func TestLoadConfFromFile_AllFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.yaml")
	if err := os.WriteFile(path, []byte(sampleYAML), 0600); err != nil {
		t.Fatalf("write sample yaml: %v", err)
	}

	var c Config
	if err := loadConfFromFile(path, &c); err != nil {
		t.Fatalf("loadConfFromFile returned error: %v", err)
	}
	assertConfig(t, c, sampleConfig())
}

// 直接读取 conf/dev.yaml，断言它的实际内容。
// 该文件没有 mysql_name，因此 MysqlName 应为空
func TestLoadConfFromFile_DevYAML(t *testing.T) {
	want := sampleConfig()
	want.MysqlName = ""

	var c Config
	if err := loadConfFromFile("dev.yaml", &c); err != nil {
		t.Fatalf("loadConfFromFile(dev.yaml) returned error: %v", err)
	}
	assertConfig(t, c, want)
}

func TestLoadConfFromFile_MissingFile(t *testing.T) {
	var c Config
	if err := loadConfFromFile(filepath.Join(t.TempDir(), "missing.yaml"), &c); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfFromFile_TypeMismatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, []byte("version: not-a-number\n"), 0600); err != nil {
		t.Fatalf("write bad yaml: %v", err)
	}

	var c Config
	if err := loadConfFromFile(path, &c); err == nil {
		t.Fatal("expected error when int field gets a string")
	}
}
