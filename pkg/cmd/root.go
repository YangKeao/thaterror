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
	"os"
	"strings"

	"github.com/YangKeao/thaterror/pkg/impl"

	zglob "github.com/mattn/go-zglob"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "thaterror",
	Short: "thaterror is a code generator for error handling",
	Run:   walkFiles,
}

var (
	path           string
	filter         string
	outputFileName string
)

func init() {
	RootCmd.PersistentFlags().StringVar(&path, "path", ".", "the root path of your project")
	RootCmd.PersistentFlags().StringVar(&filter, "filter", "**/error.go", "only files matching the pattern will be walked")
	RootCmd.PersistentFlags().StringVarP(&outputFileName, "output", "o", "zz_generated.thaterror.go", "the output filename of generated file")
}

func walkFiles(cmd *cobra.Command, args []string) {
	err := os.Chdir(path)
	if err != nil {
		log.Fatal(err)
	}

	files, err := zglob.Glob(filter)
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
			importMap[impt.Name.Name] = impt.Path.Value
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
}
