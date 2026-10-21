package config

// Example is the whole configuration file, every key commented and set to what cairn uses when the file says nothing: what config init writes, and what the README shows.
const Example = `# Where cairn keeps your tasks: a file. Absolute, or under ~.
store = "~/.cairn/cairn.db"

# Where cairn looks for the notes of each task: a folder, one per task inside
# it (~/.cairn/notes/12 for task 12). Absolute, or under ~.
notes = "~/.cairn/notes"

# Where cairn writes its Claude Code plugin, and keeps it at its own version:
# a folder. Beside the store when it is not said. Absolute, or under ~.
plugin = "~/.cairn/claude-plugin"

# The pages every surface offers: the boards' tabs, the commands, the MCP
# tools. Each is on unless it says false.
[pages]
tasks = true
objectives = true
time = true

# Sprints: when a board opens or the tasks are listed, a task that went done
# before the sprint under way began is archived. Off unless start is said;
# a sprint lasts 14 days unless days says otherwise.
# [sprint]
# start = 2026-10-21
# days = 14
`
