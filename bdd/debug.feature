Feature: Debug mode

  A person asked to show what went wrong runs cairn with CAIRN_DEBUG=1, does
  the one thing that goes wrong, and sends what was written down.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: Debug mode is turned on by the environment, at launch

    @debug-off-writes-nothing
    Scenario: without CAIRN_DEBUG, nothing is written down
      When I run `cairn task list`
      Then there is no debug folder beside the store
      And it writes nothing to standard error

    @debug-on-writes-a-file-beside-the-store
    Scenario: with CAIRN_DEBUG=1, a run writes one file in the debug folder beside the store
      Given debug mode is on
      When I run `cairn task list`
      Then the debug folder beside the store holds 1 file
      And it says on standard error where that file is

    @debug-each-run-writes-its-own-file
    Scenario: each run writes a file of its own
      Given debug mode is on
      When I run `cairn task list`
      And I run `cairn task list`
      Then the debug folder beside the store holds 2 files

    @debug-a-folder-that-cannot-be-made-is-said
    Scenario: a debug folder that cannot be made is said, and the command still runs
      Given debug mode is on
      And the debug folder cannot be made
      When I run `cairn task list`
      Then it succeeds
      And it says on standard error that it cannot write down

  Rule: The debug folder keeps the last twenty files

    @debug-keeps-the-last-twenty-files
    Scenario: past twenty files, the oldest goes
      Given debug mode is on
      And the debug folder already holds 20 files
      When I run `cairn task list`
      Then the debug folder beside the store holds 20 files
      And the oldest of them is gone

  Rule: The file begins with what the machine is and what cairn holds

    @debug-the-file-begins-with-the-machine
    Scenario: the file begins with the machine, the configuration and the store in numbers
      Given debug mode is on
      And I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task list`
      Then the debug file says "# system"
      And the debug file says "# cairn"
      And the debug file says "# config"
      And the debug file says "# store            1 todo"
      And the debug file says "# command          task list"

  Rule: Every operation is written down once, through the engine

    @debug-an-operation-is-written-down
    Scenario: an operation is written down with what it was given and what it came to
      Given debug mode is on
      When I run `cairn task new "Write the spec" --priority high`
      Then the debug file says "engine task new"
      And the debug file says "priority=high"
      And the debug file says "→ ok"

    @debug-a-refusal-is-written-down
    Scenario: an operation refused is written down with why
      Given debug mode is on
      When I run `cairn task show 99`
      Then the debug file says "engine task show number=99 → refused"

  Rule: What the person wrote never reaches the file

    @debug-written-text-is-masked
    Scenario: a title is written down as a fingerprint and a length, never as itself
      Given debug mode is on
      When I run `cairn task new "Write the spec" --priority high`
      Then the debug file never says "Write the spec"
      And the debug file says "title=‹text "

  Rule: A report gathers every file into one archive to send

    @debug-report-packs-the-files
    Scenario: debug report packs every file kept and a header made now
      Given debug mode is on
      And I work in a folder of my own
      And I ran `cairn task list`
      When I run `cairn debug report`
      Then it succeeds
      And it says where the archive is
      And the archive holds every debug file and a header
