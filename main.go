package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
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
		// fmt.Fprintf(os.Stderr, "Usage examples:\n")
		// fmt.Fprintf(os.Stderr, "  %s <language>        (e.g., %s go, %s python, %s node)\n", os.Args[0], os.Args[0], os.Args[0], os.Args[0])
		// fmt.Fprintf(os.Stderr, "  %s --list            (To see all available languages)\n", os.Args[0])
		// fmt.Fprintf(os.Stderr, "\n")
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
	Name     string
	Table    bool
}

func (td TestTemplateData) DefaultFail() string {
	return defaultFail
}

type Arg struct {
	Type       reflect.Type
	IsStruct   bool
	IsUserDecl bool
}

func (arg Arg) PrintDefaultCtor() string {
	if arg.IsStruct {
		return fmt.Sprintf("%s{}", arg.Type.Name())
	}
	if arg.Type.Kind() == reflect.Ptr {
		return fmt.Sprintf("&%s{}", arg.Type.Name())
	}
	if arg.IsUserDecl {
		return fmt.Sprintf("%s(%s)", arg.Type.Name(), reflect.Zero(arg.Type))
	}
	return ""
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
func {{ .Name }}(t *testing.T) {
	t.Run("{{ .Name }}_0", func(t *testing.T) {

		// delete this after your implementation
		{{ .DefaultFail }}

		{{ .FuncInfo.PrintReceiverCtor }}
		{{ .FuncInfo.PrintDefaultVarArgs }}
		{{ .FuncInfo.PrintDefaultReturns }} {{ .FuncInfo.PrintCall }}({{ .FuncInfo.PrintDefaultArgs }})
		{{ .FuncInfo.PrintDefaultExpects }}
	})
}
`

const TemplateTable = `
func {{ .Name }}(t *testing.T) {
	tests := []struct {
		{{ .FuncInfo.PrintArgsAsStructFields }}
	}{
		// put your test cases here
	}

	// delete this after your implementation
	{{ .DefaultFail }}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			{{ .FuncInfo.PrintReceiverCtor }}
			{{ .FuncInfo.PrintDefaultReturns }} {{ .FuncInfo.PrintCall }}({{ .FuncInfo.PrintTableArgs }})
			{{ .FuncInfo.PrintDefaultExpects }}
		})
	}
}
`

func getSimplifiedTypeName(qualifiedType string) string {
	parts := strings.Split(qualifiedType, "/")

	_, typeName, _ := strings.Cut(parts[len(parts)-1], ".")
	return typeName
}

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

func isStandardLibrary(pkg *packages.Package, goroot string) bool {
	if len(pkg.GoFiles) == 0 {
		// If there are no Go files, it's unlikely a compilable standard library package.
		// Some pseudo-packages might have no GoFiles, but usually they are not
		// direct dependencies one would check.
		return false
	}

	// Construct the expected standard library source path
	stdLibSrcPrefix := filepath.Join(goroot, "src")

	// Check if any of the package's Go files are located within GOROOT/src
	for _, goFile := range pkg.GoFiles {
		// Clean the file path to handle potential "../" or other non-canonical forms
		cleanGoFile := filepath.Clean(goFile)

		// Check if the file path has the GOROOT/src prefix
		if strings.HasPrefix(cleanGoFile, stdLibSrcPrefix) {
			return true
		}
	}
	return false
}

func run(asTable bool) {

	pkgPath, _ := os.Getwd()
	pkgInfo := NewPackageInfo(pkgPath)

	templdatas := []TestTemplateData{}
	var genericsDetected bool
	for _, fi := range pkgInfo.TestableFuncs() {
		templData := TestTemplateData{
			Name: fmt.Sprintf("Test%s%s",
				strings.ToUpper(fi.Name[:1]),
				fi.Name[1:],
			),
			FuncInfo: fi,
			Table:    false,
		}
		templdatas = append(templdatas, templData)

		if fi.HasTypeParams {
			genericsDetected = true
		}
	}

	extraImports := []string{}
	for _, fi := range pkgInfo.Funcs {
		for _, pi := range fi.Params {
			if pi.ImportedFrom != "" {
				extraImports = append(extraImports, pi.ImportedFrom)
			}
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
	for _, td := range templdatas {
		if err := funcTestTempl.Execute(&buf, td); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

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
	if genericsDetected {
		fmt.Fprintf(os.Stderr, "Warning: functions use generic types, your test file will not compile.\n")
		fmt.Fprintf(os.Stderr, "Instantiate the types to proceed.\n")
	}
	goFmt(outFileName)
}

func goFmt(path string) error {
	cmd := exec.Command("gofmt", "-w", path)

	return cmd.Run()
}
