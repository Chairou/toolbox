package bitmap

import "testing"

func TestBitmap(t *testing.T) {
	bitmap := GetBitMap("test1", 15)
	bitmap.Set(0)
	bitmap.MSet(2, 3, 4)
	matchMap := bitmap.MExist(1, 2, 3)
	if matchMap[1] != false || matchMap[2] != true || matchMap[3] != true {
		t.Error("UNExpected error")
		t.Log(matchMap)
	}
	if bitmap.IsExist(0) != true {
		t.Error("UNExpected error")
	}
	bitmap.Clean()
	bitmap.Set(15)
	if !(bitmap.IsExist(15)) {
		t.Error("UNExpected error")
	}
	if bitmap.IsExist(1) {
		t.Error("UNExpected error")
	}
	bitmap.Destroy()

}
