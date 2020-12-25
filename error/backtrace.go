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

package error

import (
	"fmt"
	"runtime"
	"github.com/pkg/errors"
)

// Backtrace represents a call stack
type Backtrace []uintptr

// Callers gets a call stack
func Callers() *Backtrace {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st Backtrace = pcs[0:n]
	return &st
}

// Format formats a backtrace
func (s *Backtrace) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errors.Frame(pc)
				fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}
