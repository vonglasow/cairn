package config

// Example is the whole configuration file, every table written and every key explained and set to what cairn uses when the file says nothing — a key with no such value is commented under its own table, so taking the comment off lands it there: what config init writes, and what the README shows.
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

# Archiving on its own: days are counted from start, in periods of that many
# days, and when a board opens or the tasks are listed, a task that went done
# before the period under way began is archived. Off unless start is said.
[archive]
# start = 2026-10-21
days = 14

# How long a focus session lasts before cairn asks, in every answer the MCP
# server gives, to renew it by focusing the task again or to end it.
# Nothing is asked unless it is said.
[focus]
# session = "1h"

# What a deleted task does with its time: "kept" on the sheet, said to be
# deleted, or "taken" with it, as though it had never been spent. And whether a
# week of the time page shows its Saturday and Sunday: "shown", or "hidden" for
# a week of five working days.
[time]
deleted = "kept"
weekends = "shown"

# Whether the window opens with the columns, the lanes and the band as wide
# and as tall as it was left with, rather than as they start. Kept in a file
# beside the store.
[window]
keep_sizes = false
`
