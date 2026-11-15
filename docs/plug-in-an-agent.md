# Plug in an agent

With Claude Code, install the plugin; it declares the MCP server for you:

```sh
cairn plugin show     # how to install the plugin of this version
cairn plugin write    # write it where Claude Code reads it
```

With any other MCP client, declare the server yourself:

```json
{
  "cairn": {
    "command": "cairn",
    "args": ["mcp", "serve", "--access", "write"],
    "env": { "CAIRN_WORKSPACE": "work" }
  }
}
```

`--access` says how far the agent reaches: `read`, the default, only looks;
`write` also adds and changes; `full` also deletes and restores. A tool past
that level is not offered.

The agent then has one tool per command, under the same name: `task_new`,
`task_list`, `objective_show`. It answers from one workspace, the one the
server started on.

Ask it, in your own words:

- "on travaille sur cairn#12" — it focuses the task and reads its notes;
- "what should I work on?" — it weighs the board and proposes;
- "where do things stand?" — it reads back the work in focus and done.

What it did is in each task's trace, said as its own.

See [the MCP server](mcp.md) and [Claude Code](claude-code.md).
