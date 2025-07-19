package main

import (
	"golang.org/x/tools/go/packages"
)

type ParamInfo struct {
	Name           string
	TypeName       string // Fully qualified type string (e.g., "string", "io.Reader")
	TypeNameSimple string // Fully qualified type string (e.g., "string", "io.Reader")
	Pkg            packages.Package
}
