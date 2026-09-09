package executor

import (
	"fmt"
	"time"

	"github.com/go-python/gpython/py"
)

type runResult struct {
	obj py.Object
	err error
}

// RunWithTimeout executes compiled code with a wall-clock deadline.
// If the deadline expires, returns a timeout error. The underlying
// goroutine may continue until op/alloc limits stop it.
func RunWithTimeout(ctx py.Context, code *py.Code, globals, locals py.StringDict, timeout time.Duration) (py.Object, error) {
	ch := make(chan runResult, 1)
	go func() {
		obj, err := ctx.RunCode(code, globals, locals, nil)
		ch <- runResult{obj, err}
	}()
	select {
	case r := <-ch:
		return r.obj, r.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("execution timed out after %s", timeout)
	}
}

var blockedBuiltins = []string{
	"open",
	"__import__",
	"exec",
	"eval",
	"compile",
}

// ApplySandbox replaces dangerous builtins with functions that raise
// RuntimeError. Must be called after context creation, before running code.
func ApplySandbox(ctx py.Context) {
	builtins := ctx.Store().Builtins
	for _, name := range blockedBuiltins {
		builtins.Globals[name] = makeBlockedBuiltin(name)
	}
}

func makeBlockedBuiltin(name string) *py.Method {
	return py.MustNewMethod(name, func(self py.Object, args py.Tuple) (py.Object, error) {
		return nil, py.ExceptionNewf(py.RuntimeError, "%s() is not allowed in sandbox mode", name)
	}, 0, name+"() is blocked in sandbox mode")
}
