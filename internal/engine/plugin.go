package engine

import (
	"path/filepath"

	"cairn/internal/build"
	claudeplugin "cairn/plugin/claude"
)

// PluginWritten says where the plugin was written, which is what a person hands Claude Code next.
type PluginWritten struct {
	Plugin string `json:"plugin"`
}

// PluginWrite writes the Claude Code plugin the binary carries, as a marketplace Claude Code installs from.
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
			dir = PluginBeside(on.Path)
		}
		if err := claudeplugin.Write(dir, build.Read().Version); err != nil {
			return Answer{}, err
		}
		return Answer{Plugin: &dir}, nil
	},
}

func pluginText(dir string) string {
	return "the Claude Code plugin is written in " + dir + "\n\n" +
		"in Claude Code:\n" +
		"  /plugin marketplace add " + dir + "\n" +
		"  /plugin install " + claudeplugin.Name + "@" + claudeplugin.Name + "\n"
}

// PluginBeside is where the plugin is written when no folder is named: beside the store, with the rest of what cairn keeps.
func PluginBeside(store string) string { return filepath.Join(filepath.Dir(store), "claude-plugin") }

// KeepPluginCurrent writes the plugin beside the store again when a cairn of another version wrote it, and says whether it did: a plugin nobody wrote there is not written.
func KeepPluginCurrent(at Places, version string) (bool, error) {
	dir := PluginBeside(at.Store)
	was, written := claudeplugin.WrittenAt(dir)
	if !written || was == version {
		return false, nil
	}
	return true, claudeplugin.Write(dir, version)
}
