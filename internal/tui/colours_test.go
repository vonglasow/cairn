package tui

import (
	"cairn/internal/theme"
	"cairn/internal/tui/design"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

var hex = regexp.MustCompile(`#(%02x|[0-9a-fA-F]{3}\b|[0-9a-fA-F]{6}\b)`)

func TestNothingInTheBoardNamesAColourTheTerminalDoesNotSet(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	designed, err := filepath.Glob("design/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, designed...)
	seen := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(seen, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if fault := ownColour(n); fault != "" {
				t.Errorf("%s: %s, where only the scheme's colours belong", seen.Position(n.Pos()), fault)
			}
			return true
		})
	}
}

var named = []string{"Color", "ANSIColor", "RGBColor", "Black", "White", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan",
	"BrightBlack", "BrightWhite", "BrightRed", "BrightGreen", "BrightYellow", "BrightBlue", "BrightMagenta", "BrightCyan"}

func ownColour(n ast.Node) string {
	switch node := n.(type) {
	case *ast.BasicLit:
		if node.Kind == token.STRING && hex.MatchString(node.Value) {
			return "a colour written in hex, " + node.Value
		}
	case *ast.SelectorExpr:
		id, ok := node.X.(*ast.Ident)
		if !ok || id.Name != "lipgloss" {
			return ""
		}
		if slices.Contains(named, node.Sel.Name) {
			return "lipgloss." + node.Sel.Name
		}
	case *ast.CompositeLit:
		if sel, ok := node.Type.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "color" {
				return "a colour made by hand, color." + sel.Sel.Name
			}
		}
	}
	return ""
}

func TestTheGuardRefusesWhatItIsThereToRefuse(t *testing.T) {
	for source, refused := range map[string]bool{
		`lipgloss.Color("203")`:           true,
		`lipgloss.Color("1")`:             true,
		`lipgloss.ANSIColor(39)`:          true,
		`lipgloss.BrightBlack`:            true,
		`"#f38ba8"`:                       true,
		`"#%02x%02x%02x"`:                 true,
		`color.RGBA{R: 1}`:                true,
		`lipgloss.Red`:                    true,
		`lipgloss.BrightBlue`:             true,
		`"#1 in the list"`:                false,
		`lipgloss.NewStyle().Faint(true)`: false,
	} {
		expr, err := parser.ParseExpr(source)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		ast.Inspect(expr, func(n ast.Node) bool {
			found = found || ownColour(n) != ""
			return true
		})
		if found != refused {
			t.Errorf("%s: refused is %v, want %v", source, found, refused)
		}
	}
}

func TestALabelColourNoneOfTheFourteenIsTheCrumbs(t *testing.T) {
	b := Board{skin: design.SkinOf(design.PaletteOf(theme.Default()))}
	if got, want := b.wearing("nothing").Render("●"), b.skin.Crumb.Render("●"); got != want {
		t.Errorf("a colour none of the fourteen is drawn %q, want the crumb's %q", got, want)
	}
}
