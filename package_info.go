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
	ModuleName    string
	PackageIsMain bool
	Path          string
	Name          string
	Imports       map[string]struct{}
	Funcs         []FuncInfo
	goRoot        string
}

func NewPackageInfo(pkgPath string) PackageInfo {
	goroot, err := getGOROOT()
	if err != nil {
		panic(err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			// packages.LoadFiles | // <--- REMOVE THIS LINE
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
		goRoot:  goroot,
		Imports: make(map[string]struct{}),
	}

	for _, pkg := range pkgs {
		if pkg.Module != nil {
			pkgInfo.ModuleName = pkg.Module.Path
			pkgInfo.PackageIsMain = pkg.Module.Main
		}
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

	// for _, param := range fi.Params {
	// 	pi.Imports[param.Pkg.Path()] = &param.Pkg
	// }
	// for _, ret := range fi.Returns {
	// 	pi.Imports[ret.Pkg.Path()] = &ret.Pkg
	// }
	for k, _ := range fi.RequiredImports {
		pi.Imports[k] = struct{}{}
	}
	pi.Funcs = append(pi.Funcs, *fi)
}

func (pi PackageInfo) PrintImports() string {
	var sb strings.Builder
	for name, _ := range pi.Imports {
		var importStr string
		// if isStandardLibrary(pkg, pi.goRoot) {
		// 	importStr = strings.TrimPrefix(pi.goRoot, name)
		// } else {
		// 	importStr = name
		// }
		importStr = name
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
