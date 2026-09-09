package executor

// AST rewriter for Auto tab: rewrites user Python so native list operations
// route through VisualList for visualization.
//
// gpython AST node types (github.com/go-python/gpython/ast):
//
//   ast.List      {Elts []Expr, Ctx ExprContext}     — [1,2,3]
//   ast.Name      {Id Identifier, Ctx ExprContext}    — list, x, VisualList
//   ast.Call      {Func Expr, Args []Expr, ...}       — list(...), f(x)
//   ast.ListComp  {Elt Expr, Generators []Comprehension} — [x*2 for x in r]
//   ast.Comprehension {Target Expr, Iter Expr, Ifs []Expr}
//
// Pipeline: parser.ParseString → AST → rewrite → source-level edits → py.Compile
//
// compileAst is unexported in gpython's compile package, so we identify
// rewrite targets from the AST (accurate node types + positions) but apply
// the changes to the source string, then compile the rewritten source.
//
// Rewrite rules:
//   1. List literal  [x,y,z] in Load ctx → VisualList([x,y,z])
//   2. Name("list")  in Call ctx         → Name("VisualList")
//   3. ListComp      [expr for x in it]  → desugar to for loop + .append()

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

type rewriteOp struct {
	line   int
	col    int
	kind   opKind
	endCol int // for list literals / listcomp: column of the closing ]
	// ListComp desugar fields
	expr   string
	target string
	iter   string
	seqID  int
}

type opKind int

const (
	opWrapList        opKind = iota // [x,y] → VisualList([x,y])
	opRenameList                    // list( → VisualList(
	opDesugarListComp               // [expr for x in it] → for loop
)

// RewriteSource parses Python source, identifies list constructs via AST,
// and returns rewritten source with list operations routing through VisualList.
func RewriteSource(src string) (string, error) {
	tree, err := parser.ParseString(src+"\n", py.ExecMode)
	if err != nil {
		return "", err
	}

	var ops []rewriteOp
	compSeq := 0
	srcLines := strings.Split(src, "\n")

	ast.Walk(tree, func(node ast.Ast) bool {
		switch n := node.(type) {
		case *ast.List:
			if n.Ctx == ast.Load {
				endCol := findClosingBracket(src, n.GetLineno(), n.GetColOffset())
				if endCol >= 0 {
					ops = append(ops, rewriteOp{
						line:   n.GetLineno(),
						col:    n.GetColOffset(),
						kind:   opWrapList,
						endCol: endCol,
					})
				}
			}
		case *ast.Call:
			if name, ok := n.Func.(*ast.Name); ok && string(name.Id) == "list" {
				ops = append(ops, rewriteOp{
					line: name.GetLineno(),
					col:  name.GetColOffset(),
					kind: opRenameList,
				})
			}
		case *ast.ListComp:
			if len(n.Generators) != 1 || len(n.Generators[0].Ifs) != 0 {
				return true
			}
			endCol := findClosingBracket(src, n.GetLineno(), n.GetColOffset())
			if endCol < 0 {
				return true
			}
			lineIdx := n.GetLineno() - 1
			if lineIdx < 0 || lineIdx >= len(srcLines) {
				return true
			}
			compText := srcLines[lineIdx][n.GetColOffset()+1 : endCol]
			expr, target, iter := splitComprehension(compText)
			if expr == "" {
				return true
			}
			ops = append(ops, rewriteOp{
				line:   n.GetLineno(),
				col:    n.GetColOffset(),
				kind:   opDesugarListComp,
				endCol: endCol,
				expr:   expr,
				target: target,
				iter:   iter,
				seqID:  compSeq,
			})
			compSeq++
		}
		return true
	})

	if len(ops) == 0 {
		return src, nil
	}

	// Apply in reverse order so earlier positions stay valid
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].line != ops[j].line {
			return ops[i].line > ops[j].line
		}
		return ops[i].col > ops[j].col
	})

	lines := strings.Split(src, "\n")
	for _, op := range ops {
		idx := op.line - 1 // AST lines are 1-based
		if idx < 0 || idx >= len(lines) {
			continue
		}
		line := lines[idx]
		switch op.kind {
		case opWrapList:
			if op.col < len(line) && op.endCol < len(line) {
				lines[idx] = line[:op.col] + "VisualList(" + line[op.col:op.endCol+1] + ")" + line[op.endCol+1:]
			}
		case opRenameList:
			if op.col+4 <= len(line) && line[op.col:op.col+4] == "list" {
				lines[idx] = line[:op.col] + "VisualList" + line[op.col+4:]
			}
		case opDesugarListComp:
			var indent string
			for _, ch := range line {
				if ch != ' ' && ch != '\t' {
					break
				}
				indent += string(ch)
			}
			tmpVar := fmt.Sprintf("_vlc%d", op.seqID)
			newLine := line[:op.col] + tmpVar + line[op.endCol+1:]
			preamble := []string{
				indent + tmpVar + " = VisualList([])",
				indent + "for " + op.target + " in " + op.iter + ":",
				indent + "    " + tmpVar + ".append(" + op.expr + ")",
			}
			expanded := make([]string, 0, len(lines)+len(preamble))
			expanded = append(expanded, lines[:idx]...)
			expanded = append(expanded, preamble...)
			expanded = append(expanded, newLine)
			if idx+1 < len(lines) {
				expanded = append(expanded, lines[idx+1:]...)
			}
			lines = expanded
		}
	}

	return strings.Join(lines, "\n"), nil
}

// splitComprehension splits "expr for target in iter" into its three parts.
func splitComprehension(text string) (expr, target, iter string) {
	forIdx := findTopLevelKeyword(text, " for ")
	if forIdx < 0 {
		return "", "", ""
	}
	expr = strings.TrimSpace(text[:forIdx])
	rest := text[forIdx+5:]

	inIdx := findTopLevelKeyword(rest, " in ")
	if inIdx < 0 {
		return "", "", ""
	}
	target = strings.TrimSpace(rest[:inIdx])
	iter = strings.TrimSpace(rest[inIdx+4:])
	return expr, target, iter
}

// findTopLevelKeyword finds keyword at bracket/string depth 0.
func findTopLevelKeyword(text, keyword string) int {
	depth := 0
	inStr := byte(0)
	for i := 0; i <= len(text)-len(keyword); i++ {
		ch := text[i]
		if inStr != 0 {
			if ch == inStr && (i == 0 || text[i-1] != '\\') {
				inStr = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			inStr = ch
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		default:
			if depth == 0 && text[i:i+len(keyword)] == keyword {
				return i
			}
		}
	}
	return -1
}

// findClosingBracket scans the source for the ] that matches the [ at the
// given line and column. Returns the column of ] or -1 if not found.
func findClosingBracket(src string, line, col int) int {
	lines := strings.Split(src, "\n")
	idx := line - 1
	if idx < 0 || idx >= len(lines) {
		return -1
	}

	var buf bytes.Buffer
	for i := idx; i < len(lines); i++ {
		if i > idx {
			buf.WriteByte('\n')
		}
		buf.WriteString(lines[i])
	}
	text := buf.String()

	start := col
	if start >= len(text) || text[start] != '[' {
		return -1
	}

	depth := 0
	inStr := byte(0)
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inStr != 0 {
			if ch == inStr && (i == 0 || text[i-1] != '\\') {
				inStr = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			inStr = ch
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				// Translate back to column on the original line
				if i < len(lines[idx]) {
					return i
				}
				return -1 // multi-line list — not supported in v1
			}
		}
	}
	return -1
}
