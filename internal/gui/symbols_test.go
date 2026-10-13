package gui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gioui.org/font"
)

func TestEveryCharacterTheWindowWritesHasAFace(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	faces := Faces()
	written := map[rune]string{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(parsed, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			said, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			for _, r := range said {
				if r > 127 {
					written[r] = name
				}
			}
			return true
		})
	}
	if len(written) == 0 {
		t.Fatal("no character past ASCII found: nothing here checks anything")
	}
	for r, name := range written {
		if !covered(faces, r) {
			t.Errorf("%c (U+%04X), written in %s, has no face cairn carries", r, r, name)
		}
	}
}

func covered(faces []font.FontFace, r rune) bool {
	for _, face := range faces {
		if _, ok := face.Face.Face().NominalGlyph(r); ok {
			return true
		}
	}
	return false
}
