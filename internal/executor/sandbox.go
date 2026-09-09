package executor

import (
	"github.com/go-python/gpython/py"
)

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
