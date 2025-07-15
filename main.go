package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
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

type PackageInfo struct {
	ModuleName       string
	PackageIsMain    bool
	Path             string
	Name             string
	Imports          map[string]*packages.Package
	MustPrintImports map[string]*packages.Package
	Funcs            []FuncInfo
	goRoot           string
}

func (pi PackageInfo) TestableFuncs() []FuncInfo {
	funcInfo := []FuncInfo{}
	for _, fi := range pi.Funcs {
		if fi.Name == "init" {
			continue
		}
		funcInfo = append(funcInfo, fi)
	}
	return funcInfo
}

func (pi *PackageInfo) Add(fi *FuncInfo) {
	requiredImport := fi.RequiredImports()
	if len(pi.Imports) > 0 {
		for _, ri := range requiredImport {
			if ri == pi.ModuleName {
				continue
			}
			if ri == pi.ModuleName+"/"+pi.Name {
				continue
			}
			// typePath
			pkg, ok := pi.Imports[ri]
			if !ok {
				fmt.Printf("moduleName=%s, required pkg=%s, currentpkg=%s\n", pi.ModuleName, ri, pi.Name)
				fmt.Printf("req=%s not found in %+v\n", ri, pi.Imports)
				panic("bad")
			}
			pi.MustPrintImports[ri] = pkg
		}
	}

	for _, param := range fi.Params {
		a, b, isPtr := strings.Cut(param.ZeroValue, "*")
		var shortZero string
		if isPtr {
			shortZero = b
		} else {
			shortZero = a
		}

		for path := range pi.Imports {
			shortZero = strings.TrimPrefix(shortZero, path)
		}
		prependPgk := false
		if strings.HasPrefix(shortZero, ".") {
			prependPgk = true
		}
		shortZero = strings.TrimPrefix(shortZero, ".")
		if prependPgk {
			pkgs := strings.Split(param.ImportedFrom, "/")
			shortZero = pkgs[len(pkgs)-1] + "." + shortZero
		}
		shortZero = strings.TrimPrefix(shortZero, pi.Path)
		shortZero = strings.TrimPrefix(shortZero, ".")
		if isPtr {
			param.ZeroValue = "*" + shortZero
		} else {
			param.ZeroValue = shortZero
		}
	}
	for _, ret := range fi.Returns {
		a, b, isPtr := strings.Cut(ret.ZeroValue, "*")
		var shortZero string
		if isPtr {
			shortZero = b
		} else {
			shortZero = a
		}

		for path := range pi.Imports {
			shortZero = strings.TrimPrefix(shortZero, path)
		}
		prependPgk := false
		if strings.HasPrefix(shortZero, ".") {
			prependPgk = true
		}
		shortZero = strings.TrimPrefix(shortZero, ".")
		if prependPgk {
			pkgs := strings.Split(ret.ImportedFrom, "/")
			shortZero = pkgs[len(pkgs)-1] + "." + shortZero
		}
		shortZero = strings.TrimPrefix(shortZero, pi.Path)
		shortZero = strings.TrimPrefix(shortZero, ".")
		if isPtr {
			ret.ZeroValue = "*" + shortZero
		} else {
			ret.ZeroValue = shortZero
		}
	}

	pi.Funcs = append(pi.Funcs, *fi)
}

func (pi PackageInfo) PrintImports() string {
	var sb strings.Builder
	for name, pkg := range pi.MustPrintImports {
		var importStr string
		if isStandardLibrary(pkg, pi.goRoot) {
			importStr = strings.TrimPrefix(pi.goRoot, name)
		} else {
			importStr = name
		}
		sb.WriteString(fmt.Sprintf("\"%s\"\n", importStr))
	}
	return sb.String()
}

type FuncInfo struct {
	Name          string
	IsExported    bool
	ReceiverType  string // Fully qualified type string (e.g., "myproject.com/mymodule/mypackage.MyStruct")
	ReceiverShort string
	Params        []*ParamInfo
	Returns       []*ParamInfo // For Go, return values are also like parameters
	ImportedFrom  string
	HasTypeParams bool
}

func (fi FuncInfo) RequiredImports() []string {
	var required []string
	for _, p := range fi.Params {
		if p.ImportedFrom != "" {
			required = append(required, p.ImportedFrom)
		}

	}
	for _, p := range fi.Returns {
		if p.ImportedFrom != "" {
			required = append(required, p.ImportedFrom)
		}

	}
	return required
}

type ParamInfo struct {
	Name         string
	TypeName     string // Fully qualified type string (e.g., "string", "io.Reader")
	ZeroValue    string // Go code for the zero value (e.g., `""`, `0`, `nil`)
	ImportedFrom string
}

func generateZeroValue(typ types.Type) string {
	switch t := typ.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return "false"
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
			return "0"
		case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr:
			return "0"
		case types.Float32, types.Float64:
			return "0.0"
		case types.Complex64, types.Complex128:
			return "(0 + 0i)"
		case types.String:
			return `""`
		default: // e.g., types.UnsafePointer
			return "nil"
		}
	case *types.Pointer, *types.Interface, *types.Map, *types.Chan, *types.Slice, *types.Signature:
		return "nil"
	case *types.Struct:
		// For structs, return an empty composite literal: "pkg.StructType{}"
		return t.String() + "{}"
	case *types.Array:
		// For arrays, return an empty composite literal: "[N]Type{}"
		return t.String() + "{}"
	default:
		// Fallback for any other complex type, often nil or its type name followed by {}
		// For named types, t.String() gives the fully qualified name.
		return t.String() + "{}" // This might not be perfect for all cases, but generally works.
	}
}

func (fi FuncInfo) PrintDefaultArgs() string {
	var args strings.Builder
	written := 0
	for _, arg := range fi.Params {
		if written > 0 {
			args.WriteString(", ")
		}
		args.WriteString(arg.Name)
		written++
	}
	return args.String()
}

func (fd FuncInfo) PrintDefaultReturns() string {
	returnNames := ""
	for i := range fd.Returns {
		if i > 0 {
			returnNames += ", "
		}
		returnNames += fmt.Sprintf("result%d", i)
	}
	if len(fd.Returns) > 0 {
		returnNames += ":="
	}
	return returnNames
}

func (fi FuncInfo) PrintReceiverCtor() string {
	ctor := ""
	if fi.ReceiverType == "" {
		return ctor
	}
	return fmt.Sprintf("receiver := %s{}", fi.ReceiverShort)
}

func (fi FuncInfo) PrintCall() string {
	if fi.ReceiverType == "" {
		return fi.Name
	}
	return fmt.Sprintf("receiver.%s", fi.Name)
}

func (fd FuncInfo) PrintDefaultExpects() string {
	const tmpl = `
		var %s %s
		if !reflect.DeepEqual(%s, %s) {
			t.Errorf("Expected %%v, got %%v", %s, %s)
		}
	`
	var expects strings.Builder
	for i, retType := range fd.Returns {
		expectName := fmt.Sprintf("expect%d", i)
		// expcectType := retType.TypeName
		expectType := retType.ZeroValue
		resultName := fmt.Sprintf("result%d", i)
		s := fmt.Sprintf(tmpl, expectName, expectType, resultName, expectName, expectName, resultName)
		expects.WriteString(s)
	}
	return expects.String()
}

func (fd FuncInfo) PrintDefaultVarArgs() string {
	var sb strings.Builder
	for _, param := range fd.Params {
		def := fmt.Sprintf("var %s %s\n", param.Name, param.ZeroValue)
		sb.WriteString(def)
	}
	return sb.String()
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
					fInfo := FuncInfo{
						Name: funcDecl.Name.Name,
					}

					// check if func has receiver
					if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
						recvTypeExpr := funcDecl.Recv.List[0].Type
						if typ := pkg.TypesInfo.TypeOf(recvTypeExpr); typ != nil {
							fInfo.ReceiverType = typ.String()
							short := getSimplifiedTypeName(typ.String())
							fInfo.ReceiverShort = short
						} else {
							fInfo.ReceiverType = "UNKNOWN_RECEIVER_TYPE"
						}
					}

					if tParams := funcDecl.Type.TypeParams; tParams != nil {
						if list := tParams.List; list != nil {
							if len(list) > 0 {
								fInfo.HasTypeParams = true
							}
						}
					}
					// check params
					if funcDecl.Type.Params != nil {
						for _, field := range funcDecl.Type.Params.List {
							if typ := pkg.TypesInfo.TypeOf(field.Type); typ != nil {
								paramTypeName := typ.String()
								zeroVal := paramTypeName
								var importedFrom string
								prefix, pkgId, _ := extractPackagePrefix(paramTypeName)
								if pkgId != "" {
									importedFrom = pkgId + "/" + prefix
								} else if prefix != "" {
									importedFrom = prefix
								} else {
									importedFrom = ""
								}
								if len(field.Names) == 0 { // Unnamed parameter
									fInfo.Params = append(
										fInfo.Params,
										&ParamInfo{TypeName: paramTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom},
									)
								} else {
									for _, name := range field.Names {
										fInfo.Params = append(
											fInfo.Params,
											&ParamInfo{Name: name.Name, TypeName: paramTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom},
										)
									}
								}
							} else {
								fInfo.Params = append(fInfo.Params, &ParamInfo{Name: "UNKNOWN_PARAMETER_TYPE", TypeName: "UNKNOWN_PARAMETER_TYPE"})
							}

						}
					}

					if funcDecl.Type.Results != nil {
						for _, field := range funcDecl.Type.Results.List {
							if typ := pkg.TypesInfo.TypeOf(field.Type); typ != nil {
								// Get the fully qualified type name for the return value
								returnTypeName := typ.String()
								zeroVal := returnTypeName
								var importedFrom string
								prefix, pkgId, _ := extractPackagePrefix(returnTypeName)
								if pkgId != "" {
									importedFrom = pkgId + "/" + prefix
								} else if prefix != "" {
									importedFrom = prefix
								} else {
									importedFrom = ""
								}
								if len(field.Names) == 0 { // Unnamed return value
									fInfo.Returns = append(fInfo.Returns, &ParamInfo{TypeName: returnTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom})
								} else {
									for _, name := range field.Names {
										fInfo.Returns = append(fInfo.Returns, &ParamInfo{Name: name.Name, TypeName: returnTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom})
									}
								}
							} else {
								fInfo.Returns = append(fInfo.Returns, &ParamInfo{Name: "UNKNOWN_RETURN_TYPE", TypeName: "UNKNOWN_RETURN_TYPE"})
							}
						}
					}
					pkgInfo.Add(&fInfo)
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
