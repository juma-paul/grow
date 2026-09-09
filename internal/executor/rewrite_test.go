package executor

import (
	"strings"
	"testing"
)

func TestRewriteListLiteral(t *testing.T) {
	src := `x = [1, 2, 3]`
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	want := `x = VisualList([1, 2, 3])`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRewriteEmptyList(t *testing.T) {
	src := `x = []`
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	want := `x = VisualList([])`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRewritePreservesNonList(t *testing.T) {
	src := "x = dict()\ny = \"hello\""
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if got != src {
		t.Errorf("source was modified:\ngot  %q\nwant %q", got, src)
	}
}

func TestRewriteListCall(t *testing.T) {
	src := `x = list(range(10))`
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	want := `x = VisualList(range(10))`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRewriteListCallNoArgs(t *testing.T) {
	src := `x = list()`
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	want := `x = VisualList()`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestRewriteMultipleOnSameLine(t *testing.T) {
	src := `x, y = [1], list()`
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "VisualList([1])") {
		t.Errorf("list literal not wrapped: %q", got)
	}
	if !strings.Contains(got, "VisualList()") {
		t.Errorf("list() not renamed: %q", got)
	}
}

func TestRewriteMultiLine(t *testing.T) {
	src := "a = []\nb = list()\nc = 42"
	got, err := RewriteSource(src)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "a = VisualList([])" {
		t.Errorf("line 0 = %q", lines[0])
	}
	if lines[1] != "b = VisualList()" {
		t.Errorf("line 1 = %q", lines[1])
	}
	if lines[2] != "c = 42" {
		t.Errorf("line 2 = %q", lines[2])
	}
}
