package tui

import (
	"strings"
	"testing"
)

func offCaps(drawn string) []string {
	var off []string
	var line strings.Builder
	onCap, escape := false, ""
	for _, r := range drawn {
		switch {
		case r == 0x1b:
			escape = "\x1b"
		case escape != "" && r == 'm':
			escape += "m"
			if strings.Contains(escape, "48;") {
				onCap = true
			} else if escape == "\x1b[m" || escape == "\x1b[0m" {
				onCap = false
			}
			escape = ""
		case escape != "":
			escape += string(r)
		case r == '\n':
			off = append(off, line.String())
			line.Reset()
		case !onCap:
			line.WriteRune(r)
		default:
			line.WriteRune(' ')
		}
	}
	return append(off, line.String())
}

func TestEveryKeyTheTerminalNamesIsOnItsCap(t *testing.T) {
	marks := []string{"⏎", "⌘", "⇧", "⌫", "⇥", "^"}
	words := map[string]bool{"esc": true, "enter": true}
	b := standing(twoTasks())
	var wrong []string
	for name, drawn := range map[string]string{
		"the board":   b.Draw(),
		"the help":    hit(b, "?").Draw(),
		"the filters": hit(b, "F").Draw(),
		"the detail":  hit(b, "ctrl+l", "enter").Draw(),
	} {
		for _, line := range offCaps(drawn) {
			for _, word := range strings.Fields(line) {
				named := words[strings.Trim(word, ",.:")]
				for _, mark := range marks {
					named = named || strings.Contains(word, mark)
				}
				if named {
					wrong = append(wrong, name+": "+strings.TrimSpace(line))
					break
				}
			}
		}
	}
	if len(wrong) > 0 {
		t.Errorf("%d line(s) name a key off its cap:\n%s", len(wrong), strings.Join(wrong, "\n"))
	}
}
