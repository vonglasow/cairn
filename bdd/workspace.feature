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

    @workspace-its-notes-are-in-its-folder
    Scenario: the notes of the default workspace are in its folder
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I add a task titled "Write the spec"
      And I run `cairn task show 1`
      Then it says the context lives at "~/cairn/work/notes/1"

    @workspace-no-default-changes-nothing
    Scenario: workspaces named without a default change nothing
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", with no default
      When I add a task titled "Write the spec"
      Then the task was written to "~/.cairn/cairn.db"

  Rule: A flag on the command replaces everything, the environment included

    @workspace-the-flag-picks-another
    Scenario: --workspace picks another workspace for one command
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task new "Buy bread" --workspace home`
      Then the task was written to "~/cairn/home/cairn.db"

    @workspace-the-flag-replaces-the-environment
    Scenario: --workspace replaces what the environment says
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I run `cairn task new "Buy bread" --workspace home`
      Then the task was written to "~/cairn/home/cairn.db"

    @workspace-the-environment-replaces-the-default
    Scenario: the environment replaces the default workspace
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I add a task titled "Write the spec"
      Then the task was written to "~/elsewhere/cairn.db"

    @workspace-a-store-named-keeps-its-own-notes
    Scenario Outline: a store the environment names keeps its own notes, never the default workspace's
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "<store>"
      When I add a task titled "Write the spec"
      And I run `cairn task show 1`
      Then it says the context lives at "<notes>"

      Examples:
        | store                 | notes                |
        | ~/elsewhere/cairn.db  | ~/.cairn/notes/1     |
        | ~/cairn/home/cairn.db | ~/cairn/home/notes/1 |

    @workspace-the-store-flag-names-a-store
    Scenario: --store names a store for one command
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task new "Buy bread" --store ~/loose/cairn.db`
      Then the task was written to "~/loose/cairn.db"

    @workspace-one-flag-or-the-other
    Scenario: --workspace and --store together are not understood
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task list --workspace home --store ~/loose/cairn.db`
      Then it exits with code 2
      And it names "--workspace"
      And it names "--store"

  Rule: A workspace named is one that is there

    @workspace-an-unknown-name-is-refused
    Scenario: a workspace nobody named is refused, with the ones there are
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task list --workspace play`
      Then it refuses, and names "play"
      And it names "home"
      And it names "work"

    @workspace-a-default-nobody-named-is-refused
    Scenario: a default that names no workspace is refused
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "play"
      When I run `cairn task list`
      Then it refuses, and names "play"

    @workspace-a-folder-that-is-not-there-is-refused
    Scenario: a workspace whose folder is not there is refused, naming the folder
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the folder "~/cairn/home" is not there
      When I run `cairn task list --workspace home`
      Then it refuses, and names the folder "~/cairn/home"

    @workspace-a-folder-that-is-not-absolute-is-refused
    Scenario: a workspace whose folder is not absolute is refused, naming it
      Given a configuration naming the workspace "work" at "cairn/work"
      When I run `cairn task list`
      Then it refuses, and says the place must be absolute
      And it names "workspaces.work"

  Rule: The workspaces are listed, added, chosen and shown

    @workspace-list-names-each-and-the-default
    Scenario: workspace list names each workspace and its folder, the default marked
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace list`
      Then it succeeds
      And it lists "work" at "~/cairn/work" as the default
      And it lists "home" at "~/cairn/home"

    @workspace-list-as-json
    Scenario: workspace list answers a machine
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace list --json`
      Then it succeeds
      And the structured answer holds "work" at "~/cairn/work" as the default

    @workspace-list-marks-the-time
    Scenario: workspace list marks a workspace whose time is not on, as its own settings say
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        time = "off"
        """
      And the global configuration also says:
        """
        [pages]
        time = "hidden"
        """
      When I run `cairn workspace list`
      Then it lists "home" at "~/cairn/home", marked "time off"
      And it lists "work" at "~/cairn/work", marked "time hidden"
      When I run `cairn workspace list --json`
      Then the structured answer holds "home" with its time "off"

    @workspace-add-writes-it-and-makes-its-folder
    Scenario: workspace add writes the workspace into the file and makes its folder
      Given a configuration that says the notes live at "~/vault"
      When I run `cairn workspace add play ~/cairn/play`
      Then it succeeds
      And the folder "~/cairn/play/notes" is there
      And the configuration still holds, unchanged, everything it held
      When I run `cairn workspace list`
      Then it lists "play" at "~/cairn/play"

    @workspace-add-without-a-file-writes-only-the-table
    Scenario: workspace add with no configuration writes one holding the workspaces alone
      Given no configuration file
      When I run `cairn workspace add play ~/cairn/play`
      Then it succeeds
      And the configuration reads, whole:
        """
        [workspaces]
        play = "~/cairn/play"
        """

    @workspace-add-joins-the-table-there
    Scenario: a workspace added joins the workspaces already named
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace add play ~/cairn/play`
      Then it succeeds
      And the configuration still holds, unchanged, everything it held
      When I run `cairn workspace list`
      Then it lists "work" at "~/cairn/work" as the default
      And it lists "play" at "~/cairn/play"

    @workspace-add-refuses-a-name-known
    Scenario: workspace add refuses a name already there
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace add work ~/cairn/other`
      Then it refuses, and names "work"
      And the configuration still holds, unchanged, everything it held

    @workspace-add-refuses-a-name-that-is-not-plain
    Scenario Outline: workspace add refuses a name that is not plain
      When I run `cairn workspace add <name> ~/cairn/play`
      Then it refuses, and names "<name>"
      And the folder "~/cairn/play" is not made

      Examples:
        | name    |
        | Work    |
        | my_work |
        | default |

    @workspace-add-refuses-a-folder-that-is-not-absolute
    Scenario: workspace add refuses a folder that is not absolute
      When I run `cairn workspace add play cairn/play`
      Then it refuses, and says the place must be absolute
      And it names "cairn/play"

    @workspace-use-sets-the-default
    Scenario: workspace use makes another the default, and says which it was
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace use home`
      Then it succeeds
      And it says the default was "work"
      When I add a task titled "Buy bread"
      Then the task was written to "~/cairn/home/cairn.db"

    @workspace-use-refuses-an-unknown-name
    Scenario: workspace use refuses a workspace nobody named
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace use play`
      Then it refuses, and names "play"
      And the configuration still holds, unchanged, everything it held

    @workspace-use-refuses-a-folder-gone
    Scenario: workspace use refuses a workspace whose folder is gone, naming the folder
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the folder "~/cairn/home" is not there
      When I run `cairn workspace use home`
      Then it refuses, and names the folder "~/cairn/home"
      And the configuration still holds, unchanged, everything it held

    @workspace-show-says-the-default-answers
    Scenario: workspace show says the default answers, and where its places are
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace show`
      Then it succeeds
      And it says the workspace "work" answers, why "the default (no --workspace, no CAIRN_STORE)"
      And it says the store is at "~/cairn/work/cairn.db"
      And it says the notes are at "~/cairn/work/notes"

    @workspace-show-says-the-flag-answers
    Scenario: workspace show says the flag chose the workspace
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace show --workspace home`
      Then it says the workspace "home" answers, why "--workspace — over the default, work"

    @workspace-show-says-no-workspace-names-the-store
    Scenario: a store the environment names and no workspace does is said, with how to name it
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I run `cairn workspace show`
      Then it says no workspace names the store, why "CAIRN_STORE — over the default, work"
      And it says "cairn workspace add"

    @workspace-show-why-names-the-rungs-passed-over
    Scenario Outline: workspace show says which rung answered, and the rungs it passed over
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the environment says the store lives at "~/cairn/work/cairn.db"
      When I run `cairn workspace show <flag>`
      Then it says "why        <why>"

      Examples:
        | flag                 | why                                                          |
        |                      | CAIRN_STORE — over the default, home                         |
        | --workspace home     | --workspace — over CAIRN_STORE (work) and the default (home) |
        | --store ~/x/cairn.db | --store — over CAIRN_STORE (work) and the default (home)     |

    @workspace-show-why-as-json
    Scenario: workspace show tells a machine the rung that answered, and those it passed over
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the environment says the store lives at "~/cairn/work/cairn.db"
      When I run `cairn workspace show --json`
      Then the structured answer says why is "CAIRN_STORE", over the default "home"

    @workspace-show-says-the-pages
    Scenario: workspace show says the pages that are on, and where that is said
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        objectives = false
        """
      When I run `cairn workspace show --workspace home`
      Then it says "pages      tasks and time, set in the store's settings"
      When I run `cairn workspace show`
      Then it says "pages      tasks, objectives and time, by default"

    @workspace-config-show-lists-them
    Scenario: config show lists the workspaces
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn config show`
      Then it says every key of the configuration, under its table:
        | table      | key              |
        | workspaces | default = "work" |
        | workspaces | home =           |
        | workspaces | work =           |

  Rule: An MCP server keeps the workspace it started on

    @workspace-a-server-keeps-the-default-it-started-on
    Scenario: a server started on the default keeps answering from it when the default changes
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And a cairn server started with `cairn mcp serve --access full`
      When I run `cairn workspace use home`
      And I ask the server to add a task titled "Write the spec"
      Then `cairn task list --workspace work` says "Write the spec"
      And no store was written to "~/cairn/home/cairn.db"

    @workspace-a-server-keeps-the-workspace-it-was-given
    Scenario: a server started with --workspace answers from it, whatever the default
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And a cairn server started with `cairn mcp serve --access full --workspace home`
      When I ask the server to add a task titled "Buy bread"
      Then `cairn task list --workspace home` says "Buy bread"
      And no store was written to "~/cairn/work/cairn.db"

    @workspace-the-server-says-the-workspace-it-was-given
    Scenario: workspace_show through a server says its workspace, why, its time, and that switching is not done there
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And a cairn server started with `cairn mcp serve --workspace home`
      When I ask the server for `workspace_show`
      Then the server says the workspace "home" answers, why "when this server started, --workspace — over the default, work"
      And the server says "time       on, by default"
      And the server says "switch     not from here: this server keeps its store"
      And the server's structured answer says the workspace is "home"

    @workspace-the-server-says-the-default-it-started-on
    Scenario: a server started on the default says so, after the default changed
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And a cairn server started with `cairn mcp serve`
      When I run `cairn workspace use home`
      And I ask the server for `workspace_show`
      Then the server says the workspace "work" answers, why "when this server started, the default (no --workspace, no CAIRN_STORE)"

    @workspace-show-says-the-time
    Scenario: workspace show says whether time is kept, and where that is said
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        time = "off"
        """
      When I run `cairn workspace show --workspace home`
      Then it says "time       off, set in the store's settings"
      And it does not say "switch"

  Rule: A store a board opens by path is remembered among the recent folders

    @workspace-a-board-opened-by-path-is-recent
    Scenario: a store a board opened with --store is listed among the recent folders
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      And I run `cairn workspace list`
      Then it lists the recent folder "~/loose"

    @workspace-a-board-opened-by-the-environment-is-recent
    Scenario: a store the environment names, opened by a board, is listed among the recent folders
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I open the board with `cairn tui`
      And I run `cairn workspace list`
      Then it lists the recent folder "~/elsewhere"

    @workspace-a-command-leaves-the-recent-folders-alone
    Scenario: a command run with --store leaves the recent folders as they were
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn task new "Buy bread" --store ~/loose/cairn.db`
      And I run `cairn workspace list`
      Then it lists no recent folder

    @workspace-a-workspace-opened-is-not-a-recent-folder
    Scenario: a board opened on a workspace leaves the recent folders as they were
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --workspace home`
      And I run `cairn workspace list`
      Then it lists no recent folder

    @workspace-ten-recent-folders-are-kept
    Scenario: ten recent folders are kept, the newest first
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And ten recent folders, "~/old/1" opened first and "~/old/10" last
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      And I run `cairn workspace list`
      Then the recent folders listed are 10, "~/loose" first and "~/old/10" next
      And it does not list the recent folder "~/old/1"

    @workspace-a-recent-folder-gone-is-not-found
    Scenario: a recent folder whose store is gone is listed as not found
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/trip/cairn.db`
      And the folder "~/trip" is not there
      And I run `cairn workspace list`
      Then it lists the recent folder "~/trip" as not found

    @workspace-recent-folders-as-json
    Scenario: workspace list answers a machine with the recent folders too
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      And I run `cairn workspace list --json`
      Then the structured answer holds the recent folder "~/loose", opened today

    @workspace-recent-folders-are-not-the-configuration
    Scenario: the recent folders are kept in the state directory, never in the configuration
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      Then the file "~/.local/state/cairn/recent.json" names "~/loose/cairn.db"
      And the configuration still holds, unchanged, everything it held

    @workspace-recent-folders-follow-xdg-state-home
    Scenario: XDG_STATE_HOME says where the recent folders are kept
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says XDG_STATE_HOME is "~/state"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      Then the file "~/state/cairn/recent.json" names "~/loose/cairn.db"

    @workspace-a-recent-folder-kept-as-a-workspace-leaves-the-list
    Scenario: a recent folder a workspace now names is no longer listed among them
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I open the board with `cairn tui --store ~/loose/cairn.db`
      And I run `cairn workspace add loose ~/loose`
      And I run `cairn workspace list`
      Then it lists no recent folder
      And it lists "loose" at "~/loose"

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

    @workspace-colour-read-off-its-name
    Scenario: a workspace whose settings say no colour wears one of the label colours, the same each time
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace list`
      Then "work" wears one of the label colours
      And "work" wears the same colour when the workspaces are listed again

    @workspace-colour-shown-and-as-json
    Scenario: workspace show says the colour, and so does the structured answer of workspace list
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [workspace]
        colour = "teal"
        """
      When I run `cairn workspace show --workspace home`
      Then it says "colour     teal"
      When I run `cairn workspace list --json`
      Then the structured answer holds "home" wearing "teal"

    @workspace-a-colour-not-in-the-palette-is-refused
    Scenario: a colour that is none of the label colours is refused, naming it
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [workspace]
        colour = "chartreuse"
        """
      When I run `cairn task list --workspace home`
      Then it refuses, and names "chartreuse"
      And it names "workspace.colour"

    @workspace-colour-in-the-global-file-is-passed-over
    Scenario: a colour written in the global configuration is passed over, since it belongs to a store
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [workspace]
        colour = "green"
        """
      When I run `cairn task list`
      Then it says on standard error that "workspace.colour" is ignored

  Rule: A workspace removed leaves the configuration, and nothing else

    @workspace-remove-drops-only-the-entry
    Scenario: workspace remove drops the workspace from the configuration and touches no file
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And I run `cairn task new "Buy bread" --workspace home`
      When I run `cairn workspace remove home`
      Then it succeeds
      And the task was written to "~/cairn/home/cairn.db"
      When I run `cairn workspace list`
      Then it does not list "home"
      And it lists "work" at "~/cairn/work" as the default

    @workspace-remove-refuses-an-unknown-name
    Scenario: workspace remove refuses a workspace nobody named
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace remove play`
      Then it refuses, and names "play"
      And the configuration still holds, unchanged, everything it held

    @workspace-remove-refuses-the-default
    Scenario: workspace remove refuses the default, and says how to choose another
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace remove work`
      Then it refuses, and names "work"
      And it names "cairn workspace use"
      And the configuration still holds, unchanged, everything it held

    @workspace-a-folder-gone-is-listed-not-found
    Scenario: a workspace whose folder is gone is listed as not found
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the folder "~/cairn/home" is not there
      When I run `cairn workspace list`
      Then it lists "home" at "~/cairn/home" as not found
