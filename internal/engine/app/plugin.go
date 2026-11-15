package app

import (
	"cairn/internal/build"
	claudeplugin "cairn/plugin/claude"
	"path/filepath"
	"strings"
)

// PluginWritten says where the plugin was written, which is what a person hands Claude Code next.
type PluginWritten struct {
	Plugin string `json:"plugin"`
}

// PluginShown says the version of the cairn running and the commands that install its plugin in Claude Code from its tag on GitHub, none when no tag holds it.
type PluginShown struct {
	Version string   `json:"version"`
	Install []string `json:"install,omitempty" jsonschema:"the commands to type in Claude Code, in order; none when this cairn is no version released"`
}

// InstallFromTag is how to install the plugin from the tag of a version released; nothing for one that is not.
func InstallFromTag(version string) []string {
	if _, released := ReleaseOf(version); !released {
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

// PluginShownText is plugin show for a person.
func PluginShownText(shown PluginShown) string {
	if len(shown.Install) == 0 {
		return "cairn " + shown.Version + " is no version released, and no tag on GitHub holds its plugin:\n" +
			"cairn plugin write writes the plugin it carries, and says how to install it\n"
	}
	return "the plugin of cairn " + shown.Version + " is on GitHub, at its tag\n\n" + installText(shown.Install)
}

// PluginText is plugin write for a person.
func PluginText(dir string) string {
	return "the Claude Code plugin is written in " + dir + "\n\n" +
		"in Claude Code:\n" +
		"  /plugin marketplace add " + dir + "\n" +
		"  /plugin install " + claudeplugin.Name + "@" + claudeplugin.Name + "\n"
}

// PluginBeside is where the plugin is written when no folder is named: beside the store, with the rest of what cairn keeps.
func PluginBeside(store string) string { return filepath.Join(filepath.Dir(store), "claude-plugin") }
