package main

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

type FuncInfo struct {
	Name          string
	IsExported    bool
	ReceiverType  string // Fully qualified type string
	ReceiverShort string
	Params        []*ParamInfo
	Returns       []*ParamInfo // For Go, return values are also like parameters
	ImportedFrom  string
	HasTypeParams bool
}

func NewFunctionInfo(funcDecl *ast.FuncDecl, pkg *packages.Package) *FuncInfo {

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
			fInfo.Params = append(fInfo.Params, NewParamInfo(field, pkg)...)

		}
	}

	// check returns
	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			fInfo.Returns = append(fInfo.Returns, NewParamInfo(field, pkg)...)

		}
	}
	return &fInfo
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

func (fi FuncInfo) PrintDefaultReturns() string {
	returnNames := ""
	for i := range fi.Returns {
		if i > 0 {
			returnNames += ", "
		}
		returnNames += fmt.Sprintf("result%d", i)
	}
	if len(fi.Returns) > 0 {
		returnNames += ":="
	}
	return returnNames
}

func (fi FuncInfo) PrintReceiverCtor() string {
	ctor := ""
	if fi.ReceiverType == "" {
		return ctor
	}
	return fmt.Sprintf("var receiver %s", fi.ReceiverShort)
}

func (fi FuncInfo) PrintCall() string {
	if fi.ReceiverType == "" {
		return fi.Name
	}
	return fmt.Sprintf("receiver.%s", fi.Name)
}

func (fi FuncInfo) PrintDefaultExpects() string {
	const tmpl = `
		var %s %s
		if !reflect.DeepEqual(%s, %s) {
			t.Errorf("Expected %%v, got %%v", %s, %s)
		}
	`
	var expects strings.Builder
	for i, retType := range fi.Returns {
		expectName := fmt.Sprintf("expect%d", i)
		// expcectType := retType.TypeName
		expectType := retType.ZeroValue
		resultName := fmt.Sprintf("result%d", i)
		s := fmt.Sprintf(tmpl, expectName, expectType, resultName, expectName, expectName, resultName)
		expects.WriteString(s)
	}
	return expects.String()
}

func (fi FuncInfo) PrintDefaultVarArgs() string {
	var sb strings.Builder
	for _, param := range fi.Params {
		def := fmt.Sprintf("var %s %s\n", param.Name, param.ZeroValue)
		sb.WriteString(def)
	}
	return sb.String()
}

func (fi FuncInfo) PrintArgsAsStructFields() string {
	var sb strings.Builder
	// add test run name to struct
	field := fmt.Sprintf("%s %s\n", "name", "string")
	sb.WriteString(field)
	for _, param := range fi.Params {
		field := fmt.Sprintf("%s %s\n", param.Name, param.ZeroValue)
		sb.WriteString(field)
	}
	return sb.String()
}

func (fi FuncInfo) PrintTableArgs() string {
	var args strings.Builder
	written := 0
	for _, arg := range fi.Params {
		if written > 0 {
			args.WriteString(", ")
		}
		args.WriteString("tt." + arg.Name)
		written++
	}
	return args.String()
}
