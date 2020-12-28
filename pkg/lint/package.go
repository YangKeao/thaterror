// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
)

type Package struct {
	fset  *token.FileSet
	files map[string]*ast.File

	TypesPkg  *types.Package
	TypesInfo *types.Info
}

func LintPackage(filenames []string) {
	if len(filenames) == 0 {
		return
	}

	pkg := &Package{
		fset:  token.NewFileSet(),
		files: make(map[string]*ast.File),
	}
	for _, f := range filenames {
		astFile, err := parser.ParseFile(pkg.fset, f, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		pkg.files[f] = astFile
	}

	config := &types.Config{
		// By setting a no-op error reporter, the type checker does as much work as possible.
		Error:       func(error) {},
		Importer:    importer.ForCompiler(pkg.fset, "source", nil),
		FakeImportC: true,
	}

	info := &types.Info{
		Types:  make(map[ast.Expr]types.TypeAndValue),
		Defs:   make(map[*ast.Ident]types.Object),
		Uses:   make(map[*ast.Ident]types.Object),
		Scopes: make(map[ast.Node]*types.Scope),
	}

	var files []*ast.File
	for _, f := range pkg.files {
		files = append(files, f)
		break
	}
	typesPkg, err := config.Check(files[0].Name.Name, pkg.fset, files, info)
	if err != nil {
		// log.Fatal(err)
	}

	pkg.TypesPkg = typesPkg
	pkg.TypesInfo = info

	for path, file := range pkg.files {
		pkg.lintFile(path, file)
	}
}

func (pkg *Package) lintFile(path string, file *ast.File) {
	fmt.Println("lint file " + path)
	ast.Walk(&Walker{
		Package: pkg,
		Path:    path,
	}, file)
}

type Walker struct {
	*Package
	Path string
}

func (walker *Walker) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		walker.checkFuncResult(node)
	}

	return walker
}

func (walker *Walker) checkFuncResult(fn *ast.FuncDecl) {
	rawType := walker.TypesInfo.TypeOf(fn.Name)

	fnType, ok := rawType.Underlying().(*types.Signature)
	if !ok {
		return
	}

	retTypes := fnType.Results()

	for i := 0; i < retTypes.Len(); i++ {
		nt, ok := retTypes.At(i).Type().(*types.Named)
		if ok && isTypeError(nt) {
			pos := fn.Pos()
			log.Fatalf("%+v at %s:%+v shouldn't use `error` as a return type", fn.Name.Name, walker.Path, walker.fset.Position(pos))
		}
	}
}
