package site

import (
	"bytes"
	"html"
	"path"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const (
	repository = "https://github.com/vonglasow/cairn"
	address    = "https://vonglasow.github.io/cairn/"
)

var (
	aTag      = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	tags      = regexp.MustCompile(`<[^>]*>`)
	spaces    = regexp.MustCompile(`\s+`)
	markdown  = goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()))
	githubURL = func(to, version string) string {
		ref := "main"
		if aTag.MatchString(version) {
			ref = version
		}
		return repository + "/blob/" + ref + "/" + to
	}
)

type heading struct{ ID, Title string }

type written struct {
	title string
	body  string
	toc   []heading
}

func headingID(said string) string {
	var id strings.Builder
	for _, r := range strings.ToLower(said) {
		switch {
		case r == ' ':
			id.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			id.WriteRune(r)
		}
	}
	return id.String()
}

type ids map[string]int

func (seen ids) Generate(value []byte, _ ast.NodeKind) []byte {
	id := headingID(string(value))
	n := seen[id]
	seen[id] = n + 1
	if n > 0 {
		id += "-" + strconv.Itoa(n)
	}
	return []byte(id)
}

func (seen ids) Put(value []byte) { seen[string(value)]++ }

func linked(from, href, version string) string {
	if href == "" || strings.HasPrefix(href, "#") || strings.Contains(href, "://") || strings.HasPrefix(href, "mailto:") {
		return href
	}
	to, fragment, cut := strings.Cut(href, "#")
	if cut {
		fragment = "#" + fragment
	}
	to = path.Clean(path.Join(path.Dir(from), to))
	if name, ok := pageFor(to); ok {
		return name + fragment
	}
	return githubURL(to, version) + fragment
}

func plain(n ast.Node, src []byte) string {
	var said strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			said.Write(t.Segment.Value(src))
			if t.SoftLineBreak() {
				said.WriteByte(' ')
			}
		case *ast.String:
			said.Write(t.Value)
		case *ast.CodeSpan:
			for child := t.FirstChild(); child != nil; child = child.NextSibling() {
				if s, ok := child.(*ast.Text); ok {
					said.Write(s.Segment.Value(src))
				}
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return said.String()
}

func write(from string, src []byte, version string) (written, error) {
	doc := markdown.Parser().Parse(text.NewReader(src), parser.WithContext(parser.NewContext(parser.WithIDs(ids{}))))
	var out written
	var title ast.Node
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := n.(type) {
		case *ast.Link:
			t.Destination = []byte(linked(from, string(t.Destination), version))
		case *ast.Heading:
			if t.Level == 1 && title == nil {
				title, out.title = t, plain(t, src)
				return ast.WalkSkipChildren, nil
			}
			if id, ok := t.AttributeString("id"); ok && t.Level == 2 {
				out.toc = append(out.toc, heading{ID: string(id.([]byte)), Title: plain(t, src)})
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return written{}, err
	}
	if title != nil {
		doc.RemoveChild(doc, title)
	}
	var body bytes.Buffer
	if err := markdown.Renderer().Render(&body, src, doc); err != nil {
		return written{}, err
	}
	out.body = body.String()
	return out, nil
}

func searchable(body string) string {
	return strings.TrimSpace(spaces.ReplaceAllString(html.UnescapeString(tags.ReplaceAllString(body, " ")), " "))
}
