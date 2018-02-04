package main

const (
	cppTemplate = `//////////////////////////////////////////
// for {{.Path}}
//////////////////////////////////////////

#include "{{.Prefix}}{{.FileName}}.hpp"
{{range .Classes}}
#include "{{.DeclarationFile}}"
{{- end}}

{{- range .Classes}}

//////////////////////////////////////////
// for {{.Name}}
//////////////////////////////////////////

I{{.Name}} *Mock{{.Name}}::mock = nullptr;
bool Mock{{.Name}}::mockEnable = false;
bool *Mock{{.Name}}::useMock = nullptr;

Mock{{.Name}}::Mock{{.Name}}() {
    origin = new {{.Name}}();
    callCount = 0;
}

Mock{{.Name}}::~Mock{{.Name}}() {
    delete origin;
}

{{- $class := .Name -}}
{{range .Funcs}}

// for {{.Return}} {{$class}}::{{.Name}}({{.ArgWithTypes}})
{{.Return}} Mock{{$class}}::{{.Name}}({{.ArgWithTypes}}) {
    if (mockEnable && useMock[callCount]) {
        return mock->{{.Name}}({{.Args}});
    } else {
        return origin->{{.Name}}({{.Args}});
    }
    callCount++;
}
{{- end}}
{{- end}}
`

	hppTemplate = `//////////////////////////////////////////
// for {{.Path}}
//////////////////////////////////////////

#pragma once
{{range .Classes}}
#include "{{.DeclarationFile}}"
{{- end}}

{{range .Classes}}

//////////////////////////////////////////
// for {{.Name}}
//////////////////////////////////////////

class I{{.Name}} {
  public:
    ~I{{.Name}}() = 0;
{{range .Funcs}}
    virtual {{.Return}} {{.Name}}({{.ArgWithTypes}}) = 0;
{{- end}}
};

class Mock{{.Name}} {
  public:
    // for mock
    static I{{.Name}} *mock;
    static bool mockEnable;
    static bool *useMock;

    Mock{{.Name}}();
    ~Mock{{.Name}}();
{{range .Funcs}}
    {{.Return}} {{.Name}}({{.ArgWithTypes}});
{{- end}}

  private:
    // for original
    {{.Name}} *origin;
    int callCount;
};
{{- end}}
`
)
