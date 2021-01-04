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

## Motivation

Error handling in `Go` is always a problem. There are tons of articles and
libraries about it. In 2020, this problem is still far from being solved, though
the
[`xerrors`](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
(based on Go 2 proposal) seems to be a little step towards the answer. However,
from the very begining, our demands for error handling is quite simple: we just
need to know all possible errors the function will return, and handle them
differently.

To prove this statement, I will show you an example. For the function `New` in
the package `"github.com/hashicorp/golang-lru"`, the function signature is:

```go
func New(size int) (*Cache, error)
```

It gives an error. But how could creating an lru result in an error (without
considering the `memory allocation failed` error)? After exploring the source
code of this function, you will know the only error is that: `"Must provide a
positive size"`. As the user of this function, I'm confident enough to ignore
the error now 😃 (though printing it out could be a better idea), as I will call
it with a constant positive integer. 

However, exploring the source code may take a lot of time. If the author of the
packages writes down all the possible errors in the document of the function,
this problem will be solved, but little programmers will do it well (which has
been proved in the `Go` ecosystem today 👿).

Let's take a look on how other programing languages solve the problem.

### Java

The type of exceptions can be declared at the end of a function signature.

```java
import java.io.*;
public class className {
    public void methodName() throws ExceptionOne, ExceptionTwo {
        possibleExceptionTwo();
        throw new ExceptionOne();
    }
}
```

So that the function signature will tell you all possible (checked) exceptions.

### Rust

Rust has a powerful
[`enum`](https://doc.rust-lang.org/book/ch06-01-defining-an-enum.html), and it
can be used as a generic parameter:

```rust
enum Error {
    ErrorOne,
    ErrorTwo,
    ErrorThree(String),
    ErrorFour(u8, u8, u8, u8),
}

type Result<T> = std::result::Result<T, Error>

fn some() -> Result<()>
```

The definition of `Result` in `std` has nothing mystery: 

```rust
enum Result<T, E> {
   Ok(T),
   Err(E),
}
```

The programmers can use `match`, `if let`, etc, to assert whether it's an `Err`
or `Ok`. A `?` operator has been added to simplify the error assertion and type
convertion
([document](https://doc.rust-lang.org/edition-guide/rust-2018/error-handling-and-panics/the-question-mark-operator-for-easier-error-handling.html)).

The approach is very similar with `Java`. The only difference is that the
"union" of error type has to be declared in `Rust`. The anonymous sum type
[RFC](https://github.com/rust-lang/rfcs/issues/294) has been posted to solve
this problem.

### C

The most significant "C" way is by defining error number or error return value.
The `errno` is a global thread local variable, through which you can get the
number of last error. Documents of the system call (or library function) should
describe the value.

For example, the signature of `open` function:

```c
int open(const char *pathname, int flags);
```

The manual page will describe all possible errors and how it occurs:

> `open()`, `openat()`, and `creat()` can fail with the following errors:
>
> **EACCES** The requested access to the file is not allowed, or search
>        permission is denied for one of the directories in the path prefix of
>        pathname, or the file did not exist yet and write access to the parent
>        directory is not allowed.  (See also path_resolution(7).)
>
> **EACCES** Where O_CREAT is specified, the protected_fifos or
>        protected_regular sysctl is enabled, the file already exists and is a
>        FIFO or regular file, the owner of the file is neither the current user
>        nor the owner of the containing directory, and the containing directory
>        is both world- or group-writable and sticky.  For details, see the
>        descriptions of /proc/sys/fs/protected_fifos and
>        /proc/sys/fs/protected_regular in proc(5).
>
> **EBUSY**  O_EXCL was specified in flags and pathname refers to a block device
>        that is in use by the system (e.g., it is mounted).
>
> ...

There are hundreds of error numbers, which can be regarded as a "super union" of
errors (and you can call it error universe 😄️).

## Implementation

The implementation for `+thaterror:error` and `+thaterror:transparent` is quite
straightforward. It creates a template for every error and the implementation of
`Error() string` functions is rendering the template. For more documents about
the `template` you can find in
[`text/template`](https://golang.org/pkg/text/template/).

Though there is no tagged union (or sum type) in go, we can use `interface` to
simulate it. This method has been used in some libraries, e.g.
[`ast`](https://golang.org/pkg/go/ast/)

```go
// All statement nodes implement the Stmt interface.
type Stmt interface {
    Node
    stmtNode()
}

func (*BadStmt) stmtNode()        {}
func (*DeclStmt) stmtNode()       {}
func (*EmptyStmt) stmtNode()      {}
func (*LabeledStmt) stmtNode()    {}
func (*ExprStmt) stmtNode()       {}
```

If you specify the `Stmt` as the type of a variable, the all possible types of
it are those listed below. There is also a tool
[typeswitch](https://github.com/gostaticanalysis/typeswitch) to check whether
all types has appeared in the switch branch (but I haven't tried it).

In `thaterror`, we use the same way to implement `+thaterror:wrap`. We will
create an interface for the type which could `Wrap` other types, then we
implement this interface for all these types.

```go
// +thaterror:wrap=*MissingTemplateName
// +thaterror:wrap="pkg/commonerror".*IOError
type Error struct {
	Err error
}
```

`thaterror` will generate an interface for it to conclude all possible wrapped errors:

```go
type ErrorWrapUnion interface {
    PkgwebhookconfigError()
    error
}
```

And related type, e.g. "*IOError" will implement this function:

```go
func (err *IOError) PkgwebhookconfigError()                          {}
```

## Install & Use

`thaterror` hasn't prepared to be widely used. The documents and tests are not 
rich enough. You can install and read the help information to have a try. If 
you have any suggestion on the error handling tools or lints, feel free to open
an issue and help us to improve `thaterror`.
