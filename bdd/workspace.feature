Feature: Workspaces, each a folder holding a store and its notes

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: The default workspace answers when nothing else says

    @workspace-the-default-answers
    Scenario: the default workspace holds the tasks added
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I add a task titled "Write the spec"
      Then the task was written to "~/cairn/work/cairn.db"

  Rule: A workspace always answers; with none named, main, in ~/.cairn

    @workspace-main-when-none
    Scenario: nothing configured, the workspace main answers from ~/.cairn
      When I run `cairn workspace show`
      Then it says the workspace is "main", at "~/.cairn"

  Rule: The workspace whose folder the current directory sits under answers, under the flag and the environment

    @workspace-the-directory-under-the-environment
    Scenario: CAIRN_WORKSPACE wins over the directory
      Given the workspaces "home" at "~/.cairn" and "cairn" at "~/code/cairn/.cairn", with no default
      And I am in "~/code/cairn"
      And the environment names the workspace "home"
      When I add a task titled "Buy bread"
      Then the task was written to "~/.cairn/cairn.db"

  Rule: A workspace may name its notes; <folder>/notes otherwise

    @workspace-names-its-notes
    Scenario: a workspace's notes are where the configuration says
      Given the workspace "cairn" at "~/code/cairn/.cairn", its notes at "~/vaults/cairn", the default being "cairn"
      When I add a task titled "Write the spec"
      And I run `cairn task show 1`
      Then it says the context lives at "~/vaults/cairn/1"

  Rule: A configuration of v0.11.0 is turned into workspaces on its own, the places it used kept as they were

    @workspace-old-configuration-upgraded
    Scenario: the top-level places become the workspace main, the default, and it says so
      Given a configuration that says the store lives at "~/data/cairn.db" and the notes at "~/data/notes"
      When I add a task titled "Write the spec"
      Then the task was written to "~/data/cairn.db"
      And the configuration names the workspace "main" at "~/data", its notes at "~/data/notes", the default
      And the configuration names no top-level store or notes
      And the configuration as it was is kept in "config.toml.v0.11.0"
      And it says on standard error what it changed, and where the old configuration is

  Rule: Notes a workspace was moved away from are offered to follow, in a terminal

    @workspace-notes-changed-offered-to-move
    Scenario: the notes changed in the configuration, cairn offers to move them
      Given the workspace "cairn" whose notes were at "~/code/cairn/.cairn/notes", holding task 1's folder
      And the configuration now names its notes at "~/vaults/cairn"
      When I run `cairn task list` in a terminal, and answer "y"
      Then "~/vaults/cairn/1" is there
      And "~/code/cairn/.cairn/notes" is gone

    @workspace-notes-changed-left-on-no
    Scenario: on no, nothing is moved, and the next run asks again
      Given the workspace "cairn" whose notes were at "~/code/cairn/.cairn/notes", holding task 1's folder
      And the configuration now names its notes at "~/vaults/cairn"
      When I run `cairn task list` in a terminal, and answer "n"
      Then nothing was moved
      When I run `cairn task list` in a terminal
      Then it asks whether to move the notes

    @workspace-notes-changed-without-a-person
    Scenario: with nobody to answer, nothing is moved and it is said
      Given the workspace "cairn" whose notes were at "~/code/cairn/.cairn/notes", holding task 1's folder
      And the configuration now names its notes at "~/vaults/cairn"
      When the MCP server starts
      Then nothing was moved
      And it says on standard error that a run of cairn in a terminal will offer to move them

    @workspace-notes-not-moved-onto-others
    Scenario: notes are not moved into a folder that already holds something
      Given the workspace "cairn" whose notes were at "~/code/cairn/.cairn/notes", holding task 1's folder
      And the configuration now names its notes at "~/vaults/cairn", which holds "Inbox.md"
      When I run `cairn task list` in a terminal
      Then it says it cannot move them there, naming "~/vaults/cairn"
      And nothing was moved

  Rule: A flag on the command replaces everything, the environment included

    @workspace-the-flag-picks-another
    Scenario: --workspace picks another workspace for one command
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task new "Buy bread" --workspace home`
      Then the task was written to "~/cairn/home/cairn.db"

  Rule: CAIRN_WORKSPACE in the environment names a workspace, under a place the environment names

    @workspace-the-environment-names-a-workspace
    Scenario: CAIRN_WORKSPACE names the workspace whose store and notes answer
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment names the workspace "home"
      When I add a task titled "Write the spec"
      Then the task was written to "~/cairn/home/cairn.db"
      When I run `cairn task show 1`
      Then it says the context lives at "~/cairn/home/notes/1"

  Rule: A workspace named is one that is there

    @workspace-an-unknown-name-is-refused
    Scenario: a workspace nobody named is refused, with the ones there are
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task list --workspace play`
      Then it refuses, and names "play"
      And it names "home"
      And it names "work"

  Rule: The workspaces are listed, added, chosen and shown

    @workspace-list-names-each-and-the-default
    Scenario: workspace list names each workspace and its folder, the default marked
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace list`
      Then it succeeds
      And it lists "work" at "~/cairn/work" as the default
      And it lists "home" at "~/cairn/home"

  Rule: An MCP server keeps the workspace it started on

    @workspace-a-server-keeps-the-default-it-started-on
    Scenario: a server started on the default keeps answering from it when the default changes
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And a cairn server started with `cairn mcp serve --access full`
      When I run `cairn workspace use home`
      And I ask the server to add a task titled "Write the spec"
      Then `cairn task list --workspace work` says "Write the spec"
      And no store was written to "~/cairn/home/cairn.db"

  Rule: A store a board opens by path is remembered among the recent folders

    @workspace-a-board-opened-by-path-is-recent
    Scenario: a store a board opened with --store is listed among the recent folders
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      And I run `cairn workspace list`
      Then it lists the recent folder "~/loose"

  Rule: Each workspace wears a colour, said in its settings or read off its name

    @workspace-colour-said-in-its-settings
    Scenario: a workspace wears the colour its settings say
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [workspace]
        colour = "green"
        """
      When I run `cairn workspace list`
      Then it lists "home" at "~/cairn/home" wearing "green"

  Rule: A workspace removed leaves the configuration, and nothing else

    @workspace-remove-drops-only-the-entry
    Scenario: workspace forget drops the workspace from the configuration and touches no file
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And I run `cairn task new "Buy bread" --workspace home`
      When I run `cairn workspace forget home`
      Then it succeeds
      And the task was written to "~/cairn/home/cairn.db"
      When I run `cairn workspace list`
      Then it does not list "home"
      And it lists "work" at "~/cairn/work" as the default
