Feature: Backups of the store

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A backup of the store can be made

    @backup-add
    Scenario: a backup can be made whenever I want one
      Given I added a task titled "Write the spec"
      When I run `cairn backup new`
      Then it succeeds
      And it numbers the backup 1

  Rule: Backups can be listed and read

    @backup-show-unknown
    Scenario: showing a backup that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup show 99`
      Then it fails
      And it says there is no such backup

  Rule: A backup can be deleted

    @backup-delete-unknown
    Scenario: deleting a backup that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup delete 99 --yes`
      Then it fails
      And it says there is no such backup

  Rule: A backup can be restored

    @backup-restore
    Scenario: putting a backup back returns the store to what it held
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I added a task titled "Ship it"
      When I run `cairn backup restore 1 --yes`
      Then it succeeds
      And `cairn task list` lists 1 task
