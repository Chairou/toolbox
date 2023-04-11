package cal

import "testing"

func TestCeilFloat(t *testing.T) {
	pai := 3.1415926
	result := CeilFloat(pai, 4)
	t.Log(result)
}

func TestRoundFloat(t *testing.T) {
	pai := 3.1415926
	result := RoundFloat(pai, 4)
	t.Log(result)
}

func TestFloorFloat(t *testing.T) {
	pai := 3.1415926
	result := FloorFloat(pai, 4)
	t.Log(result)
}

func TestPercent(t *testing.T) {
	pai := 0.1415926
	result := Percent(pai, 4, CUT_OFF_ROUND)
	t.Log(result)
}

func TestPercentVal(t *testing.T) {
	pai := 0.1415926
	result := PercentVal(pai, 4, CUT_OFF_ROUND)
	t.Log(result)
}

func TestAnyPercent(t *testing.T) {
	pai := "0.1415926"
	result := AnyPercent(pai, 4, CUT_OFF_CEIL)
	t.Log(result)
}
