package bitmap

import "testing"

func TestBitmap(t *testing.T) {
	bitmap := NewBitMap("test1", 16)
	err := bitmap.Set(0)
	if err != nil {
		t.Error(err)
	}

	err = bitmap.MSet(2, 3, 4)
	if err != nil {
		t.Error(err)
	}

	matchMap := bitmap.MExist(1, 2, 3)
	if matchMap[1] != false || matchMap[2] != true || matchMap[3] != true {
		t.Error("UNExpected error")
		t.Log(matchMap)
	}
	if bitmap.IsExist(0) != true {
		t.Error("UNExpected error")
	}
	bitmap.PrintAllBits()
	bitmap.Clean()
	err = bitmap.Set(15)
	if err != nil {
		t.Error(err)
	}
	if !(bitmap.IsExist(15)) {
		t.Error("UNExpected error")
	}
	if bitmap.IsExist(1) {
		t.Error("UNExpected error")
	}
	bitmap.PrintAllBits()
	bitmap.Destroy()

}
