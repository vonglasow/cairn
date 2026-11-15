Feature: Where cairn looks for its own places

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: One file names the workspaces, at the place configuration lives

    @config-is-found-where-configuration-lives
    Scenario: the file is found without being named
      Given a configuration whose default workspace is "work" at "~/work"
      When I add a task titled "Write the spec"
      Then the task was written to "~/work/cairn.db"

  Rule: When nothing says a place, it falls under the home

    @config-nothing-said-puts-the-store-under-the-home
    Scenario: neither the environment nor a file says where the store is
      Given a configuration that says nothing about the store
      When I add a task titled "Write the spec"
      Then the task was written to "~/.cairn/cairn.db"

  Rule: What the environment says overrides everything

    @config-the-environment-overrides-the-file-for-the-store
    Scenario: the environment overrides a store the file named
      Given a configuration whose default workspace is "work" at "~/work"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I add a task titled "Write the spec"
      Then the task was written to "~/elsewhere/cairn.db"

  Rule: A configuration can be written for you, every key in it

    @config-init-writes-every-key
    Scenario: config init writes the whole file, every key set to what is used when nothing is said
      When I run `cairn config init`
      Then it succeeds
      And it names the file it wrote
      And the file it wrote is the configuration the documentation shows
      When I run `cairn config show`
      Then it says the store is at "~/.cairn/cairn.db", named by the default workspace, "main"
      And it says the plugin is at "~/.cairn/claude-plugin", named by the file

  Rule: The Claude Code plugin is written where the file says, or beside the store

    @config-names-the-plugin-folder
    Scenario: the file says where the Claude Code plugin is written, and every command writes it there
      Given a configuration that says the plugin lives at "~/claude/cairn"
      When I run `cairn plugin write`
      Then it succeeds
      And "~/claude/cairn" holds the Claude Code plugin
      When I run `cairn config show`
      Then it says the plugin is at "~/claude/cairn", named by the file

  Rule: Every place named is absolute

    @config-a-place-that-is-not-absolute-is-refused
    Scenario Outline: a place that is not absolute is refused, and named
      Given a configuration whose default workspace is "main" at "~/.cairn", its notes at "<place>"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it refuses, and says the place must be absolute
      And it names "<place>"

      Examples:
        | place        |
        | vault        |
        | ./vault      |
        | ../vault     |

  Rule: The file is TOML, and says what each place is for

    @config-a-key-under-the-wrong-table-names-its-own
    Scenario: a key cairn knows, written under another table, is passed over with the table it belongs under
      Given a configuration that says "keep_sizes" is "true" under "[pages]"
      When I run `cairn task list`
      Then it succeeds
      And it says on standard error that "pages.keep_sizes" is ignored
      And it says on standard error "[window]"

  Rule: A configuration that cannot be read stops the command

    @config-that-is-not-there-is-the-ordinary-case
    Scenario: no file at all stops nothing
      When I add a task titled "Write the spec"
      Then the task was written to "~/.cairn/cairn.db"

  Rule: Archiving on its own counts from a day, in periods of a number of days

    @config-show-says-what-is-archived-on-its-own
    Scenario: archiving on its own is shown
      Given a configuration that archives every 10 days from 2026-10-21
      When I run `cairn config show`
      Then it says it archives every 10 days from 2026-10-21

  Rule: A focus session asks to be renewed only when the configuration says after how long

    @config-a-focus-session-that-cannot-be-read-is-refused
    Scenario: a focus session length that is not a duration is refused
      Given focus sessions last "an hour"
      When I run `cairn task focus`
      Then it fails
      And it names "focus.session"

  Rule: A label asks for the approach to be agreed first only when the configuration names it

    @config-show-says-the-approach-label
    Scenario: config show says which label asks for the approach to be agreed first
      Given a configuration whose approach label is "discuss"
      When I run `cairn config show`
      Then it says "label" is "discuss"

  Rule: A deleted task takes its time with it only when the configuration says so

    @config-show-says-a-deleted-task-takes-its-time
    Scenario: config show says a deleted task takes its time with it
      Given a deleted task takes its time with it
      When I run `cairn config show`
      Then it says "deleted" is "taken"

    @config-what-a-deleted-task-does-with-its-time-is-one-of-two
    Scenario: what a deleted task does with its time is refused unless kept or taken
      Given the time of a deleted task is "gone"
      When I run `cairn task list`
      Then it fails
      And it names "time.deleted"

  Rule: The weekends are left out of a week only when the configuration says so

    @config-weekends-hidden-or-shown
    Scenario: config show says the weekends are hidden, and shown when nothing is said
      When I run `cairn config show`
      Then it says "weekends" is "shown"
      Given the weekends are hidden
      When I run `cairn config show`
      Then it says "weekends" is "hidden"

    @config-weekends-is-one-of-two
    Scenario: weekends that is neither shown nor hidden is refused
      Given the weekends are "sometimes"
      When I run `cairn task list`
      Then it fails
      And it names "time.weekends"

  Rule: The window keeps its sizes only when the configuration says so

    @config-show-says-the-window-keeps-its-sizes
    Scenario: config show says the window opens with the sizes it was left with
      Given the window keeps its sizes
      When I run `cairn config show`
      Then it says "keep_sizes" is "true"

    @config-keep-sizes-is-yes-or-no
    Scenario: keep_sizes that is not true or false is refused
      Given the window's keep_sizes is "always"
      When I run `cairn task list`
      Then it fails
      And it names "window.keep_sizes"

  Rule: The window's board arrives by the intro the configuration names

    @config-the-intro-is-said
    Scenario: config show says the intro the configuration names
      Given the board arrives by "deck"
      When I run `cairn config show`
      Then it says "intro" is "deck"

  Rule: Showing the configuration says what cairn will use, and where it got it

    @config-show-reports-what-it-cannot-make-sense-of
    Scenario: a configuration it cannot make sense of is reported, not refused
      Given a configuration that is not TOML
      When I run `cairn config show`
      Then it succeeds
      And it says the file cannot be made sense of
      And it says the other commands will refuse until it is
      And `cairn task list` on that same configuration does refuse
