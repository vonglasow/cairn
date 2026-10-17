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
	Thing:   "plugin",
	Verb:    "write",
	Short:   "Write the Claude Code plugin this cairn carries",
	Does:    "Write the Claude Code plugin this cairn carries, as a marketplace Claude Code installs from: where the configuration's plugin key says, beside the store when it says nothing. Writing it again replaces what was written before, and the MCP server writes it again whenever another version of cairn wrote it.",
	Example: `cairn plugin write`,
	Level:   Write,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: PluginWritten{},
	Run: func(on On, _ Input) (Answer, error) {
		dir := on.Plugin
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

// KeepPluginCurrent writes the plugin again, where the configuration says or beside the store, when a cairn of another version wrote it, and says whether it did: a plugin nobody wrote there is not written.
func KeepPluginCurrent(at Places, version string) (bool, error) {
	dir := at.Plugin
	if dir == "" {
		dir = PluginBeside(at.Store)
	}
	was, written := claudeplugin.WrittenAt(dir)
	if !written || was == version {
		return false, nil
	}
	return true, claudeplugin.Write(dir, version)
}
