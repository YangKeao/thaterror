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

	"github.com/iancoleman/strcase"
)

func (e *Error) getUnionName() string {
	return e.TypeName + "Wrap" + "Union"
}

func (e *Error) splitErrorWrapType(wrapType string) (string, string) {
	sections := strings.Split(wrapType, "\".")
	if len(sections) > 1 {
		path := sections[0][1:]
		typ := sections[1]

		return path, typ
	} else {
		return "", sections[0]
	}
}

func (e *Error) getUnionFunctionName() string {
	return strcase.ToCamel(strings.ReplaceAll(e.Path, "/", "") + e.TypeName)
}
