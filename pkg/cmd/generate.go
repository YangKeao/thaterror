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

package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/YangKeao/thaterror/pkg/filemanager"
	"github.com/YangKeao/thaterror/pkg/impl"

	zglob "github.com/mattn/go-zglob"
	"github.com/spf13/cobra"
)

// GenerateCmd is the cobra command to generate error handling codes
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate codes for error handling",
	Run:   generateCmd,
}

var (
	generateFilter string
	outputFileName string
)

func init() {
	GenerateCmd.PersistentFlags().StringVar(&generateFilter, "filter", "**/error.go", "only files matching the pattern will be walked")
	GenerateCmd.PersistentFlags().StringVarP(&outputFileName, "output", "o", "zz_generated.thaterror.go", "the output filename of generated file")
}

func generateCmd(cmd *cobra.Command, args []string) {
	files, err := zglob.Glob(generateFilter)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		log.Printf("iterating file: %s\n", file)

		fset := token.NewFileSet()
		goFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		importMap := make(map[string]string)
		for _, impt := range goFile.Imports {
			if impt.Name != nil {
				importMap[impt.Name.Name] = impt.Path.Value
			} else {
				// TODO: get the last section of Path
				importMap[impt.Path.Value] = impt.Path.Value
			}
		}

		types := []*impl.UnintializedErrorType{}
		cmap := ast.NewCommentMap(fset, goFile, goFile.Comments)
		for node, commentGroups := range cmap {
			comments := []*ast.Comment{}
			related := false
			for _, commentGroup := range commentGroups {
				for _, comment := range commentGroup.List {
					if strings.Contains(comment.Text, "+thaterror") {
						related = true
					}
					comments = append(comments, comment)
				}
			}

			if related {
				errType := &impl.UnintializedErrorType{
					Node:     node,
					Comments: comments,
				}
				types = append(types, errType)
			}
		}

		impl.Pkg(file, goFile.Name.Name, importMap, types, outputFileName)
	}

	err = filemanager.Render()
	if err != nil {
		log.Fatal(err)
	}
}
