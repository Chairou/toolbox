package test

import (
	"fmt"
	"sync"
	"testing"
)

func TestMap(t *testing.T) {
	aMap := make(map[string]string, 0)
	aMap["aaaa"] = "nnn"
	wg := sync.WaitGroup{}
	wg.Add(3)
	go walkMap(aMap)
	go walkMap(aMap)
	go walkMap(aMap)
	wg.Wait()
}


func walkMap(abc map[string]string) {
	for true{
		for k, v :=range abc {
			fmt.Println(k, v)
		}
	}
}
