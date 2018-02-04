package ctags

import (
	"io/ioutil"
	"testing"
)

const (
	TagFile = "test/tags"
)

var expectFiles = []File{
	{
		Path: ".\\inc\\fibonacci.hpp",
		Classes: []*Class{
			{
				Name: "AsyncDPFibonacci",
			},
			{
				Name: "DPFibonacci",
			},
			{
				Name: "Fibonacci",
			},
			{
				Name: "MemorizeFibonacci",
			},
			{
				Name: "RecursiveFibonacci",
			},
		},
	},
	{
		Path: ".\\src\\fibonacci.cpp",
		Classes: []*Class{
			{
				Name: "AsyncDPFibonacci",
			},
			{
				Name: "DPFibonacci",
			},
			{
				Name: "MemorizeFibonacci",
			},
			{
				Name: "RecursiveFibonacci",
			},
			{
				Name: "Fibonacci",
			},
		},
	},
	{
		Path: ".\\inc\\counter.hpp",
		Classes: []*Class{
			{
				Name: "Counter",
			},
		},
	},
	{
		Path: ".\\src\\counter.cpp",
		Classes: []*Class{
			{
				Name: "Counter",
			},
		},
	},
	{
		Path: ".\\src\\main.cpp",
	},
}

func TestParser(t *testing.T) {
	bytes, _ := ioutil.ReadFile(TagFile)
	files := NewParser().Parse(string(bytes))

	if len(files) != 5 {
		t.Fatalf("len(files) != 5: actual %d", len(files))
	}

	for indexFile, f := range files {
		expectFile := expectFiles[indexFile]
		if expectFile.Path != f.Path {
			t.Fatalf("%s != %s", expectFile.Path, f.Path)
		}

		for indexClass, c := range f.Classes {
			expectClass := expectFile.Classes[indexClass]
			if expectClass.Name != c.Name {
				t.Fatalf("%s != %s", expectClass.Name, c.Name)
			}
		}
	}
}
