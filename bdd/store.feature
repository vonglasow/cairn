Feature: The store

  Everything cairn keeps lives in one file. Who may hold it, and what happens
  when it cannot be used at all, is what this file says.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A command waits its turn rather than dropping the work

    One person, one machine — and still more than one cairn at a time: a
    command typed while another is running, a shell loop, and before long a
    server answering on the side while the terminal is still in use.

    A store held by somebody else is not a request cairn cannot make sense
    of. It is a moment to wait through, and waiting is what a command does,
    because the alternative is work that was accepted and never written.

    Making one is the same moment as holding one. A file that does not exist
    yet is made by whoever gets there first, and the others wait for it
    exactly as they wait for a write — measured, because they did not:
    switching a fresh file to the journal cairn needs takes a lock the
    ordinary wait does not cover.

    @store-busy-being-made
    Scenario: commands making the store at the same time all land
      When 4 commands add a task at the same time
      Then `cairn task list` lists 4 tasks

    @store-busy-waits
    Scenario: commands writing at the same time all land
      Given I added a task titled "Write the spec"
      When 4 commands add a task at the same time
      Then `cairn task list` lists 5 tasks

    @store-busy-reading
    Scenario: reading while something else writes still answers
      Given I added a task titled "Write the spec"
      When 4 commands read and write at the same time
      Then it succeeds
      And `cairn task list` lists 3 tasks

  Rule: A store that cannot be used at all is reported as such

    Code 3 is the answer that says "the store, not your request". A script
    reads it to tell a mistake it should fix from a machine it should look
    at, so the two must never be confused.

    @store-read-only
    Scenario: a store nothing may write to is reported
      Given I added a task titled "Write the spec"
      And the store cannot be written to
      When I run `cairn task add "Ship it"`
      Then it exits with code 3
      And it writes the reason to standard error

    @store-nowhere-to-live
    Scenario: nowhere to put the default store is reported
      Given the environment says nothing about where anything lives
      When I run `cairn task list`
      Then it exits with code 3
      And it writes the reason to standard error
