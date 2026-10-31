Feature: Where cairn looks for its own places

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: One file says both places, at the place configuration lives

    @config-is-found-where-configuration-lives
    Scenario: the file is found without being named
      Given a configuration that says the store lives at "~/work/cairn.db"
      When I add a task titled "Write the spec"
      Then the task was written to "~/work/cairn.db"

    @config-says-where-the-notes-live
    Scenario: the file says where the notes are looked for
      Given a configuration that says the notes live at "~/vault"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it says the context lives at "~/vault/1"

  Rule: When nothing says a place, it falls under the home

    @config-nothing-said-puts-the-store-under-the-home
    Scenario: neither the environment nor a file says where the store is
      Given a configuration that says nothing about the store
      When I add a task titled "Write the spec"
      Then the task was written to "~/.cairn/cairn.db"

    @config-nothing-said-puts-the-notes-under-the-home
    Scenario: neither the environment nor a file says where the notes are
      Given a configuration that says nothing about the notes
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it says the context lives at "~/.cairn/notes/1"

  Rule: What the environment says overrides everything

    @config-the-environment-overrides-the-file-for-the-store
    Scenario: the environment overrides a store the file named
      Given a configuration that says the store lives at "~/work/cairn.db"
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I add a task titled "Write the spec"
      Then the task was written to "~/elsewhere/cairn.db"

    @config-the-environment-overrides-the-file-for-the-notes
    Scenario: the environment overrides notes the file named
      Given a configuration that says the notes live at "~/vault"
      And the environment says the notes live at "~/elsewhere"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it says the context lives at "~/elsewhere/1"

    @config-the-environment-overrides-the-home
    Scenario: the environment overrides what the home would have given
      Given no configuration file
      And the environment says the notes live at "~/vault"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it says the context lives at "~/vault/1"

    @config-the-environment-leaves-the-file-as-it-was
    Scenario: a passing choice does not settle into the file
      Given a configuration that says the notes live at "~/vault"
      And the environment says the notes live at "~/elsewhere"
      When I add a task titled "Write the spec"
      Then the configuration still says the notes live at "~/vault"

  Rule: A configuration can be written for you, every key in it

    @config-init-writes-every-key
    Scenario: config init writes the whole file, every key set to what is used when nothing is said
      When I run `cairn config init`
      Then it succeeds
      And it names the file it wrote
      And the file it wrote is the configuration the README shows
      When I run `cairn config show`
      Then it says the store is at "~/.cairn/cairn.db", named by the file
      And it says the plugin is at "~/.cairn/claude-plugin", named by the file

    @config-init-a-key-uncommented-is-read-under-its-table
    Scenario Outline: any key of the file config init wrote, taken out of its comment, is read under its own table
      Given `cairn config init` wrote the configuration
      And I take the comment off the key "<key>" in it
      When I run `cairn task list`
      Then it succeeds
      When I run `cairn config show`
      Then it says "<key>" is "<value>"

      Examples:
        | key        | value      |
        | start      | 2026-10-21 |
        | session    | 1h         |

    @config-init-a-value-changed-is-read
    Scenario: a value changed in the file config init wrote is read
      Given `cairn config init` wrote the configuration
      And I change "keep_sizes" to "true" in it
      When I run `cairn config show`
      Then it says "keep_sizes" is "true"

    @config-init-leaves-a-file-alone
    Scenario: config init never writes over a configuration already there
      Given a configuration that says the store lives at "~/work/cairn.db"
      When I run `cairn config init`
      Then it fails
      And it says a configuration is already there
      When I run `cairn config show`
      Then it says the store is at "~/work/cairn.db", named by the file

  Rule: The Claude Code plugin is written where the file says, or beside the store

    @config-names-the-plugin-folder
    Scenario: the file says where the Claude Code plugin is written, and every command writes it there
      Given a configuration that says the plugin lives at "~/claude/cairn"
      When I run `cairn plugin write`
      Then it succeeds
      And "~/claude/cairn" holds the Claude Code plugin
      When I run `cairn config show`
      Then it says the plugin is at "~/claude/cairn", named by the file

    @config-nothing-said-puts-the-plugin-beside-the-store
    Scenario: nothing said, the plugin is written beside the store
      Given a configuration that says the store lives at "~/work/cairn.db"
      When I run `cairn config show`
      Then it says the plugin is at "~/work/claude-plugin", named by nothing

    @config-a-plugin-folder-that-is-not-absolute-is-refused
    Scenario: a plugin folder that is not absolute is refused, and named
      Given a configuration that says the plugin lives at "claude/cairn"
      When I run `cairn plugin write`
      Then it refuses, and says the place must be absolute
      And it names "claude/cairn"

  Rule: Every place named is absolute

    @config-a-place-that-is-not-absolute-is-refused
    Scenario Outline: a place that is not absolute is refused, and named
      Given a configuration that says the notes live at "<place>"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it refuses, and says the place must be absolute
      And it names "<place>"

      Examples:
        | place        |
        | vault        |
        | ./vault      |
        | ../vault     |

    @config-a-store-that-is-not-absolute-is-refused
    Scenario: a store that is not absolute is refused, and named
      Given a configuration that says the store lives at "./cairn.db"
      When I add a task titled "Write the spec"
      Then it refuses, and says the place must be absolute
      And it names "./cairn.db"

  Rule: The file is TOML, and says what each place is for

    @config-is-toml @later
    Scenario: the file names each place under its own name
      Given no configuration file
      When I add a task titled "Write the spec"
      Then the configuration reads as TOML
      And it names the store under "store"
      And it names the notes under "notes"

    @config-posed-says-what-each-place-is-for @later
    Scenario: the file explains itself
      Given no configuration file
      When I add a task titled "Write the spec"
      Then the configuration says above the store what it holds
      And it says above the notes what is looked for there

    @config-that-is-not-toml-is-refused
    Scenario: a file that cannot be made sense of stops the command
      Given a configuration that is not TOML
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it refuses, and names the file it could not make sense of

    @config-a-key-under-the-wrong-table-names-its-own
    Scenario: a key cairn knows, written under another table, is refused with the table it belongs under
      Given a configuration that says "keep_sizes" is "true" under "[pages]"
      When I run `cairn task list`
      Then it refuses, and names "pages.keep_sizes"
      And it names "[window]"

    @config-an-unknown-name-is-refused
    Scenario: a name cairn does not know is refused rather than passed over
      Given a configuration that says the notes live at "~/elsewhere"
      And that same configuration says something under "vault"
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it refuses, and names "vault"
      And it names what it accepts instead

  Rule: A configuration that cannot be read stops the command

    @config-that-cannot-be-read-is-refused @later
    Scenario: a file that is there but cannot be read
      Given a configuration file that cannot be read
      And I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it refuses, and names the file it could not read

    @config-that-is-not-there-is-the-ordinary-case
    Scenario: no file at all stops nothing
      When I add a task titled "Write the spec"
      Then the task was written to "~/.cairn/cairn.db"

  Rule: The file is posed at the first run, and never again

    @config-is-posed-at-the-first-run @later
    Scenario: the first command leaves a configuration behind
      Given no configuration file
      When I add a task titled "Write the spec"
      Then a configuration was posed at the usual place
      And it says where the store is
      And it says where the notes live

    @config-is-posed-once @later
    Scenario: what the person wrote is left alone
      Given a configuration that says the notes live at "~/vault"
      When I add a task titled "Write the spec"
      Then the configuration still says the notes live at "~/vault"

    @config-a-file-half-said-is-left-half-said @later
    Scenario: a file that speaks of one place is not completed with the other
      Given a configuration that says only where the notes live
      When I add a task titled "Write the spec"
      Then the configuration still says nothing about the store

    @config-holds-the-defaults-not-what-the-environment-said @later
    Scenario: a passing override is not written down as a choice
      Given no configuration file
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I add a task titled "Write the spec"
      Then the configuration says the store lives at "~/.cairn/cairn.db"
      And the task was written to "~/elsewhere/cairn.db"

    @config-that-cannot-be-posed-does-not-stop-the-command @later
    Scenario: a configuration that could not be written is said, and the work goes on
      Given the usual place cannot be written to
      When I add a task titled "Write the spec"
      Then the task was added
      And it says the configuration could not be posed

    @config-is-not-posed-by-a-server-that-only-reads @later
    Scenario: a server that was told to only read leaves the disk alone
      Given no configuration file
      When I start the server told to only read
      And I ask it to show task 1
      Then no configuration was posed

  Rule: Archiving on its own counts from a day, in periods of a number of days

    @config-show-says-what-is-archived-on-its-own
    Scenario: archiving on its own is shown
      Given a configuration that archives every 10 days from 2026-10-21
      When I run `cairn config show`
      Then it says it archives every 10 days from 2026-10-21

    @config-show-says-nothing-is-archived-on-its-own
    Scenario: with no archiving configured, it says nothing is archived on its own
      When I run `cairn config show`
      Then it says nothing is archived on its own

    @config-archiving-in-periods-of-no-day-is-refused
    Scenario: archiving in periods of no day is refused
      Given a configuration that archives every 0 days from 2026-10-21
      When I run `cairn task list`
      Then it fails
      And it names "archive.days"

    @config-archiving-counts-from-a-day
    Scenario: archiving that does not count from a day is refused
      Given a configuration that archives counting from "soon"
      When I run `cairn task list`
      Then it fails
      And it names "archive.start"

  Rule: A focus session asks to be renewed only when the configuration says after how long

    @config-show-says-how-long-a-focus-session-lasts
    Scenario: config show says how long a focus session lasts before it asks to be renewed
      Given focus sessions last "1h"
      When I run `cairn config show`
      Then it says "session" is "1h"

    @config-a-focus-session-that-cannot-be-read-is-refused
    Scenario: a focus session length that is not a duration is refused
      Given focus sessions last "an hour"
      When I run `cairn task focus`
      Then it fails
      And it names "focus.session"

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

  Rule: Showing the configuration says what cairn will use, and where it got it

    @config-show-says-every-key
    Scenario: config show says every key of the configuration, with the value cairn uses
      When I run `cairn config show`
      Then it says every key of the configuration, under its table:
        | table   | line                |
        |         | store = "           |
        |         | notes = "           |
        |         | plugin = "          |
        | pages   | tasks = true        |
        | pages   | objectives = true   |
        | pages   | time = true         |
        | archive | days = 14           |
        | time    | deleted = "kept"    |
        | time    | weekends = "shown"  |
        | window  | keep_sizes = false  |
      And it says nothing is archived on its own
      And it says a focus session is never asked to be renewed

    @config-show-reads-as-a-configuration
    Scenario: what config show says is a configuration cairn reads as the same
      Given a configuration that archives every 10 days from 2026-10-21
      And that same configuration keeps the window's sizes
      When I run `cairn config show`
      And I write what it said as the configuration
      And I run `cairn config show`
      Then it says the same keys and values as before

    @config-show-names-the-file-it-reads
    Scenario: it says which file it read
      Given a configuration that says the store lives at "~/work/cairn.db"
      When I run `cairn config show`
      Then it names the file it read

    @config-show-says-what-named-each-place
    Scenario: it says what named each place
      Given a configuration that says the store lives at "~/work/cairn.db"
      And the environment says the notes live at "~/vault"
      When I run `cairn config show`
      Then it says the store is at "~/work/cairn.db", named by the file
      And it says the notes are at "~/vault", named by the environment

    @config-show-says-when-nothing-named-a-place
    Scenario: a place nobody named falls to the default, and says so
      When I run `cairn config show`
      Then it says the store is at "~/.cairn/cairn.db", named by nothing

    @config-show-travels-in-the-json
    Scenario: an agent gets the same facts in the shape a machine reads
      Given a configuration that says the store lives at "~/work/cairn.db"
      When I run `cairn config show --json`
      Then the structured answer carries the file it read, each place, and what named it

    @config-show-reports-what-it-cannot-make-sense-of
    Scenario: a configuration it cannot make sense of is reported, not refused
      Given a configuration that is not TOML
      When I run `cairn config show`
      Then it succeeds
      And it says the file cannot be made sense of
      And it says the other commands will refuse until it is
      And `cairn task list` on that same configuration does refuse

    @config-show-answers-an-agent-about-a-file-it-cannot-make-sense-of
    Scenario: an agent asking about a broken configuration is answered, not refused
      Given a configuration that is not TOML
      When I run `cairn config show --json`
      Then it succeeds
      And the structured answer carries the file it read, and what it could not make sense of
