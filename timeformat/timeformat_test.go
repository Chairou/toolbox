package timeformat

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"testing"
	"time"
)

type Person struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Birthday Time   `json:"birthday"`
}

func TestTimeJson(t *testing.T) {
	now := Time(time.Now())
	t.Log(now)
	src := `{"id":5,"name":"xiaoming","birthday":"2026-06-30 16:09:51"}`
	p := new(Person)
	err := json.Unmarshal([]byte(src), p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
	t.Log(time.Time(p.Birthday))
	//js, _ := json.Marshal(p)
	js, _ := jsoniter.Marshal(p)
	t.Log(string(js))
}
