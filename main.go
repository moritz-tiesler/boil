package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// todo return num of generated tests
func main() {
	pkgPath, _ := os.Getwd()
	pkgInfo := listPackageFuncs(pkgPath)

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

	funcTestTempl, err := template.New("TestFunction").Parse(Template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	for _, td := range templdatas {
		if err := funcTestTempl.Execute(&buf, td); err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
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

func listPackageFuncs(pkgPath string) PackageInfo {

	goroot, err := getGOROOT()
	if err != nil {
		panic(err)
	}
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.LoadFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedImports |
			packages.NeedModule |
			packages.NeedCompiledGoFiles |
			packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		panic(err)
	}

	if len(pkgs) != 1 {
		panic("expected one pkg")
	}

	if packages.PrintErrors(pkgs) > 0 {
		panic("packages containged errors")
	}

	pkgInfo := PackageInfo{
		goRoot:           goroot,
		MustPrintImports: make(map[string]*packages.Package),
	}

	for _, pkg := range pkgs {
		if pkg.Module != nil {
			pkgInfo.ModuleName = pkg.Module.Path
			pkgInfo.PackageIsMain = pkg.Module.Main
		}
		pkgInfo.Imports = pkg.Imports
		pkgInfo.Name = pkg.Name
		pkgInfo.Path = pkg.PkgPath
		for _, file := range pkg.Syntax {
			fset := pkg.Fset
			fName := fset.Position(file.Pos()).Filename
			if strings.HasSuffix(fName, "_test.go") {
				continue
			}
			ast.Inspect(file, func(n ast.Node) bool {
				if funcDecl, ok := n.(*ast.FuncDecl); ok {
					funcInfo := NewFunctionInfo(funcDecl, pkg)
					pkgInfo.Add(funcInfo)
				}
				return true
			})
		}

	}
	return pkgInfo

}

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

func getGOROOT() (string, error) {
	// goroot := os.Getenv("GOROOT")

	cmd := exec.Command("go", "env", "GOROOT")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to read GOOROOT environment variable")
	}
	goroot := string(out)
	if goroot == "" {
		return "", fmt.Errorf("GOROOT environment variable is not set")
	}
	return filepath.Clean(goroot), nil
}
