package site

import (
	"html"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	"cairn/internal/theme"
)

var escapes = regexp.MustCompile("\x1b\\[([0-9;]*)m")

type pen struct {
	fg, bg          string
	bold, dim, swap bool
}

func (p pen) classes() string {
	fg, bg := p.fg, p.bg
	if p.swap {
		fg, bg = "z", "z"
		if p.bg != "" {
			fg = p.bg
		}
		if p.fg != "" {
			bg = p.fg
		}
	}
	var said []string
	if fg != "" {
		said = append(said, "t"+fg)
	}
	if bg != "" {
		said = append(said, "u"+bg)
	}
	if p.bold {
		said = append(said, "b")
	}
	if p.dim {
		said = append(said, "d")
	}
	return strings.Join(said, " ")
}

func (p pen) took(codes []string) pen {
	for i := 0; i < len(codes); i++ {
		n, _ := strconv.Atoi(codes[i])
		switch {
		case n == 0:
			p = pen{}
		case n == 1:
			p.bold = true
		case n == 2:
			p.dim = true
		case n == 22:
			p.bold, p.dim = false, false
		case n == 7:
			p.swap = true
		case n == 27:
			p.swap = false
		case n >= 30 && n <= 37:
			p.fg = strconv.Itoa(n - 30)
		case n >= 90 && n <= 97:
			p.fg = strconv.Itoa(n - 90 + 8)
		case n == 39:
			p.fg = ""
		case n >= 40 && n <= 47:
			p.bg = strconv.Itoa(n - 40)
		case n >= 100 && n <= 107:
			p.bg = strconv.Itoa(n - 100 + 8)
		case n == 49:
			p.bg = ""
		case (n == 38 || n == 48) && i+2 < len(codes) && codes[i+1] == "5":
			if slot, _ := strconv.Atoi(codes[i+2]); slot < 16 && n == 38 {
				p.fg = strconv.Itoa(slot)
			} else if slot < 16 {
				p.bg = strconv.Itoa(slot)
			}
			i += 2
		case (n == 38 || n == 48) && i+1 < len(codes) && codes[i+1] == "2":
			i += 4
		}
	}
	return p
}

func terminalHTML(said string) string {
	var out strings.Builder
	var now pen
	at := 0
	open := false
	flush := func(upTo int) {
		if upTo <= at {
			return
		}
		if classes := now.classes(); classes != "" && !open {
			out.WriteString(`<span class="` + classes + `">`)
			open = true
		}
		out.WriteString(html.EscapeString(said[at:upTo]))
	}
	for _, found := range escapes.FindAllStringSubmatchIndex(said, -1) {
		flush(found[0])
		next := now.took(strings.Split(said[found[2]:found[3]], ";"))
		if next != now && open {
			out.WriteString("</span>")
			open = false
		}
		now, at = next, found[1]
	}
	flush(len(said))
	if open {
		out.WriteString("</span>")
	}
	return out.String()
}

func terminal16(s theme.Scheme) [16]color.NRGBA {
	slots := s.Slots()
	var t [16]color.NRGBA
	for i, slot := range []int{0x0, 0x8, 0xB, 0xA, 0xD, 0xE, 0xC, 0x5, 0x3, 0x8, 0xB, 0xA, 0xD, 0xE, 0xC, 0x7} {
		t[i] = slots[slot]
	}
	return t
}
