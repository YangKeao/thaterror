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
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"github.com/YangKeao/thaterror/pkg/filemanager"
	"github.com/dave/jennifer/jen"
)

const errDeclPattern = "+thaterror:error="
const errWrapPattern = "+thaterror:wrap="
const multiLineErrDeclStart = "+thaterror:error:start"
const multilineErrDeclEnd = "+thaterror:error:end"
const transparentPattern = "+thaterror:transparent"

// Pkg generates error related functions for a pkg
func Pkg(path string, pkgName string, importMap map[string]string, types []*UnintializedErrorType, outputFileName string) {
	f, err := filemanager.OpenFile(path, pkgName)
	if err != nil {
		log.Fatal(err)
	}

	for _, typ := range types {
		generateTyp(f, typ, filepath.Dir(path))
	}
}

func generateTyp(f *jen.File, typ *UnintializedErrorType, path string) {
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
		Path:          path,
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

	if len(e.WrapTypes) > 0 {
		e.generateErrorWrapInterface(f)
		e.generateErrorWrap(f)
		e.generateErrorUnwrap(f)
	}
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

func (e *Error) generateErrorWrapInterface(f *jen.File) {
	f.Type().Id(e.getUnionName()).Interface(
		jen.Id(e.getUnionFunctionName()).Params(),
		jen.Error(),
	)

	for _, typ := range e.WrapTypes {
		var err error
		jenFile := f
		path, typ := e.splitErrorWrapType(typ)
		if len(path) != 0 {
			// TODO: make it more flexible to get the pkg name but not the last element of filepath
			jenFile, err = filemanager.OpenFile(path+"/error.go", filepath.Base(path))
			if err != nil {
				log.Fatal(err)
			}
		}

		jenFile.Func().Params(
			jen.Id("err").Id(typ),
		).Id(e.getUnionFunctionName()).Params().Block()
	}
}

func (e *Error) generateErrorWrap(f *jen.File) {
	ptrTypName := "*" + e.TypeName

	f.Func().Id(e.TypeName + "Wrap").Params(jen.Id("err").Id(e.getUnionName())).Id(ptrTypName).Block(
		jen.Return(
			jen.Op("&").Id(e.TypeName).Values(jen.Dict{
				jen.Id("Err"): jen.Id("err"),
			}),
		),
	)
}

func (e *Error) generateErrorUnwrap(f *jen.File) {
	ptrTypName := "*" + e.TypeName

	fun := f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Unwrap").Params().Id(e.getUnionName())

	fun.Block(
		jen.Return(
			jen.Id("err").Dot("Err").Assert(jen.Id(e.getUnionName())),
		),
	)
}

func (e *Error) generateTransparentErrorFunc(f *jen.File) {
	ptrTypName := "*" + e.TypeName
	f.Func().Params(
		jen.Id("err").Id(ptrTypName),
	).Id("Error").Params().String().Block(
		jen.Return(jen.Id("err").Dot("Error").Call()),
	)
}
