package main

import (
	"fmt"
	"go/ast"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

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

func NewPackageInfo(pkgPath string) PackageInfo {
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
