package executor

import (
	"testing"
	"time"

	_ "github.com/go-python/gpython/stdlib"
)

func TestRunAutoListLiteral(t *testing.T) {
	evts, err := RunAuto("x = [1, 2, 3]\nx.append(4)\n", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(evts) == 0 {
		t.Fatal("no events produced")
	}
	appendCount := 0
	for _, e := range evts {
		if e.Type() == "append_begin" {
			appendCount++
		}
	}
	// 3 from constructor + 1 from x.append(4) = 4
	if appendCount != 4 {
		t.Errorf("append_begin count = %d, want 4", appendCount)
	}
}

func TestRunAutoListComp(t *testing.T) {
	evts, err := RunAuto("y = [i*2 for i in range(5)]\n", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	appendCount := 0
	for _, e := range evts {
		if e.Type() == "append_begin" {
			appendCount++
		}
	}
	// 5 appends from the desugared for loop
	if appendCount != 5 {
		t.Errorf("append_begin count = %d, want 5", appendCount)
	}
}

func TestRunAutoMatchesWrapper(t *testing.T) {
	// Run 20 appends through Auto path
	autoEvts, err := RunAuto(`
x = []
for i in range(20):
    x.append(i)
`, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	// Run 20 appends through Wrapper path (existing singleton)
	pvl, _ := runPython(t, `
for i in range(20):
    VisualList.append(i)
`)

	if len(autoEvts) != len(pvl.Events) {
		t.Fatalf("event count: auto=%d, wrapper=%d", len(autoEvts), len(pvl.Events))
	}

	for i := range autoEvts {
		if autoEvts[i].Type() != pvl.Events[i].Type() {
			t.Errorf("event[%d]: auto=%q, wrapper=%q", i, autoEvts[i].Type(), pvl.Events[i].Type())
		}
	}
}
