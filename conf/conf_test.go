package conf

import (
	"testing"
)

func TestYamlUnmarshal(t *testing.T) {
	conf := LoadConfig()
	t.Logf("%#v", conf)
	env := GetEnvironment("env")
	if env == "" {
		env = "dev"
	}
	if env != conf.Env {
		t.Error("unexpected environment")
	}
}
