package test

import (
	"regexp"
	"testing"
)

func TestRegexp(t *testing.T) {

	name := "mall-rec-model-release-v2"
	compileRegex := regexp.MustCompile("^(.*)-.*?-.*?$")
	matchArr := compileRegex.FindStringSubmatch(name)
	paasName := ""
	if len(matchArr) == 2 {
		paasName = matchArr[1]
	}
	t.Log(len(matchArr), paasName)
}
