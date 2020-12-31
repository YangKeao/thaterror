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

package filemanager

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dave/jennifer/jen"
)

var fileMap = make(map[string]*jen.File)

// OpenFile opens a file for generating
func OpenFile(path string, pkgName string) (*jen.File, error) {
	path = Path(path)
	f, ok := fileMap[path]
	if ok {
		return f, nil
	}

	if len(pkgName) == 0 {
		log.Fatal("pkgName cannot be empty")
	}
	jenFile := jen.NewFile(pkgName)
	fileMap[path] = jenFile
	return jenFile, nil
}

// Path returns the generated path of file
func Path(path string) string {
	path = filepath.Dir(path)

	return path + "/zz_generated.thaterror.go"
}

func Render() error {
	for path, f := range fileMap {
		log.Printf("render file %s", path)
		file, err := os.Create(path)
		if err != nil {
			log.Fatalf("fail to open file %s", err)
		}

		err = f.Render(file)
		if err != nil {
			log.Fatalf("fail to render %s", err)
		}
	}

	return nil
}
