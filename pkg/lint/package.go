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
	"go/parser"
	"go/token"
	"log"
	"strings"
)

const ignorePattern = "+thaterror:ignore"

type Package struct {
	fset  *token.FileSet
	files map[string]*ast.File
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

		if astFile.Doc != nil {
			for _, comment := range astFile.Doc.List {
				if strings.Contains(comment.Text, ignorePattern) {
					log.Printf("ignore pkg of file %s", f)
					return
				}
			}
		}
	}

	var files []*ast.File
	for _, f := range pkg.files {
		files = append(files, f)
		break
	}

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
	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			if strings.Contains(comment.Text, ignorePattern) {
				return
			}
		}
	}

	retTypes := fn.Type.Results
	if retTypes == nil {
		return
	}

	for _, field := range retTypes.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "error" {
			log.Fatalf("%+v at %+v shouldn't use `error` as a return type", fn.Name.Name, walker.fset.Position(fn.Pos()))
		}
	}
}
