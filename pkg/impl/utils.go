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
	"strings"

	"github.com/dave/jennifer/jen"
)

func (e *Error) allWrapTypeCaseFunc() func(g *jen.Group) {
	return func(g *jen.Group) {
		for _, typ := range e.WrapTypes {
			sections := strings.Split(typ, "\".")
			if len(sections) > 1 {
				ptr := false

				pkg := sections[0][:]
				if pkg[0] == '*' {
					ptr = true
					pkg = pkg[1:]
				}
				pkg = pkg[1:]
				typ := sections[1]

				if ptr {
					g.Op("*").Qual(pkg, typ)
				} else {
					g.Qual(pkg, typ)
				}
			} else {
				g.Id(typ)
			}
		}
	}
}
