package main

import "go/types"

type ParamInfo struct {
	Name           string
	TypeName       string // Fully qualified type string (e.g., "string", "io.Reader")
	TypeNameSimple string // Fully qualified type string (e.g., "string", "io.Reader")
	Pkg            types.Package
}
