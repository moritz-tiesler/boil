package main

import "go/types"

type ParamInfo struct {
	Name           string
	TypeName       string
	TypeNameSimple string
	Pkg            types.Package
}
