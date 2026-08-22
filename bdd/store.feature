Feature: More than one cairn at a time

  One person, one machine — and still more than one cairn at a time: a
  command typed while another is running, a shell loop, and before long a
  server answering on the side while the terminal is still in use. They all
  reach for the same file.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A command waits its turn rather than dropping the work

    A store held by somebody else is not a request cairn cannot make sense
    of. It is a moment to wait through, and waiting is what a command does,
    because the alternative is work that was accepted and never written.

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
