# thaterror

a code generator for generating good error handling pattern for golang 👿

## Demo

```go
// TraceError is the error in tracing a process
// +thaterror:transparent
// +thaterror:wrap=*PtraceError
// +thaterror:wrap=*WaitPidError
// +thaterror:wrap="pkg/commonerror".*IOError
// +thaterror:wrap="pkg/commonerror".*ParseIntError
// +thaterror:wrap="pkg/mapreader".*Error
type TraceError struct {
	Err error
}

// PtraceError represents an error returned by ptrace syscall
// +thaterror:error=fail to ptrace({{.Operation}}) on process {{.Tid}}
type PtraceError struct {
	Err       error
	Tid       int
	Operation int
}

// WaitPidError means the waitpid syscall returns an error
// +thaterror:error=waitpid failed
type WaitPidError struct{}

func Trace(pid int) (*TracedProgram, *TraceError) {
    return nil, TraceErrorWrap(&commonerror.IOError{
        Err: err,
    })
}
```

As shown in the demo, we can add a lot of `+thaterror:wrap` annotation for the
struct `TraceError`, and the code generator will generate `TraceErrorWrap`
function. Only the listed types (in the annotation) can be passed into the
`TraceErrorWrap` function. Passing other types of variables into the
`TraceErrorWrap` function will result in a compiling error.

`+thaterror:wrap` is the most important annotation. The others like
`+thaterror:error` and `+thaterror:transparent` are just helpers to implement
`error` interface. With `+thaterror:error`, the error message is generated with
the specified template, the `.` of which is the struct it self.

## Implementation

The implementation for `+thaterror:error` and `+thaterror:transparent` is quite
straightforward

## Motivation

Error handling in `Go` is always a problem. There are tons of articles and
libraries about it. In 2020, this problem is still far from being solved, though
the
[`xerrors`](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
(based on Go 2 proposal) seems to be a little step towards the answer. However,
from the very begining, our demands for error handling is quite simple: we just
need to know all possible errors the function will return, and handle them
differently.

To prove this statement, I will show you an example. For the function `New` in the 
package `"github.com/hashicorp/golang-lru"`, the function signature is:

```go
func New(size int) (*Cache, error)
```

It gives an error. But how could creating an lru result in an error (without
considering the `memory allocation failed` error)? After exploring the source code of 
this function, you will know the only error is that: `"Must provide a positive size"`.
As the user of this function, I'm confident enough to ignore the error now 😃
(though printing it out could be a better idea), as I will call it with a
constant positive integer. 

However, exploring the source code may take a lot of time. If the author of the
packages writes down all the possible errors in the document of the function,
this problem will be solved, but little programmers will do it well (which has
been proved in the `Go` ecosystem today 👿).

