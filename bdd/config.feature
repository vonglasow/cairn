Feature: Where cairn looks for its own places
  Cairn has two places to find before it can do anything: where the tasks are
  kept, and where the notes taken elsewhere are looked for. Both are found the
  same way, so there is one rule to know rather than two.

  A TOML file says them, at the place this system keeps configuration. What it does
  not say falls under `~/.cairn`, the same way on every machine. And above all of
  that, the environment overrides whatever was settled — it does not take part
  in the chain, it replaces its answer.

  Every `~` below is that home, never the one whoever runs these owns: a run
  poses files and reads them back, so it is given a home of its own, made for it
  and thrown away after.

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: One file says both places, at the place configuration lives

    The file sits where configuration lives on this system, so it is found
    without being named, and it holds the same thing whatever machine you are
    sitting at.

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

    The same rule on every machine, so the answer to "where are my tasks" and
    "where are my notes" never depends on which one you are sitting at. Both sit
    under one folder, so what cairn keeps stays together.

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

    A place named in the environment is a passing choice — a different store for
    one command, a different vault for one session. It is not another step in
    the chain: whatever the file said, and whatever the home would have given,
    it answers instead. And it changes nothing on disk, so the next command
    without it finds things exactly as they were.

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

  Rule: Every place named is absolute

    A place that depends on where the command was run from is a place that
    answers differently depending on who asks. A configuration is read from
    everywhere, so what it names has to mean one thing.

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

    Cairn poses this file for a person to open and change, so it has to be
    readable by one: a name, a place, and a line above saying what that place
    is. That is the whole reason the form carries comments — a file holding two
    paths and nothing else leaves you guessing which does what.

    A name cairn does not know is refused rather than passed over. A setting
    somebody believes is in force and that does nothing is worse than no setting
    at all — the same silence that was fixed once already, on arguments an agent
    sent that nobody read. It is paid for: an older cairn refuses a file written
    by a newer one, and unlike the store, a configuration has no migration to
    carry it across. Taken knowingly.

    @config-is-toml
    Scenario: the file names each place under its own name
      Given no configuration file
      When I add a task titled "Write the spec"
      Then the configuration reads as TOML
      And it names the store under "store"
      And it names the notes under "notes"

    @config-posed-says-what-each-place-is-for
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

    Reading a file that is there but unreadable is not the same as there being
    no file. The first is a problem to say out loud; the second is the ordinary
    case.

    @config-that-cannot-be-read-is-refused
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

    Nothing has to be written by hand to get started: the first command posts a
    file holding the places cairn would have used anyway. Opening it is then
    enough to see where the tasks are kept and where the notes are looked for.

    It is posed once. Whatever is in it afterwards is the person's, and cairn
    does not put it back the way it was — not even the half of it that is
    missing.

    @config-is-posed-at-the-first-run
    Scenario: the first command leaves a configuration behind
      Given no configuration file
      When I add a task titled "Write the spec"
      Then a configuration was posed at the usual place
      And it says where the store is
      And it says where the notes live

    @config-is-posed-once
    Scenario: what the person wrote is left alone
      Given a configuration that says the notes live at "~/vault"
      When I add a task titled "Write the spec"
      Then the configuration still says the notes live at "~/vault"

    @config-a-file-half-said-is-left-half-said
    Scenario: a file that speaks of one place is not completed with the other
      Given a configuration that says only where the notes live
      When I add a task titled "Write the spec"
      Then the configuration still says nothing about the store

    @config-holds-the-defaults-not-what-the-environment-said
    Scenario: a passing override is not written down as a choice
      Given no configuration file
      And the environment says the store lives at "~/elsewhere/cairn.db"
      When I add a task titled "Write the spec"
      Then the configuration says the store lives at "~/.cairn/cairn.db"
      And the task was written to "~/elsewhere/cairn.db"

    @config-that-cannot-be-posed-does-not-stop-the-command
    Scenario: a configuration that could not be written is said, and the work goes on
      Given the usual place cannot be written to
      When I add a task titled "Write the spec"
      Then the task was added
      And it says the configuration could not be posed

    @config-is-not-posed-by-a-server-that-only-reads
    Scenario: a server that was told to only read leaves the disk alone
      Given no configuration file
      When I start the server told to only read
      And I ask it to show task 1
      Then no configuration was posed

  Rule: Showing the configuration says what cairn will use, and where it got it

    A person reaches for this when something is not where they expected it. So it
    answers with the place it will use, and beside it what named that place — the
    file, the environment, or nothing at all. It names the file it read, because
    "which file" is half the question when two machines disagree.

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
