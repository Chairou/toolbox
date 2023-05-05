package color

import "testing"

func TestColor(t *testing.T) {
	t.Log(SetColor(Red, "1234"), "5678")
	t.Log(SetFbColor(Red, YellowBackground, 1234), "5678")
}
