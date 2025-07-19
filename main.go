package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// flags
var asTable bool

func init() {
	const (
		tableDefault bool   = false
		tableUsage   string = "generate tests in table format"
	)
	flag.BoolVar(&asTable, "table", tableDefault, tableUsage)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  This tool generates tests for the go package in the current directory.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults() // Prints descriptions for defined flags (like --list)
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// todo return num of generated tests
func main() {
	flag.Parse()

	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	run(asTable)

}

type TestTemplateData struct {
	FuncInfo FuncInfo
}

func (td TestTemplateData) DefaultFail() string {
	return defaultFail
}

const TemplateImports = `package {{ .Name }}

import (
	"testing"
	"reflect"
	{{ .PrintImports }}
)
`

const defaultFail = `t.Fatalf("test not implemented")`

const Template = `
{{ range .}}
func {{ .FuncInfo.PrintTestName }}(t *testing.T) {
	t.Run("{{ .FuncInfo.PrintTestName }}_0", func(t *testing.T) {

		// delete this after your implementation
		{{ .DefaultFail }}

		{{ .FuncInfo.PrintReceiverCtor }}
		{{ .FuncInfo.PrintDefaultVarArgs }}
		{{ .FuncInfo.PrintDefaultReturns }} {{ .FuncInfo.PrintCall }}({{ .FuncInfo.PrintDefaultArgs }})
		{{ .FuncInfo.PrintDefaultExpects }}
	})
}
{{ end }}
`

const TemplateTable = `
{{ range .}}
func {{ .FuncInfo.PrintTestName }}(t *testing.T) {
	tests := []struct {
		testName string
		{{ .FuncInfo.PrintArgsAsStructFields }}
	}{
		// put your test cases here
	}

	// delete this after your implementation
	{{ .DefaultFail }}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			{{ .FuncInfo.PrintReceiverCtor }}
			{{ .FuncInfo.PrintDefaultReturns }} {{ .FuncInfo.PrintCall }}({{ .FuncInfo.PrintTableArgs }})
			{{ .FuncInfo.PrintDefaultExpects }}
		})
	}
}
{{ end }}
`

func extractPackagePrefix(typeName string) (string, string, string) {
	typeName = strings.TrimPrefix(typeName, "*")
	var prefix string
	parts := strings.Split(typeName, "/")
	last := max(0, len(parts)-1)
	pre, shortName, cut := strings.Cut(parts[last], ".")
	if cut {
		prefix = pre
	}
	var packageId string
	if last > 0 {
		packageId = strings.Join(parts[:last], "/")
	}
	return prefix, packageId, shortName
}

func run(asTable bool) {

	pkgPath, _ := os.Getwd()
	pkgInfo := NewPackageInfo(pkgPath)

	templdatas := []TestTemplateData{}
	// var genericsDetected boolG
	for _, fi := range pkgInfo.TestableFuncs() {
		templData := TestTemplateData{
			FuncInfo: fi,
		}
		templdatas = append(templdatas, templData)

	}

	extraImports := []string{}
	for _, fi := range pkgInfo.Funcs {
		for _, pi := range fi.Params {
			extraImports = append(extraImports, pi.Pkg.Path())
		}
	}
	var buf bytes.Buffer
	importTempl, err := template.New("TestImports").Parse(TemplateImports)
	if err != nil {
		panic(err)
	}
	if err := importTempl.Execute(&buf, pkgInfo); err != nil {
		panic(err)
	}
	buf.WriteString("\n")

	var testTemplate string
	if asTable {
		testTemplate = TemplateTable
	} else {
		testTemplate = Template
	}

	funcTestTempl, err := template.New("TestFunction").Parse(testTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	if err := funcTestTempl.Execute(&buf, templdatas); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	outString := buf.String()

	outFileName := pkgInfo.Name + "_test.go"
	err = os.WriteFile(outFileName, []byte(outString), 0644)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "Created file %s\n", outFileName)
	fmt.Fprintf(os.Stdout, "Created %d tests\n", len(templdatas))
	// if genericsDetected {
	// 	fmt.Fprintf(os.Stderr, "Warning: functions use generic types, your test file will not compile.\n")
	// 	fmt.Fprintf(os.Stderr, "Instantiate the types to proceed.\n")
	// }
	goFmt(outFileName)
}

func goFmt(path string) error {
	cmd := exec.Command("gofmt", "-w", path)

	return cmd.Run()
}
