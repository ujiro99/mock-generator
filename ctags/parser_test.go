package ctags

import (
	"io/ioutil"
	"testing"
)

const (
	TagFile = "test/tags"
)

func TestParser(t *testing.T) {
	bytes, _ := ioutil.ReadFile(TagFile)
	files := NewParser().Parse(string(bytes))

	if len(files) != 5 {
		t.Errorf("len(files) != 5: actual %d", len(files))
		for _, v := range files {
			t.Log(v.Path)
		}
	}
}
