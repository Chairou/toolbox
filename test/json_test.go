package test

import (
	jsoniter "github.com/json-iterator/go"
	"testing"
)

func TestJsonUnmarshal(t *testing.T) {
	type Sample struct {
		Name	string	`json:"name"`
		Age		int		`json:"age"`
	}
	var people Sample

	tmpStr := "{\"name\":\"Chair\", \"age\":44}"
	err := jsoniter.UnmarshalFromString(tmpStr, &people)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", people)
	}

}
