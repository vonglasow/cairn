# Dependencies

Read this before adding one. `CLAUDE.md` asks for a written reason before a
dependency arrives; this is where every reason is written, one line each, so
the list of what cairn is built from and the reasons it is built from them are
one table rather than scattered across the documents that happened to add them.

Adding a dependency means adding its row here in the same commit. A row whose
reason no longer holds is a dependency to take out.

| dependency | what it is for | why not without it |
|---|---|---|
| `modernc.org/sqlite` | the store | SQLite written in Go: one static binary that cross-compiles, with no C toolchain and no cgo — a store a person keeps on their own machine, see [architecture.md](architecture.md) |
| `github.com/spf13/cobra` | the command tree | the tree is built from the engine's description of every operation; cobra is what reads flags and positionals off it, with help and usage for free |
| `github.com/spf13/pflag` | reading what was typed when a flag is refused | already beneath cobra; named directly so that `-30m` read as three flags is recognised from the parser's own error rather than out of its sentence, which an upgrade may reword |
| `github.com/charmbracelet/fang` | how the command line says help and refusals | styled help, versions and errors over cobra without writing a renderer; refusals reach stderr the same way everywhere |
| `github.com/modelcontextprotocol/go-sdk` | the MCP server | the protocol, measured at 250–350 lines by hand against 35 with it — see [mcp.md](mcp.md) |
| `github.com/google/jsonschema-go` | the schema of every tool's answer | read off the Go type that crosses the wire, so the description and the thing described cannot drift — see [mcp.md](mcp.md) |
| `github.com/pelletier/go-toml/v2` | the configuration file | a file a person opens has to carry comments; strict decoding names a key cairn does not know — see [config.md](config.md) |
| `charm.land/bubbletea/v2` | the terminal board | there is no board on a terminal without an event loop and screen state |
| `charm.land/lipgloss/v2` | drawing the terminal board | borders, joins and widths measured in cells; already pulled in by fang |
| `charm.land/huh/v2` | the terminal board's form | fields, focus and a text area that wraps, rather than a form engine written for one screen |
| `charm.land/bubbles/v2` | the terminal board's scrolling and keys | a viewport for what is longer than its box, and key bindings the help can list |
| `gioui.org` | the window | a vector-drawn window identical on macOS and Linux, in the same single binary, with no webview and no second language — see [gui.md](gui.md) |
| `gioui.org/x` | the system's file dialog, to export and import the store from the window | Gio has none; `explorer` is the one Gio's authors keep beside it, the native dialog on Windows (Win32) and Linux (the desktop portal over D-Bus). Not on macOS: its Objective-C file would be a second package linking `-lobjc`, and the linker says so on every build, so the window opens AppKit's panels itself there (`appkit_darwin.go`). Only that package is imported — see [gui.md](gui.md) |
| `golang.org/x/image` | measuring text in the window | the fixed-point numbers Gio measures and places text in; Gio takes them, and they are not in the standard library |
| `github.com/cucumber/godog` | running the scenarios | every `.feature` under `bdd/` is executed as Go tests by it — the scenarios are the contract, see [workflow.md](workflow.md) |

Four of these reasons were written down only when this table was: `cobra`,
`fang`, `godog` and `modernc.org/sqlite` arrived with the first commit and
nobody wrote why. They are written here as the code shows them used, not as
somebody remembered deciding.
