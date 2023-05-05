package test

import "testing"

func TestMake(t *testing.T) {
	list := make([]string, 0, 1024)
	list = append(list, "111")
	t.Log(list)
}

func TestMakeMap(t *testing.T) {
	map1 := make(map[string]string, 32)
	map1["aaa"] = "111"
	t.Log(map1)
}
