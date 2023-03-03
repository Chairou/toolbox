package test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRand(t *testing.T) {
	rand.Seed(time.Now().Unix())
	cntMap := make(map[int] int, 0)

	for i:=0; i< 1000000 ; i++ {
		num := rand.Intn(10)
		cntMap[num] = cntMap[num] + 1
	}

	for k,v :=range cntMap {
		fmt.Println(k,v)
	}
}
