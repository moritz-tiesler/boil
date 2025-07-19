package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

type FuncInfo struct {
	Name               string
	IsExported         bool
	ReceiverType       string // Fully qualified type string
	ReceiverTypeSimple string
	Params             []*ParamInfo
	Returns            []*ParamInfo // For Go, return values are also like parameters
	HasTypeParams      bool
	RequiredImports    map[string]struct{}
}

func NewFunctionInfo(funcDecl *ast.FuncDecl, pkg *packages.Package) *FuncInfo {

	fInfo := FuncInfo{
		Name:            funcDecl.Name.Name,
		RequiredImports: map[string]struct{}{},
	}

	fnType := pkg.TypesInfo.TypeOf(funcDecl.Name) // returns types.Signature

	// TODO: this could be an easier way to access the types
	if sig, ok := fnType.(*types.Signature); ok {
		if sig.Recv() != nil {
			recVar := sig.Recv()
			fInfo.ReceiverType = getQualifiedTypeName(recVar.Type())
			fInfo.ReceiverTypeSimple = getSimplifiedTypeName(recVar.Type(), pkg.Types)
			if p := getOriginatingPackage(sig.Recv().Type(), pkg.Types); p != nil {
				fInfo.RequiredImports[p.Path()] = struct{}{}
			}
		}

		i := 0
		if params := sig.Params(); params != nil {
			for paramVar := range params.Variables() { // <--- Changed here: use Variables() and range loop
				paramTypeName := getQualifiedTypeName(paramVar.Type())

				paramTypeNameSimple := getSimplifiedTypeName(paramVar.Type(), pkg.Types)
				paramName := paramVar.Name()
				paramVar.Origin()
				if paramName == "" {
					paramName = fmt.Sprintf("param%d", i) // Provide a default name for unnamed parameters
				}
				if p := getOriginatingPackage(paramVar.Type(), pkg.Types); p != nil {
					fInfo.RequiredImports[p.Path()] = struct{}{}
				}
				fInfo.Params = append(fInfo.Params, &ParamInfo{
					Name:           paramName,
					TypeName:       paramTypeName,
					TypeNameSimple: paramTypeNameSimple,
					Pkg:            *paramVar.Pkg(),
				})
				i++
			}
		}

		i = 0
		if results := sig.Results(); results != nil {
			for returnVar := range results.Variables() { // <--- Changed here: use Variables() and range loop
				returnTypeName := getQualifiedTypeName(returnVar.Type())
				returnTypeNameSimple := getSimplifiedTypeName(returnVar.Type(), pkg.Types)

				returnName := returnVar.Name()
				if returnName == "" {
					returnName = fmt.Sprintf("result%d", i) // Provide a default name for unnamed results
				}
				if p := getOriginatingPackage(returnVar.Type(), pkg.Types); p != nil {
					fInfo.RequiredImports[p.Path()] = struct{}{}
				}
				fInfo.Returns = append(fInfo.Returns, &ParamInfo{
					Name:           returnName,
					TypeName:       returnTypeName,
					TypeNameSimple: returnTypeNameSimple,
					Pkg:            *returnVar.Pkg(),
				})
			}
		}

	}
	return &fInfo
}

func (fi FuncInfo) GetRequiredImports() []string {
	var required []string
	for _, p := range fi.Params {
		required = append(required, p.Pkg.Path())
	}

	for _, p := range fi.Returns {
		required = append(required, p.Pkg.Path())
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
	return fmt.Sprintf("var receiver %s", fi.ReceiverTypeSimple)
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
		expectType := retType.TypeNameSimple
		resultName := fmt.Sprintf("result%d", i)
		s := fmt.Sprintf(tmpl, expectName, expectType, resultName, expectName, expectName, resultName)
		expects.WriteString(s)
	}
	return expects.String()
}

func (fi FuncInfo) PrintDefaultVarArgs() string {
	var sb strings.Builder
	for _, param := range fi.Params {
		def := fmt.Sprintf("var %s %s\n", param.Name, param.TypeNameSimple)
		sb.WriteString(def)
	}
	return sb.String()
}

func (fi FuncInfo) PrintArgsAsStructFields() string {
	var sb strings.Builder
	for _, param := range fi.Params {
		field := fmt.Sprintf("%s %s\n", param.Name, param.TypeNameSimple)
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

func getOriginatingPackage(typ types.Type, currentPkgTypes *types.Package) *types.Package {
	switch t := typ.(type) {
	case *types.Named:
		// If it's a named type and belongs to a package different from the current one
		if t.Obj().Pkg() != nil && t.Obj().Pkg() != currentPkgTypes {
			return t.Obj().Pkg()
		}
	case *types.Pointer:
		return getOriginatingPackage(t.Elem(), currentPkgTypes)
	case *types.Slice:
		return getOriginatingPackage(t.Elem(), currentPkgTypes)
	case *types.Array:
		return getOriginatingPackage(t.Elem(), currentPkgTypes)
	case *types.Map:
		if pkg := getOriginatingPackage(t.Key(), currentPkgTypes); pkg != nil {
			return pkg
		}
		if pkg := getOriginatingPackage(t.Elem(), currentPkgTypes); pkg != nil {
			return pkg
		}
	case *types.Chan:
		return getOriginatingPackage(t.Elem(), currentPkgTypes)
	case *types.Signature:
		// For function types (e.g., if a param is a function), check its parameters and results
		for i := 0; i < t.Params().Len(); i++ {
			if pkg := getOriginatingPackage(t.Params().At(i).Type(), currentPkgTypes); pkg != nil {
				return pkg
			}
		}
		for i := 0; i < t.Results().Len(); i++ {
			if pkg := getOriginatingPackage(t.Results().At(i).Type(), currentPkgTypes); pkg != nil {
				return pkg
			}
		}
		// For basic types, structs (unless they are underlying a Named type already handled),
		// and interfaces, they don't directly point to an external originating package here.
	}
	return nil
}
