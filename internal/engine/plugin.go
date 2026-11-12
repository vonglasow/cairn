package engine

import (
	"path/filepath"
	"strings"

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
		if err := claudeplugin.Write(dir, on.Home.version()); err != nil {
			return Answer{}, err
		}
		return Answer{Plugin: &dir}, nil
	},
}

// PluginShown says the version of the cairn running and the commands that install its plugin in Claude Code from its tag on GitHub, none when no tag holds it.
type PluginShown struct {
	Version string   `json:"version"`
	Install []string `json:"install,omitempty" jsonschema:"the commands to type in Claude Code, in order; none when this cairn is no version released"`
}

// PluginShow says how Claude Code installs the plugin of the version of this cairn, from its tag on GitHub.
var PluginShow = Operation{
	Thing:   "plugin",
	Verb:    "show",
	Short:   "Say how to install the Claude Code plugin of this cairn's version",
	Does:    "Say the commands that install, in Claude Code, the plugin of this cairn's version from its tag on GitHub; the first removes a plugin of another version, and is left out the first time. A cairn that is no version released has no tag: its plugin is written with cairn plugin write.",
	Example: `cairn plugin show`,
	Level:   Read,
	Needs:   NoStore,
	Answers: PluginShown{},
	Run: func(on On, _ Input) (Answer, error) {
		shown := PluginShown{Version: on.Home.version(), Install: installFromTag(on.Home.version())}
		return Answer{Install: &shown}, nil
	},
}

func installFromTag(version string) []string {
	if _, released := releaseOf(version); !released {
		return nil
	}
	return []string{
		"/plugin marketplace remove " + claudeplugin.Name,
		"/plugin marketplace add " + strings.TrimPrefix(build.Home, "https://github.com/") + "@" + version,
		"/plugin install " + claudeplugin.Name + "@" + claudeplugin.Name,
	}
}

func installText(install []string) string {
	return "in Claude Code, the first line only where a plugin of another version is installed:\n" +
		"  " + strings.Join(install, "\n  ") + "\n"
}

func pluginShownText(shown PluginShown) string {
	if len(shown.Install) == 0 {
		return "cairn " + shown.Version + " is no version released, and no tag on GitHub holds its plugin:\n" +
			"cairn plugin write writes the plugin it carries, and says how to install it\n"
	}
	return "the plugin of cairn " + shown.Version + " is on GitHub, at its tag\n\n" + installText(shown.Install)
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
