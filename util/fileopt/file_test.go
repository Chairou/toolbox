package fileopt

import "testing"

func TestGetDirFilesCommon(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./", []string{"*"})
	if err != nil {
		t.Error(err)
	}
	t.Log(files)
}

func TestGetDirFilesSingle(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./", []string{".txt"})
	if err != nil {
		t.Error(err)
	}
	t.Log(files)
}

func TestGetDirFilesMulti(t *testing.T) {
	files, err := GetAllFilesFromDirectory("./", []string{".txt", ".go"})
	if err != nil {
		t.Error(err)
	}
	t.Log(files)
}

func TestStringNil(t *testing.T) {
	var ret []byte
	ret = nil
	t.Log(string(ret))
}
