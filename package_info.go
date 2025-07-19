package main

import (
	"fmt"
	"go/ast"
	"maps"
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
		panic(fmt.Sprintf("expected one packaged in %s", pkgPath))
	}

	if packages.PrintErrors(pkgs) > 0 {
		panic("packages contains errors")
	}

	pkgInfo := PackageInfo{
		goRoot:  goroot,
		Imports: make(map[string]struct{}),
	}

	pkg := pkgs[0]

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
	maps.Insert(pi.Imports, maps.All(fi.RequiredImports))
	pi.Funcs = append(pi.Funcs, *fi)
}

func (pi PackageInfo) PrintImports() string {
	var sb strings.Builder
	for name := range pi.Imports {
		var importStr string
		importStr = name
		sb.WriteString(fmt.Sprintf("\"%s\"\n", importStr))
	}
	return sb.String()
}

func getGOROOT() (string, error) {
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
