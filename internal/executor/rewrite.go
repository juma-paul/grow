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
