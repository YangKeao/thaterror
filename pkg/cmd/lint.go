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
	"log"
	"sync"

	"github.com/YangKeao/thaterror/pkg/lint"
	"github.com/mgechev/dots"

	zglob "github.com/mattn/go-zglob"
	"github.com/spf13/cobra"
)

// LintCmd is the cobra command to lint error handling codes
var LintCmd = &cobra.Command{
	Use:   "lint",
	Short: "lint codes for error handling",
	Run:   lintCmd,
}

var (
	lintFiles []string
	lintSkips []string
)

func init() {
	LintCmd.PersistentFlags().StringSliceVar(&lintFiles, "files", []string{"./..."}, "only files matching the pattern will be walked")
	LintCmd.PersistentFlags().StringSliceVar(&lintSkips, "skips", []string{"**/zz_generated.*.go", "**/*.pb.go"}, "skip the files match this pattern")
}

func lintCmd(cmd *cobra.Command, args []string) {
	skipFiles := []string{}
	for _, skipGlob := range lintSkips {
		fs, err := zglob.Glob(skipGlob)
		if err != nil {
			log.Fatal(err)
		}

		skipFiles = append(skipFiles, fs...)
	}

	packages, err := dots.ResolvePackages(lintFiles, skipFiles)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, pkg := range packages {
		pkg := pkg

		wg.Add(1)
		go func() {
			defer wg.Done()
			lint.LintPackage(pkg)
		}()
	}

	wg.Wait()
}
