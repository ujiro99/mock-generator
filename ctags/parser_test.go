package ctags

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TagFile = "test/tags"
)

var expectFiles = []File{
	{
		Path: "inc/fibonacci.hpp",
		Classes: []*Class{
			{
				Name: "AsyncDPFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int AsyncDPFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "execCalc",
						Signature:    "void AsyncDPFibonacci::execCalc",
						Return:       "void",
						Args:         "x, fib",
						ArgWithTypes: "int x, Fibonacci* fib",
					},
					{
						Name:         "wait",
						Signature:    "void AsyncDPFibonacci::wait",
						Return:       "void",
						Args:         "",
						ArgWithTypes: "",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "DPFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int DPFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name:            "Fibonacci",
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "MemorizeFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int MemorizeFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "RecursiveFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int RecursiveFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
		},
	},
	{
		Path: "src/fibonacci.cpp",
		Classes: []*Class{
			{
				Name: "AsyncDPFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int AsyncDPFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "execCalc",
						Signature:    "void AsyncDPFibonacci::execCalc",
						Return:       "void",
						Args:         "x, fib",
						ArgWithTypes: "int x, Fibonacci* fib",
					},
					{
						Name:         "wait",
						Signature:    "void AsyncDPFibonacci::wait",
						Return:       "void",
						Args:         "",
						ArgWithTypes: "",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "DPFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int DPFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "MemorizeFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int MemorizeFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
			{
				Name: "RecursiveFibonacci",
				Funcs: []Func{
					{
						Name:         "calc",
						Signature:    "int RecursiveFibonacci::calc",
						Return:       "int",
						Args:         "x",
						ArgWithTypes: "int x",
					},
				},
				DeclarationFile: "fibonacci.hpp",
			},
		},
	},
	{
		Path: "inc/counter.hpp",
		Classes: []*Class{
			{
				Name: "Counter",
				Funcs: []Func{
					{
						Name:         "calcTime",
						Signature:    "void Counter::calcTime",
						Return:       "void",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "calcTimeAsync",
						Signature:    "void Counter::calcTimeAsync",
						Return:       "void",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "setFibonacci",
						Signature:    "void Counter::setFibonacci",
						Return:       "void",
						Args:         "fib",
						ArgWithTypes: "Fibonacci *fib",
					},
				},
				DeclarationFile: "counter.hpp",
			},
		},
	},
	{
		Path: "src/counter.cpp",
		Classes: []*Class{
			{
				Name: "Counter",
				Funcs: []Func{
					{
						Name:         "calcTime",
						Signature:    "void Counter::calcTime",
						Return:       "void",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "calcTimeAsync",
						Signature:    "void Counter::calcTimeAsync",
						Return:       "void",
						Args:         "x",
						ArgWithTypes: "int x",
					},
					{
						Name:         "setFibonacci",
						Signature:    "void Counter::setFibonacci",
						Return:       "void",
						Args:         "fib",
						ArgWithTypes: "Fibonacci *fib",
					},
				},
				DeclarationFile: "counter.hpp",
			},
		},
	},
	{
		Path: "src/main.cpp",
		Funcs: []Func{
			{
				Name:         "exec",
				Signature:    "void exec",
				Return:       "void",
				ArgWithTypes: "Counter *counter, vector<Fibonacci *> fib",
				Args:         "counter, fib",
			},
			{
				Name:         "main",
				Signature:    "int main",
				Return:       "int",
				ArgWithTypes: "int argc, char const *argv[]",
				Args:         "argc, argv",
			},
		},
	},
}

func TestParse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	bytes, _ := ioutil.ReadFile(TagFile)
	files := NewParser().Parse(string(bytes))
	require.Equal(len(expectFiles), len(files))

	for indexFile, f := range files {
		t.Logf("file: %s", f.Path)

		expectFile := expectFiles[indexFile]
		assert.Equal(expectFile.Path, f.Path, "File.Path")
		require.Equal(len(expectFile.Classes), len(f.Classes), "len(File.Classes)")

		for indexClass, c := range f.Classes {
			t.Logf("class: %s", c.Name)

			expectClass := expectFile.Classes[indexClass]
			assert.Equal(expectClass.Name, c.Name, "Class.Name")
			require.Equal(len(expectClass.Funcs), len(c.Funcs), "len(Class.Funcs)")

			for indexFunc, f := range c.Funcs {
				t.Logf("class.func: %s", f.Name)

				expectFunc := expectClass.Funcs[indexFunc]
				assert.Equal(expectFunc.Name, f.Name, "Func.Name")
				assert.Equal(expectFunc.Signature, f.Signature, "Func.Signature")
				assert.Equal(expectFunc.Return, f.Return, "Func.Return")
				assert.Equal(expectFunc.Args, f.Args, "Func.Args")
				assert.Equal(expectFunc.ArgWithTypes, f.ArgWithTypes, "Func.ArgWithTypes")
			}

			assert.Equal(expectClass.DeclarationFile, c.DeclarationFile, "Class.DeclarationFile")
		}

		require.Equal(len(expectFile.Funcs), len(f.Funcs), "len(File.Funcs)")
		for indexFunc, f := range f.Funcs {
			t.Logf("file.func: %s", f.Name)

			expectFunc := expectFile.Funcs[indexFunc]
			assert.Equal(expectFunc.Name, f.Name, "Func.Name")
			assert.Equal(expectFunc.Signature, f.Signature, "Func.Signature")
			assert.Equal(expectFunc.Return, f.Return, "Func.Return")
			assert.Equal(expectFunc.Args, f.Args, "Func.Args")
			assert.Equal(expectFunc.ArgWithTypes, f.ArgWithTypes, "Func.ArgWithTypes")
		}
	}
}
