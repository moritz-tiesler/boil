package main

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type ParamInfo struct {
	Name         string
	TypeName     string // Fully qualified type string (e.g., "string", "io.Reader")
	ZeroValue    string // Go code for the zero value (e.g., `""`, `0`, `nil`)
	ImportedFrom string
}

func NewParamInfo(field *ast.Field, pkg *packages.Package) []*ParamInfo {
	pInfos := []*ParamInfo{}

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
			pInfos = append(pInfos, &ParamInfo{TypeName: returnTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom})
		} else {
			for _, name := range field.Names {
				pInfos = append(
					pInfos,
					&ParamInfo{Name: name.Name, TypeName: returnTypeName, ZeroValue: zeroVal, ImportedFrom: importedFrom},
				)
			}
		}
	} else {
		pInfos = append(pInfos, &ParamInfo{Name: "UNKNOWN_TYPE", TypeName: "UNKNOWN_TYPE"})
	}
	return pInfos
}
