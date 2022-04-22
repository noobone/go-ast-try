package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"strings"

	"golang.org/x/tools/go/packages"
)

type packageLoader struct {
	*packages.Config
	PkgFilter []func(pkgPath string) bool
}

func NewPackageLoader() *packageLoader {
	return &packageLoader{}
}

func (pl *packageLoader) RecursionParsePkg(pkg *packages.Package, pkgName string, pkgMap map[string]*packages.Package) {
	var returnTag bool
	for _, filter := range pl.PkgFilter {
		if !filter(pkg.PkgPath) {
			returnTag = true
		}
	}
	if returnTag {
		fmt.Println(pkg.PkgPath)
		return
	}

	if _, ok := pkgMap[pkg.ID]; !ok {
		pkgMap[pkg.ID] = pkg
	} else {
		fmt.Println(pkg.PkgPath)
		return
	}

	for _, imp := range pkg.Imports {
		pl.RecursionParsePkg(imp, imp.ID, pkgMap)
	}
}

func main() {
	PACKAGE_NAME := "github.com/noobone/go-ast-try"

	loadMode := packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles |
		packages.NeedImports |
		packages.NeedDeps |
		packages.NeedExportsFile |
		packages.NeedTypes |
		packages.NeedTypesInfo |
		packages.NeedTypesSizes |
		packages.NeedSyntax |
		packages.NeedModule

	cfg := &packages.Config{
		Mode:       loadMode,
		BuildFlags: build.Default.BuildTags,
		Dir:        "",
	}

	pl := NewPackageLoader()
	pl.Config = cfg
	pl.PkgFilter = []func(pkgPath string) bool{
		func(pkgPath string) bool {
			if strings.HasPrefix(pkgPath, PACKAGE_NAME) {
				return true
			}
			return false
		},
	}
	pkgMap := map[string]*packages.Package{}

	pkgs, err := packages.Load(pl.Config, PACKAGE_NAME)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, pkg := range pkgs {
		pl.RecursionParsePkg(pkg, PACKAGE_NAME, pkgMap)
	}

	for pkgID, pkg := range pkgMap {
		fmt.Println("package name:", pkgID)
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if parsedFuncDecl, ok := n.(*ast.FuncDecl); ok {
					ast.Inspect(parsedFuncDecl.Body, func(n ast.Node) bool {
						if parsedSelectorExpr, ok := n.(*ast.SelectorExpr); ok {
							if nestedParsedSelectorExpr, ok := parsedSelectorExpr.X.(*ast.SelectorExpr); ok {
								if parsedIdent, ok := nestedParsedSelectorExpr.X.(*ast.Ident); ok {
									if parsedIdent.Name == "a" && nestedParsedSelectorExpr.Sel.Name == "b" && parsedSelectorExpr.Sel.Name == "c" {
										parsedIdentPosition := pkg.Fset.Position(parsedIdent.NamePos)
										nestedParsedSelPosition := pkg.Fset.Position(nestedParsedSelectorExpr.Sel.NamePos)
										parsedSelPosition := pkg.Fset.Position(parsedSelectorExpr.Sel.NamePos)
										fmt.Println("a", parsedIdentPosition.Filename, parsedFuncDecl.Name.Name, parsedIdentPosition.Line)
										fmt.Println("b", nestedParsedSelPosition.Filename, parsedFuncDecl.Name.Name, nestedParsedSelPosition.Line)
										fmt.Println("c", parsedSelPosition.Filename, parsedFuncDecl.Name.Name, parsedSelPosition.Line)
										fmt.Println()
									}
								}
							}
						}
						return true
					})
				}
				return true
			})
		}
	}
}
