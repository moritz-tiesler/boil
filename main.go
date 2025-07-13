package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// todo return num of generated tests
func main() {
	pkgPath, _ := os.Getwd()
	funcInfos, pkgName := listPackageFuncs(pkgPath)
	templdatas := []TestTemplateData{}
	for _, fi := range funcInfos {
		templData := TestTemplateData{
			Name: fmt.Sprintf("Test%s%s",
				strings.ToUpper(fi.Name[:1]),
				fi.Name[1:],
			),
			FuncInfo: fi,
			Table:    false,
		}
		templdatas = append(templdatas, templData)
	}

	tmpl, err := template.New("TestFunction").Parse(Template)
	if err != nil {
		panic(err)
	}

	extraImports := []string{}
	for _, fi := range funcInfos {
		for _, pi := range fi.Params {
			if pi.ImportedFrom != "" {
				extraImports = append(extraImports, pi.ImportedFrom)
			}
		}
	}
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	buf.WriteString("import \"testing\"\n")
	for _, ei := range extraImports {
		buf.WriteString(fmt.Sprintf("import \"%s\"\n", ei))
	}
	buf.WriteString("\n")

	for _, td := range templdatas {
		if err := tmpl.Execute(&buf, td); err != nil {
			panic(err)
		}

	}

	outString := buf.String()

	outFileName := pkgName + "_test.go"
	err = os.WriteFile(outFileName, []byte(outString), 0644)

	if err != nil {
		panic(err)
	}
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

type FuncInfo struct {
	Name          string
	IsExported    bool
	ReceiverType  string // Fully qualified type string (e.g., "myproject.com/mymodule/mypackage.MyStruct")
	ReceiverShort string
	Params        []ParamInfo
	Returns       []ParamInfo // For Go, return values are also like parameters
	ImportedFrom  string
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
	args := []byte{}
	written := 0
	for _, arg := range fi.Params {
		if written > 0 {
			args = fmt.Append(args, ", ")
		}
		args = fmt.Appendf(args, "%v", arg.ZeroValue)
		written++
	}
	return string(args)
}

func (fd FuncInfo) PrintDefaultReturns() string {
	returnNames := ""
	for i := range fd.Returns {
		if i > 0 {
			returnNames += ", "
		}
		returnNames += fmt.Sprintf("result%d", i)
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
		%s := %v
		if %s != %s {
			t.Errorf("Expected %%v, got %%v", %s, %s)
		}
	`
	var expects strings.Builder
	for i, retType := range fd.Returns {
		expectName := fmt.Sprintf("expect%d", i)
		expectZero := retType.ZeroValue
		resultName := fmt.Sprintf("result%d", i)
		s := fmt.Sprintf(tmpl, expectName, expectZero, resultName, expectName, expectName, resultName)
		expects.WriteString(s)
	}
	return expects.String()
}

const defaultFail = `t.Fatalf("test not implemented")`

const Template = `
func {{ .Name }}(t *testing.T) {
	t.Run("{{ .Name }}_0", func(t *testing.T) {
		// delete this after your implementation
		{{ .DefaultFail }}
		{{ .FuncInfo.PrintReceiverCtor }}
		{{ .FuncInfo.PrintDefaultReturns }} := {{ .FuncInfo.PrintCall }}({{ .FuncInfo.PrintDefaultArgs }})
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

func listPackageFuncs(pkgPath string) ([]FuncInfo, string) {

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedTypes |
			packages.NeedSyntax |
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

	var funcsToTest []FuncInfo
	var pkgName string

	for _, pkg := range pkgs {
		pkgName = pkg.Name
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

					// check params
					if funcDecl.Type.Params != nil {
						for _, field := range funcDecl.Type.Params.List {
							if typ := pkg.TypesInfo.TypeOf(field.Type); typ != nil {
								paramTypeName := typ.String()
								zeroVal := generateZeroValue(typ)
								if len(field.Names) == 0 { // Unnamed parameter
									fInfo.Params = append(fInfo.Params, ParamInfo{TypeName: paramTypeName, ZeroValue: zeroVal})
								} else {
									for _, name := range field.Names {
										pInfo := ParamInfo{Name: name.Name, ZeroValue: zeroVal}
										prefix, pkgId, shortName := extractPackagePrefix(paramTypeName)
										if prefix != "" {
											pInfo.TypeName = prefix + "." + shortName
										} else {
											pInfo.TypeName = shortName
										}
										if prefix != "" && prefix != pkgName {
											if pkgId != "" {
												// third party
												pInfo.ImportedFrom = pkgId + "/" + prefix
											} else {
												pInfo.ImportedFrom = prefix
											}
											pInfo.ZeroValue = pInfo.TypeName + "{}"
										}
										fInfo.Params = append(fInfo.Params, pInfo)
									}
								}
							} else {
								fInfo.Params = append(fInfo.Params, ParamInfo{Name: "UNKNOWN_PARAMETER_TYPE", TypeName: "UNKNOWN_PARAMETER_TYPE"})
							}

						}
					}

					if funcDecl.Type.Results != nil {
						for _, field := range funcDecl.Type.Results.List {
							if typ := pkg.TypesInfo.TypeOf(field.Type); typ != nil {
								// Get the fully qualified type name for the return value
								returnTypeName := typ.String()
								zeroVal := generateZeroValue(typ)
								if len(field.Names) == 0 { // Unnamed return value
									fInfo.Returns = append(fInfo.Returns, ParamInfo{TypeName: returnTypeName, ZeroValue: zeroVal})
								} else {
									for _, name := range field.Names {
										fInfo.Returns = append(fInfo.Returns, ParamInfo{Name: name.Name, TypeName: returnTypeName, ZeroValue: zeroVal})
									}
								}
							} else {
								fInfo.Returns = append(fInfo.Returns, ParamInfo{Name: "UNKNOWN_RETURN_TYPE", TypeName: "UNKNOWN_RETURN_TYPE"})
							}
						}
					}
					funcsToTest = append(funcsToTest, fInfo)
				}
				return true
			})
		}

	}
	return funcsToTest, pkgName

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
