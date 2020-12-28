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
	"bufio"
	"go/ast"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

const errDeclPattern = "+thaterror:error="
const errWrapPattern = "+thaterror:wrap="
const multiLineErrDeclStart = "+thaterror:error:start"
const multilineErrDeclEnd = "+thaterror:error:end"
const transparentPattern = "+thaterror:transparent"

// Pkg generates error related functions for a pkg
func Pkg(path string, pkgName string, importMap map[string]string, types []*UnintializedErrorType, outputFileName string) {
	f := jen.NewFile(pkgName)

	for _, typ := range types {
		generateTyp(f, typ)
	}

	path = filepath.Dir(path)

	file, err := os.Create(path + "/zz_generated.thaterror.go")
	if err != nil {
		log.Fatalf("fail to open file %s", err)
	}

	w := bufio.NewWriter(file)
	err = f.Render(w)
	if err != nil {
		log.Fatalf("fail to render %s", err)
	}
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
		Transparent:   false,
		WrapTypes:     []string{},
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
		} else if strings.Contains(comment, transparentPattern) {
			errImpl.Transparent = true
		} else if index := strings.Index(comment, errWrapPattern); index != -1 {
			fromType := comment[index+len(errWrapPattern):]
			errImpl.WrapTypes = append(errImpl.WrapTypes, fromType)
		}
	}

	errImpl.impl(f)
}

func (e *Error) impl(f *jen.File) {
	if e.Transparent {
		e.generateTransparentErrorFunc(f)
	} else {
		e.generateTemplateErrorFunc(f)
	}

	e.generateErrorWrap(f)
	e.generateErrorUnwrap(f)
}

func (e *Error) generateTemplateErrorFunc(f *jen.File) {
	ptrTypName := "*" + e.TypeName

	tmplName := e.TypeName + "ErrorTmpl"
	f.Var().Defs(
		jen.Id(tmplName).Op("=").
			Qual("text/template", "Must").Call(
			jen.Qual("text/template", "New").Call(jen.Lit(tmplName)).
				Dot("Parse").Call(jen.Lit(e.ErrorTemplate))),
	)

	f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Error").Params().String().Block(
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

func (e *Error) generateErrorWrap(f *jen.File) {
	ptrTypName := "*" + e.TypeName

	if len(e.WrapTypes) == 0 {
		return
	}

	f.Func().Id(strcase.ToLowerCamel(e.TypeName) + "Wrap").Params(jen.Id("err").Error()).Id(ptrTypName).Block(
		jen.Switch(jen.Id("err").Assert(jen.Type())).Block(
			jen.CaseFunc(
				e.allWrapTypeCaseFunc(),
			).Block(
				jen.Return(
					jen.Op("&").Id(e.TypeName).Values(jen.Dict{
						jen.Id("Err"): jen.Id("err"),
					}),
				),
			),
			jen.Default().Block(
				jen.Panic(jen.Lit("unexpected error type")),
			),
		),
	)
}

func (e *Error) generateErrorUnwrap(f *jen.File) {
	ptrTypName := "*" + e.TypeName

	fun := f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Unwrap").Params().Error()
	if len(e.WrapTypes) == 0 {
		fun.Block(
			jen.Return(jen.Nil()),
		)
		return
	}

	fun.Block(
		jen.Switch(jen.Id("err").Dot("Err").Assert(jen.Type())).Block(
			jen.CaseFunc(
				e.allWrapTypeCaseFunc(),
			).Block(
				jen.Return(
					jen.Id("err").Dot("Err"),
				),
			),
			jen.Default().Block(
				jen.Panic(jen.Lit("unexpected error type")),
			),
		),
	)
}

func (e *Error) generateTransparentErrorFunc(f *jen.File) {
	ptrTypName := "*" + e.TypeName
	f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Error").Params().Error().Block(
		jen.Return(jen.Id("err").Dot("Error").Call()),
	)
}
