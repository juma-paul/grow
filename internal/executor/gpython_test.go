package executor

import (
	"os"
	"testing"

	"github.com/go-python/gpython/py"

	_ "github.com/go-python/gpython/stdlib"
)

func TestGpythonHelloWorld(t *testing.T) {
	ctx := py.NewContext(py.DefaultContextOpts())
	defer ctx.Close()

	tmp, err := os.CreateTemp(t.TempDir(), "stdout-*.txt")
	if err != nil {
		t.Fatal(err)
	}

	sys := ctx.Store().MustGetModule("sys")
	sys.Globals["stdout"] = &py.File{File: tmp, FileMode: py.FileWrite}

	_, err = py.RunSrc(ctx, `print("hello")`, "<test>", nil)
	if err != nil {
		py.TracebackDump(err)
		t.Fatalf("RunSrc failed: %v", err)
	}

	tmp.Close()

	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "hello\n" {
		t.Errorf("stdout = %q, want %q", got, "hello\n")
	}
}
