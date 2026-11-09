package tui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"cairn/internal/domain"
)

var hex = regexp.MustCompile(`#(%02x|[0-9a-fA-F]{3}\b|[0-9a-fA-F]{6}\b)`)

func TestNothingInTheBoardNamesAColourTheTerminalDoesNotSet(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
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
				t.Errorf("%s: %s, where only the terminal's own accents belong", seen.Position(n.Pos()), fault)
			}
			return true
		})
	}
}

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
		if slices.Contains([]string{"Color", "ANSIColor", "RGBColor", "Black", "White", "BrightBlack", "BrightWhite"}, node.Sel.Name) {
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
		`lipgloss.Red`:                    false,
		`lipgloss.BrightBlue`:             false,
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

func TestEveryColourOfThePaletteIsOneOfTheTerminalsAccents(t *testing.T) {
	p := Default
	for name, ink := range map[string]Ink{
		"low": p.Low, "normal": p.Normal, "high": p.High, "critical": p.Critical,
		"blocked": p.Blocked, "watching": p.Watching, "paused": p.Paused, "reached": p.Reached,
		"late": p.Late, "faint": p.Faint, "bright": p.Bright, "line": p.Line, "accent": p.Accent,
		"stalled": p.Stalled,
	} {
		if found := outsideTheTheme(ink.Style().Render("x")); len(found) > 0 {
			t.Errorf("%s is drawn in %q, outside the terminal's theme", name, found)
		}
	}
}

func TestTheQuietInksAreFaintAndBare(t *testing.T) {
	p := Default
	for name, ink := range map[string]Ink{"normal": p.Normal, "paused": p.Paused, "faint": p.Faint, "line": p.Line} {
		if got := ink.Style().Render("x"); got != "\x1b[2mx\x1b[m" {
			t.Errorf("%s is drawn %q, want faint in the terminal's own foreground", name, got)
		}
		if _, own := ink.Border().(lipgloss.NoColor); !own {
			t.Errorf("a border in %s is drawn in %v, want the terminal's own foreground", name, ink.Border())
		}
	}
	if got := p.Bright.Style().Render("x"); got != "x" {
		t.Errorf("bright is drawn %q, want the terminal's own foreground", got)
	}
}

func TestEveryLabelColourIsTheNearestOfTheTerminals(t *testing.T) {
	b := Board{skin: SkinOf(Default)}
	for _, name := range domain.LabelColours {
		want := "\x1b[" + terminalColours[labelsInTheTerminal[name]] + "m●\x1b[m"
		if got := b.wearing(name).Render("●"); got != want {
			t.Errorf("%s is drawn %q, want %q", name, got, want)
		}
	}
	if got, want := b.wearing("nothing").Render("●"), b.skin.Crumb.Render("●"); got != want {
		t.Errorf("a colour none of the fourteen is drawn %q, want the crumb's %q", got, want)
	}
}

func TestTheFramesWearTheTerminalsColoursAlone(t *testing.T) {
	for _, n := range []int{38, 48, 58, 90, 100} {
		if !paintsOfItsOwn(n) {
			t.Errorf("%d passes as one of the terminal's accents", n)
		}
	}
	for _, n := range []int{1, 2, 7, 30, 31, 36, 37, 39, 47, 49, 91, 96, 97, 107} {
		if paintsOfItsOwn(n) {
			t.Errorf("%s is refused, though the terminal's theme sets it", strconv.Itoa(n))
		}
	}
}
