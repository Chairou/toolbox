package conf

import "testing"

func TestSetAndGet(t *testing.T) {
	err := SetEnv("testName", "hairou")
	if err != nil {
		t.Error(err)
	}
	envStr := GetEnvironment("testName")
	if envStr != "chairou" {
		t.Error("get environment err")
	}
	err = UnSetEnv("testName")
	if err != nil {
		t.Error(err)
	}
}
