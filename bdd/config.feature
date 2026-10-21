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

  Rule: A sprint begins on a day and lasts a number of days

    @config-show-says-the-sprints
    Scenario: the sprints configured are shown
      Given a configuration whose sprints begin on 2026-10-21 and last 10 days
      When I run `cairn config show`
      Then it says the sprints begin on 2026-10-21, every 10 days

    @config-show-says-there-are-no-sprints
    Scenario: with no sprint configured, it says there is none
      When I run `cairn config show`
      Then it says there are no sprints

    @config-a-sprint-of-no-days-is-refused
    Scenario: a sprint lasting no day is refused
      Given a configuration whose sprints begin on 2026-10-21 and last 0 days
      When I run `cairn task list`
      Then it fails
      And it names "sprint.days"

    @config-a-sprint-must-begin-on-a-day
    Scenario: a sprint that does not begin on a day is refused
      Given a configuration whose sprints begin on "soon"
      When I run `cairn task list`
      Then it fails
      And it names "sprint.start"

  Rule: Showing the configuration says what cairn will use, and where it got it

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
