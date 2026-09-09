package executor

import (
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
