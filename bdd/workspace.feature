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

    @workspace-show-says-the-default-answers
    Scenario: workspace show says the default answers, and where its places are
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace show`
      Then it succeeds
      And it says the workspace "work" answers, because "it is the default"
      And it says the store is at "~/cairn/work/cairn.db"
      And it says the notes are at "~/cairn/work/notes"

    @workspace-show-says-the-flag-answers
    Scenario: workspace show says the flag chose the workspace
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn workspace show --workspace home`
      Then it says the workspace "home" answers, because "--workspace names it"

    @workspace-show-says-no-workspace-names-the-store
    Scenario: a store the environment names and no workspace does is said, with how to name it
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I run `cairn workspace show`
      Then it says no workspace names the store, because "CAIRN_STORE names it"
      And it says "cairn workspace add"

    @workspace-config-show-lists-them
    Scenario: config show lists the workspaces
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      When I run `cairn config show`
      Then it says every key of the configuration, under its table:
        | table      | key              |
        | workspaces | default = "work" |
        | workspaces | home =           |
        | workspaces | work =           |
