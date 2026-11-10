package config

// Example is the whole configuration file, every table written and every key explained and set to what cairn uses when the file says nothing — a key with no such value is commented under its own table, so taking the comment off lands it there: what config init writes, and what the README shows.
const Example = `# Where cairn keeps your tasks when [workspaces] names no default: a file.
# Absolute, or under ~. The default workspace, below, comes before it.
# store = "~/.cairn/cairn.db"

# Where cairn looks for the notes of each task when [workspaces] names no
# default: a folder, one per task inside it (~/.cairn/notes/12 for task 12).
# Absolute, or under ~.
# notes = "~/.cairn/notes"

# Where cairn writes its Claude Code plugin, and keeps it at its own version:
# a folder. Beside the store when it is not said. Absolute, or under ~.
plugin = "~/.cairn/claude-plugin"

# The pages every surface offers: the boards' tabs, the commands, the MCP
# tools. Each is on unless it says false. The time is "on", recorded and
# shown; "hidden", recorded and not shown, there again once it is on; or
# "off", neither: a task in focus counts no time until it is on again.
[pages]
tasks = true
objectives = true
time = "on"

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

# The label that asks an agent to agree on the approach with you before any
# code: it writes the solution and what it set aside in the task's notes, and
# waits for your word. Nothing of the kind happens unless a label is said.
[approach]
# label = "discuss"

# What a deleted task does with its time: "kept" on the sheet, said to be
# deleted, or "taken" with it, as though it had never been spent. And whether a
# week of the time page shows its Saturday and Sunday: "shown", or "hidden" for
# a week of five working days.
[time]
deleted = "kept"
weekends = "shown"

# Whether the window opens with the columns, the lanes and the band as wide
# and as tall as it was left with, rather than as they start. Kept in a file
# beside the store. And the colour scheme it wears: one cairn carries —
# default, solarized-dark, solarized-light, gruvbox-dark, gruvbox-light,
# catppuccin-mocha, catppuccin-latte, dracula, nord — or a base16 file of
# your own, ~/.config/cairn/themes/<name>.yaml. A store's settings.toml can
# name its own, so each workspace wears one.
[window]
keep_sizes = false
theme = "default"

# Whether a task may come back as a series: made with a rule, every weekday
# 10:30, and made again when it is done or skipped. Off unless it says true;
# turned off, no series makes anything and every task one made stays.
[recurrence]
on = false

# Workspaces: each a name, and the folder holding its own store and notes
# (<folder>/cairn.db and <folder>/notes). Absolute, or under ~. default names
# the one used when nothing else says; it comes before store and notes above,
# and CAIRN_WORKSPACE, CAIRN_STORE, --workspace and --store come before it.
# main is the folder cairn has always used. cairn workspace add, use and
# remove write here for you. The colour a workspace wears is its own
# settings.toml's to say, under [workspace].
[workspaces]
default = "main"
main = "~/.cairn"
# work = "~/cairn/work"
# home = "~/cairn/home"
`
