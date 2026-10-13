Feature: The store

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A command waits its turn rather than dropping the work

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

    @store-read-only
    Scenario: a store nothing may write to is reported
      Given I added a task titled "Write the spec"
      And the store cannot be written to
      When I run `cairn task new "Ship it"`
      Then it exits with code 3
      And it writes the reason to standard error

    @store-another-applications-database
    Scenario: another application's database named as the store is refused and left alone

      Given the store is another application's database
      When I run `cairn task list`
      Then it exits with code 3
      And the other application's database holds nothing of cairn's

    @store-nowhere-to-live
    Scenario: nowhere to put the default store is reported
      Given the environment says nothing about where anything lives
      When I run `cairn task list`
      Then it exits with code 3
      And it writes the reason to standard error
