package site

import (
	"fmt"
	"html"
	"slices"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cairn/internal/cli"
	"cairn/internal/config"
	"cairn/internal/mcp"
	"cairn/internal/theme"
)

var e = html.EscapeString

func commandPage(c *cobra.Command) string {
	return "cli-" + strings.ReplaceAll(strings.TrimPrefix(c.CommandPath(), c.Root().Name()+" "), " ", "-") + ".html"
}

func leaves(c *cobra.Command) []*cobra.Command {
	var found []*cobra.Command
	for _, sub := range c.Commands() {
		switch {
		case sub.Hidden || !sub.IsAvailableCommand():
		case sub.HasAvailableSubCommands():
			found = append(found, leaves(sub)...)
		default:
			found = append(found, sub)
		}
	}
	return found
}

func tooled() map[string]bool {
	served := map[string]bool{}
	for _, tool := range mcp.Tools() {
		served[tool.Name] = true
	}
	return served
}

func commandPages() []page {
	root := cli.Reference()
	served := tooled()
	type part struct {
		id, title, says string
		rows            []*cobra.Command
	}
	var parts []*part
	var pages []page
	for _, c := range leaves(root) {
		id, title, says := "more", "Everything else", ""
		if parent := c.Parent(); parent != root {
			id, title, says = parent.Name(), parent.Name(), parent.Short
		}
		if len(parts) == 0 || parts[len(parts)-1].id != id {
			parts = append(parts, &part{id: id, title: title, says: says})
		}
		parts[len(parts)-1].rows = append(parts[len(parts)-1].rows, c)
		pages = append(pages, commandOf(c, served))
	}
	var index strings.Builder
	var toc []heading
	index.WriteString(`<p class="lead">Every command, from its own help.</p>`)
	for _, p := range parts {
		toc = append(toc, heading{ID: p.id, Title: p.title})
		fmt.Fprintf(&index, `<h2 id="%s">%s</h2>`, e(p.id), e(p.title))
		if p.says != "" {
			fmt.Fprintf(&index, `<p>%s</p>`, e(p.says))
		}
		index.WriteString(`<div class="tw"><table>`)
		for _, c := range p.rows {
			fmt.Fprintf(&index, `<tr><td><a class="m" href="%s">%s</a></td><td>%s</td></tr>`, commandPage(c), e(c.CommandPath()), e(c.Short))
		}
		index.WriteString(`</table></div>`)
	}
	return append([]page{{Name: commandsPage, Title: "Commands", Group: "Reference", Top: "reference", Body: index.String(), Toc: toc}}, pages...)
}

func flagTable(flags *pflag.FlagSet) string {
	var rows strings.Builder
	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		named := "--" + f.Name
		if f.Shorthand != "" {
			named = "-" + f.Shorthand + ", " + named
		}
		fmt.Fprintf(&rows, `<tr><td class="m">%s</td><td>%s</td></tr>`, e(named), e(f.Usage))
	})
	if rows.Len() == 0 {
		return ""
	}
	return `<div class="tw"><table><tr><th>flag</th><th>says</th></tr>` + rows.String() + `</table></div>`
}

func commandOf(c *cobra.Command, served map[string]bool) page {
	var body strings.Builder
	var toc []heading
	section := func(id, title string) {
		toc = append(toc, heading{ID: id, Title: title})
		fmt.Fprintf(&body, `<h2 id="%s">%s</h2>`, id, title)
	}
	fmt.Fprintf(&body, `<p class="lead">%s</p>`, e(c.Short))
	section("usage", "Usage")
	fmt.Fprintf(&body, `<pre><code>%s</code></pre>`, e(c.UseLine()))
	if c.Long != "" {
		fmt.Fprintf(&body, `<p>%s</p>`, e(c.Long))
	}
	if flags := flagTable(c.LocalFlags()); flags != "" {
		section("flags", "Flags")
		body.WriteString(flags)
	}
	if c.Example != "" {
		section("examples", "Examples")
		fmt.Fprintf(&body, `<pre><code>%s</code></pre>`, e(c.Example))
	}
	if flags := flagTable(c.InheritedFlags()); flags != "" {
		section("every-command", "Flags every command takes")
		body.WriteString(flags)
	}
	if tool := strings.ReplaceAll(strings.TrimPrefix(c.CommandPath(), c.Root().Name()+" "), " ", "_"); served[tool] {
		section("see-also", "See also")
		fmt.Fprintf(&body, `<p><a href="%s#%s">%s</a>, the same for an agent.</p>`, toolsPage, tool, tool)
	}
	return page{Name: commandPage(c), Title: c.CommandPath(), Group: "Reference", Top: "reference", Parent: commandsPage, Body: body.String(), Toc: toc}
}

func argumentsOf(tool *sdk.Tool) string {
	schema, _ := tool.InputSchema.(map[string]any)
	properties, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]string)
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	slices.Sort(names)
	var rows strings.Builder
	for _, name := range names {
		said, _ := properties[name].(map[string]any)
		means, _ := said["description"].(string)
		if slices.Contains(required, name) {
			means += " Required."
		}
		fmt.Fprintf(&rows, `<tr><td class="m">%s</td><td>%s</td></tr>`, e(name), e(means))
	}
	if rows.Len() == 0 {
		return ""
	}
	return `<div class="tw"><table><tr><th>argument</th><th>says</th></tr>` + rows.String() + `</table></div>`
}

func toolsPageOf() page {
	var body strings.Builder
	var toc []heading
	body.WriteString(`<p class="lead">Every tool <code>cairn mcp serve</code> offers an agent, from the tools' own definitions. Each is a command of the <a href="` + commandsPage + `">command line</a>, under the same name.</p>`)
	for _, tool := range mcp.Tools() {
		toc = append(toc, heading{ID: tool.Name, Title: tool.Name})
		kind := "changes"
		switch {
		case tool.Annotations.ReadOnlyHint:
			kind = "reads"
		case tool.Annotations.DestructiveHint != nil && *tool.Annotations.DestructiveHint:
			kind = "asks first"
		}
		fmt.Fprintf(&body, `<h2 id="%s">%s <span class="tag">%s</span></h2><p><b>%s.</b> %s</p>%s`,
			tool.Name, tool.Name, kind, e(tool.Annotations.Title), e(tool.Description), argumentsOf(tool))
	}
	return page{Name: toolsPage, Title: "MCP tools", Group: "Reference", Top: "reference", Body: body.String(), Toc: toc}
}

func configurationIntro() string {
	return `<h2 id="the-whole-file">The whole file</h2><p>What <code>cairn config init</code> writes: every key, explained, set to what cairn uses when the file says nothing.</p><pre><code class="toml">` + e(config.Example) + `</code></pre>`
}

func themesPageOf(schemes []theme.Scheme, pictured bool) page {
	var body strings.Builder
	var toc []heading
	body.WriteString(`<p class="lead">Every scheme cairn carries. The window wears the one <code>[window] theme</code> names; this site wears the one you pick at the top.</p>`)
	for _, s := range schemes {
		toc = append(toc, heading{ID: s.Slug, Title: s.Name})
		variant := "dark"
		if s.Light() {
			variant = "light"
		}
		fmt.Fprintf(&body, `<h2 id="%s">%s</h2><p>%s, by %s. <code>cairn config set window.theme %s</code></p><div class="sw">`, s.Slug, e(s.Name), variant, e(s.Author), s.Slug)
		for _, slot := range s.Slots() {
			fmt.Fprintf(&body, `<i style="background:%s"></i>`, hex(slot))
		}
		body.WriteString(`</div>`)
		if pictured {
			fmt.Fprintf(&body, `<img class="shot" loading="lazy" src="%s" alt="The window in %s">`, windowPicture(s.Slug), e(s.Name))
		}
	}
	return page{Name: themesPage, Title: "Themes", Group: "Reference", Top: "reference", Body: body.String(), Toc: toc}
}
