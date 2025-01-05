package test

import (
	conf2 "github.com/Chairou/toolbox/conf"
	"testing"
)

func TestCmd(t *testing.T) {
	conf := &conf2.Config{}
	err := LoadConfFromCmd(conf)
	if err != nil {
		return
	}
	t.Log(conf.Env)

}
