package executor

import (
	"os"
	"strings"
	"testing"

	"github.com/go-python/gpython/py"

	_ "github.com/go-python/gpython/stdlib"
)

// runPython executes Python source with VisualList injected into builtins,
// captures stdout, and returns the PyVisualList and the stdout output.
func runPython(t *testing.T, src string) (*PyVisualList, string) {
	t.Helper()

	ctx := py.NewContext(py.DefaultContextOpts())
	defer ctx.Close()

	tmp, err := os.CreateTemp(t.TempDir(), "stdout-*.txt")
	if err != nil {
		t.Fatal(err)
	}

	sys := ctx.Store().MustGetModule("sys")
	sys.Globals["stdout"] = &py.File{File: tmp, FileMode: py.FileWrite}

	pvl := NewPyVisualList()

	// Inject VisualList constructor and a shared instance into the module
	module, code := compileSrc(t, ctx, src)
	module.Globals["VisualList"] = pvl

	_, err = ctx.RunCode(code, module.Globals, module.Globals, nil)
	if err != nil {
		py.TracebackDump(err)
		t.Fatalf("RunCode failed: %v", err)
	}

	tmp.Close()
	got, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	return pvl, string(got)
}

func compileSrc(t *testing.T, ctx py.Context, src string) (*py.Module, *py.Code) {
	t.Helper()
	code, err := py.Compile(src+"\n", "<test>", py.ExecMode, 0, true)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	module, err := ctx.Store().NewModule(ctx, &py.ModuleImpl{
		Info: py.ModuleInfo{FileDesc: "<test>"},
	})
	if err != nil {
		t.Fatalf("NewModule failed: %v", err)
	}
	return module, code
}

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

// M5.3: VisualList() is callable and returns an instance
func TestVisualListConstruct(t *testing.T) {
	pvl := NewPyVisualList()
	if pvl.Type() != PyVisualListType {
		t.Errorf("Type() = %v, want PyVisualListType", pvl.Type())
	}
}

// M5.4: __len__
func TestVisualListLen(t *testing.T) {
	pvl, stdout := runPython(t, `
for i in range(5):
    VisualList.append(i)
print(len(VisualList))
`)
	if pvl.inner.Len() != 5 {
		t.Errorf("Len() = %d, want 5", pvl.inner.Len())
	}
	if strings.TrimSpace(stdout) != "5" {
		t.Errorf("stdout = %q, want '5'", stdout)
	}
}

// M5.4: __getitem__
func TestVisualListGetItem(t *testing.T) {
	_, stdout := runPython(t, `
for i in range(3):
    VisualList.append(i * 10)
print(VisualList[0])
print(VisualList[2])
print(VisualList[-1])
`)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	want := []string{"0", "20", "20"}
	for i, w := range want {
		if i >= len(lines) {
			t.Fatalf("missing output line %d", i)
		}
		if strings.TrimSpace(lines[i]) != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}

// M5.4: __setitem__
func TestVisualListSetItem(t *testing.T) {
	_, stdout := runPython(t, `
for i in range(3):
    VisualList.append(i)
VisualList[1] = 99
print(VisualList[1])
`)
	if strings.TrimSpace(stdout) != "99" {
		t.Errorf("stdout = %q, want '99'", stdout)
	}
}

// M5.4: __iter__
func TestVisualListIter(t *testing.T) {
	_, stdout := runPython(t, `
for i in range(4):
    VisualList.append(i)
result = []
for x in VisualList:
    result.append(x)
print(result)
`)
	if strings.TrimSpace(stdout) != "[0, 1, 2, 3]" {
		t.Errorf("stdout = %q, want '[0, 1, 2, 3]'", stdout)
	}
}

// M5.5: .append emits events
func TestVisualListAppendEvents(t *testing.T) {
	pvl, _ := runPython(t, `
VisualList.append(42)
`)
	if pvl.inner.Len() != 1 {
		t.Errorf("Len() = %d, want 1", pvl.inner.Len())
	}
	if len(pvl.Events) == 0 {
		t.Fatal("no events emitted")
	}
	if pvl.Events[0].Type() != "append_begin" {
		t.Errorf("first event = %q, want append_begin", pvl.Events[0].Type())
	}
}

// M5.5: .pop
func TestVisualListPop(t *testing.T) {
	_, stdout := runPython(t, `
VisualList.append(10)
VisualList.append(20)
val = VisualList.pop()
print(val)
print(len(VisualList))
`)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if strings.TrimSpace(lines[0]) != "20" {
		t.Errorf("popped = %q, want '20'", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "1" {
		t.Errorf("len = %q, want '1'", lines[1])
	}
}

// M5.5: .insert
func TestVisualListInsert(t *testing.T) {
	_, stdout := runPython(t, `
VisualList.append(1)
VisualList.append(3)
VisualList.insert(1, 2)
for x in VisualList:
    print(x)
`)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	want := []string{"1", "2", "3"}
	for i, w := range want {
		if i >= len(lines) || strings.TrimSpace(lines[i]) != w {
			t.Errorf("line %d = %q, want %q", i, lines[i], w)
		}
	}
}

// M5.5: .extend
func TestVisualListExtend(t *testing.T) {
	pvl, stdout := runPython(t, `
VisualList.append(0)
VisualList.extend([1, 2, 3])
print(len(VisualList))
for x in VisualList:
    print(x)
`)
	_ = pvl
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 5 {
		t.Fatalf("expected 5 lines, got %d: %v", len(lines), lines)
	}
	if strings.TrimSpace(lines[0]) != "4" {
		t.Errorf("len = %q, want '4'", lines[0])
	}
}
