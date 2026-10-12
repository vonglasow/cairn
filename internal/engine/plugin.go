package engine

// The Claude Code plugin, written out of the binary that carries it. See
// plugin/claude and docs/claude-code.md.

import (
	"path/filepath"

	"cairn/internal/build"
	claudeplugin "cairn/plugin/claude"
)

// PluginWritten says where the plugin was written, which is what a person
// hands Claude Code next.
type PluginWritten struct {
	Plugin string `json:"plugin"`
}

// PluginWrite writes the Claude Code plugin the binary carries, as a
// marketplace Claude Code installs from.
var PluginWrite = Operation{
	Thing: "plugin",
	Verb:  "write",
	Short: "Write the Claude Code plugin this cairn carries",
	Does:  "Write the Claude Code plugin this cairn carries, as a marketplace Claude Code installs from: beside the store unless a folder is named. Writing it again replaces what was written before.",
	Example: `cairn plugin write
cairn plugin write --dir ~/claude/cairn`,
	Level:   Write,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: PluginWritten{},
	Takes: []Argument{
		{Name: "dir", Kind: Text, Means: "the folder to write it into; beside the store when not named"},
	},
	Run: func(on On, in Input) (Answer, error) {
		dir := in.Given["dir"]
		if dir == "" {
			dir = filepath.Join(filepath.Dir(on.Path), "claude-plugin")
		}
		if err := claudeplugin.Write(dir, build.Read().Version); err != nil {
			return Answer{}, err
		}
		return Answer{Plugin: &dir}, nil
	},
}

// pluginText says where the plugin is, and the two commands that install it.
func pluginText(dir string) string {
	return "the Claude Code plugin is written in " + dir + "\n\n" +
		"in Claude Code:\n" +
		"  /plugin marketplace add " + dir + "\n" +
		"  /plugin install " + claudeplugin.Name + "@" + claudeplugin.Name + "\n"
}
