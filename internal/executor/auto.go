package executor

import (
	"fmt"
	"time"

	"github.com/go-python/gpython/py"
	"github.com/juma-paul/grow/internal/events"

	_ "github.com/go-python/gpython/stdlib"
)

// RunAuto rewrites user Python so list operations route through VisualList,
// executes the rewritten source in a sandboxed gpython context, and returns
// the collected events.
func RunAuto(src string, timeout time.Duration) ([]events.Event, error) {
	rewritten, err := RewriteSource(src)
	if err != nil {
		return nil, fmt.Errorf("rewrite: %w", err)
	}

	code, err := py.Compile(rewritten+"\n", "<auto>", py.ExecMode, 0, true)
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	ctx := py.NewContext(py.DefaultContextOpts())
	ApplySandbox(ctx)

	module, err := ctx.Store().NewModule(ctx, &py.ModuleImpl{
		Info: py.ModuleInfo{FileDesc: "<auto>"},
	})
	if err != nil {
		return nil, fmt.Errorf("module: %w", err)
	}

	var collected []events.Event
	module.Globals["VisualList"] = NewVisualListFactory(func(e events.Event) {
		collected = append(collected, e)
	})

	_, err = RunWithTimeout(ctx, code, module.Globals, module.Globals, timeout)
	if err != nil {
		return collected, err
	}

	return collected, nil
}
