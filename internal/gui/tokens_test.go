package gui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoDrawingNamesAColourOrARadiusOfItsOwn(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") || name == "skin.go" {
			continue
		}
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(seen, name, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			for _, fault := range ownTokens(n) {
				t.Errorf("%s: %s, where only the skin's tokens belong", seen.Position(fault.Pos()), sourceOf(source, seen, fault))
			}
			return true
		})
	}
}

func ownTokens(n ast.Node) []ast.Node {
	switch n := n.(type) {
	case *ast.CompositeLit:
		if named(n.Type, "color", "NRGBA") {
			return []ast.Node{n}
		}
		if named(n.Type, "clip", "RRect") || named(n.Type, "clip", "Stroke") {
			var faults []ast.Node
			for _, e := range n.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok && kv.Key.(*ast.Ident).Name != "Rect" && kv.Key.(*ast.Ident).Name != "Path" && numbered(kv.Value) {
					faults = append(faults, kv)
				}
			}
			return faults
		}
	case *ast.ValueSpec:
		var faults []ast.Node
		for _, id := range n.Names {
			if strings.HasSuffix(id.Name, "Radius") {
				faults = append(faults, id)
			}
		}
		return faults
	case *ast.CallExpr:
		if at, ok := radiusOf(n); ok && numbered(n.Args[at]) {
			return []ast.Node{n.Args[at]}
		}
	}
	return nil
}

func radiusOf(call *ast.CallExpr) (int, bool) {
	if id, ok := call.Fun.(*ast.Ident); ok && (id.Name == "rounded" || id.Name == "raised") && len(call.Args) == 4 {
		return 2, true
	}
	if named(call.Fun, "clip", "UniformRRect") && len(call.Args) == 2 {
		return 1, true
	}
	return 0, false
}

func named(e ast.Expr, pkg, name string) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg && sel.Sel.Name == name
}

func numbered(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && (lit.Kind == token.INT || lit.Kind == token.FLOAT) {
			found = true
		}
		return !found
	})
	return found
}

func sourceOf(source []byte, seen *token.FileSet, n ast.Node) string {
	return string(source[seen.Position(n.Pos()).Offset:seen.Position(n.End()).Offset])
}
