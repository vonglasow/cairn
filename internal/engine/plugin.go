package engine

import (
	"cairn/internal/engine/app"
	claudeplugin "cairn/plugin/claude"
)

// pluginWrite writes the Claude Code plugin the binary carries, as a marketplace Claude Code installs from.
var pluginWrite = Operation{
	Thing:   "plugin",
	Verb:    "write",
	Short:   "Write the Claude Code plugin this cairn carries",
	Does:    "Write the Claude Code plugin this cairn carries, as a marketplace Claude Code installs from: where the configuration's plugin key says, beside the store when it says nothing. Writing it again replaces what was written before, and the MCP server writes it again whenever another version of cairn wrote it.",
	Example: `cairn plugin write`,
	Level:   write,
	Needs:   noStore,
	Again:   doesItAgain,
	Answers: app.PluginWritten{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		dir := on.Plugin
		if dir == "" {
			dir = app.PluginBeside(on.Path)
		}
		if err := claudeplugin.Write(dir, on.Home.version()); err != nil {
			return Answer{}, err
		}
		return Answer{Plugin: &dir}, nil
	},
}

// pluginShow says how Claude Code installs the plugin of the version of this cairn, from its tag on GitHub.
var pluginShow = Operation{
	Thing:   "plugin",
	Verb:    "show",
	Short:   "Say how to install the Claude Code plugin of this cairn's version",
	Does:    "Say the commands that install, in Claude Code, the plugin of this cairn's version from its tag on GitHub; the first removes a plugin of another version, and is left out the first time. A cairn that is no version released has no tag: its plugin is written with cairn plugin write.",
	Example: `cairn plugin show`,
	Level:   Read,
	Needs:   noStore,
	Answers: app.PluginShown{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		shown := app.PluginShown{Version: on.Home.version(), Install: app.InstallFromTag(on.Home.version())}
		return Answer{Install: &shown}, nil
	},
}

// KeepPluginCurrent writes the plugin again, where the configuration says or beside the store, when a cairn of another version wrote it, and says whether it did: a plugin nobody wrote there is not written.
func KeepPluginCurrent(at Places, version string) (bool, error) {
	dir := at.Plugin
	if dir == "" {
		dir = app.PluginBeside(at.Store)
	}
	was, written := claudeplugin.WrittenAt(dir)
	if !written || was == version {
		return false, nil
	}
	return true, claudeplugin.Write(dir, version)
}
