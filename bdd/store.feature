Feature: The store

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A command waits its turn rather than dropping the work

    @store-busy-being-made
    Scenario: commands making the store at the same time all land
      When 4 commands add a task at the same time
      Then `cairn task list` lists 4 tasks

  Rule: A store that cannot be used at all is reported as such

    @store-read-only
    Scenario: a store nothing may write to is reported
      Given I added a task titled "Write the spec"
      And the store cannot be written to
      When I run `cairn task new "Ship it"`
      Then it exits with code 3
      And it writes the reason to standard error
