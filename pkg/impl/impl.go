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

package impl

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

const errDeclPattern = "+thaterror:error="
const multiLineErrDeclStart = "+thaterror:error:start"
const multilineErrDeclEnd = "+thaterror:error:stop"

// Pkg generates error related functions for a pkg
func Pkg(path string, pkgName string, types []*UnintializedErrorType, outputFileName string) {
	f := jen.NewFile(pkgName)

	for _, typ := range types {
		generateTyp(f, typ)
	}

	fmt.Printf("%#v", f)
}

func generateTyp(f *jen.File, typ *UnintializedErrorType) {
	decl, ok := typ.Node.(*ast.GenDecl)
	if !ok {
		log.Fatal("node is not a *ast.GenDecl")
	}

	if decl.Tok != token.TYPE {
		log.Fatal("node.Tok is not token.TYPE")
	}

	errType, ok := decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		log.Fatal("node is not a *ast.TypeSpec")
	}

	errImpl := &Error{
		TypeName:      errType.Name.Name,
		ErrorTemplate: "",
	}

	multiLineTemplate := false
	for _, comment := range typ.Comments {
		comment := comment.Text
		if multiLineTemplate {
			errImpl.ErrorTemplate += strings.Trim(comment, " /") + "\n"
		} else if index := strings.Index(comment, errDeclPattern); index != -1 {
			tmplContent := comment[index+len(errDeclPattern):]
			errImpl.ErrorTemplate = tmplContent
		} else if strings.Contains(comment, multiLineErrDeclStart) {
			multiLineTemplate = true
		} else if strings.Contains(comment, multilineErrDeclEnd) {
			multiLineTemplate = false
		}
	}

	errImpl.impl(f)
}

func (e *Error) impl(f *jen.File) {
	generateErrorFunc(f, e.ErrorTemplate, e.TypeName)
}

func generateErrorFunc(f *jen.File, tmplContent string, typeName string) {
	ptrTypName := "*" + typeName

	tmplName := typeName + "ErrorTmpl"
	f.Const().Defs(
		jen.Id(tmplName).Op("=").
			Qual("template", "Must").Call(
			jen.Qual("template", "New").Call(jen.Lit(tmplName)).
				Dot("Parse").Call(jen.Lit(tmplContent))),
	)

	f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Error").Params().Error().Block(
		jen.Id("buf").Op(":=").Id("new").Call(jen.Qual("bytes", "Buffer")),
		jen.Id("tmplErr").Op(":=").Id(tmplName).Dot("Execute").Call(
			jen.Id("buf"),
			jen.Id("err"),
		),
		jen.If(jen.Id("tmplErr").Op("!=").Nil()).Block(
			jen.Panic(jen.Lit("fail to render error template")),
		),
		jen.Return(jen.Id("buf").Dot("String").Call()),
	)
}
