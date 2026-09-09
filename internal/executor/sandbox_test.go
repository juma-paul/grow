package executor

import (
	"strings"
	"testing"

	"github.com/go-python/gpython/py"

	_ "github.com/go-python/gpython/stdlib"
)

func runSandboxed(t *testing.T, src string) error {
	t.Helper()
	ctx := py.NewContext(py.DefaultContextOpts())
	defer ctx.Close()

	ApplySandbox(ctx)

	_, err := py.RunSrc(ctx, src, "<sandbox-test>", nil)
	return err
}

func TestSandboxBlocksOpen(t *testing.T) {
	err := runSandboxed(t, `open("/etc/passwd")`)
	if err == nil {
		t.Fatal("expected error from open(), got nil")
	}
	if !strings.Contains(err.Error(), "not allowed in sandbox") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxBlocksImport(t *testing.T) {
	err := runSandboxed(t, `import os`)
	if err == nil {
		t.Fatal("expected error from import, got nil")
	}
	if !strings.Contains(err.Error(), "not allowed in sandbox") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxBlocksExec(t *testing.T) {
	err := runSandboxed(t, `exec("x = 1")`)
	if err == nil {
		t.Fatal("expected error from exec(), got nil")
	}
	if !strings.Contains(err.Error(), "not allowed in sandbox") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxBlocksEval(t *testing.T) {
	err := runSandboxed(t, `eval("1+1")`)
	if err == nil {
		t.Fatal("expected error from eval(), got nil")
	}
	if !strings.Contains(err.Error(), "not allowed in sandbox") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSandboxAllowsListOps(t *testing.T) {
	err := runSandboxed(t, `
x = [1, 2, 3]
x.append(4)
y = len(x)
z = list(range(5))
for i in range(3):
    pass
`)
	if err != nil {
		t.Fatalf("safe code should work in sandbox: %v", err)
	}
}
