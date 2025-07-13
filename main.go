package main

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"
)

func main() {
	fmt.Println("Hello, World")
	tt := TestTemplateData{
		Name: "TestAdd",
		FuncData: FuncData{
			Name: "add",
			ArgsNames: map[string]reflect.Type{
				"a": reflect.TypeFor[int32](),
				"b": reflect.TypeFor[int32](),
			},
			Args:    []reflect.Type{reflect.TypeFor[int32](), reflect.TypeFor[int32]()},
			Returns: []reflect.Type{reflect.TypeFor[int32]()},
		},
		Table: false,
	}

	tmpl, err := template.New("TestFunction").Parse(Template)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("package %s\n\nimport \"testing\"\n\n", tt.FuncData.Name))
	if err := tmpl.Execute(&buf, tt); err != nil {
		panic(err)
	}

	err = os.WriteFile("add/add_test.go", buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

type TestTemplateData struct {
	FuncData FuncData
	Name     string
	Table    bool
}

func (td TestTemplateData) DefaultFail() string {
	return defaultFail
}

type FuncData struct {
	Name      string
	ArgsNames map[string]reflect.Type
	Args      []reflect.Type
	Returns   []reflect.Type
}

func (fd FuncData) PrintDefaultArgs() string {
	args := []byte{}
	written := 0
	for _, argType := range fd.Args {
		if written > 0 {
			args = fmt.Append(args, ", ")
		}
		zero := reflect.Zero(argType)
		args = fmt.Appendf(args, "%v", zero)
		written++
	}
	return string(args)
}

func (fd FuncData) PrintDefaultReturns() string {
	returnNames := ""
	for i := range fd.Returns {
		if i > 0 {
			returnNames += ", "
		}
		returnNames += fmt.Sprintf("result%d", i)
	}
	return returnNames
}

func (fd FuncData) PrintDefaultExpects() string {
	const tmpl = `
		%s := %v
		if %s != %s {
			t.Errorf("Expected %%v, got %%v", %s, %s)
		}
	`
	var expects strings.Builder
	for i, retType := range fd.Returns {
		expectName := fmt.Sprintf("expect%d", i)
		expectZero := reflect.Zero(retType)
		resultName := fmt.Sprintf("result%d", i)
		s := fmt.Sprintf(tmpl, expectName, expectZero, resultName, expectName, expectName, resultName)
		expects.WriteString(s)
	}
	return expects.String()
}

const defaultFail = `t.Fatalf("test not implemented")`

const Template = `
func {{ .Name }}(t *testing.T) {
	t.Run("Adding positive numbers", func(t *testing.T) {
		// delete this after your implementation
		{{ .DefaultFail }}
		{{ .FuncData.PrintDefaultReturns }} := {{ .FuncData.Name }}({{ .FuncData.PrintDefaultArgs }})
		{{ .FuncData.PrintDefaultExpects }}
	})
}
`

const TemplateTable = `
func {{ .Name }}(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{}
	// delete this after your implementation
	{{ defaultFail }}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := {{ .FuncData.Name }}(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Sum(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
`
