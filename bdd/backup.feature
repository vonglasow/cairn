Feature: Backups of the store
  A store holds work nobody wants to lose, so it can be copied at any moment —
  before a change of shape, before anything risky, or simply because a backup is
  wanted. A backup is a thing in its own right: it has a number to refer to it by,
  it can be looked at before it is trusted, and nothing throws it away but a
  deliberate deletion.

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

    @backup-number-next
    Scenario: the next backup gets the next number
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup new`
      Then it numbers the backup 2

    @backup-add-says-what-it-copied
    Scenario: a backup says how much it took, and says the same later
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I run `cairn backup new`
      Then it says the backup holds 2 tasks and 0 objectives
      And `cairn backup show 1` says the backup holds 2 tasks and 0 objectives

    @backup-holds-objectives-too
    Scenario: a backup counts the objectives it holds
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec"
      When I run `cairn backup new`
      Then it says the backup holds 1 task and 1 objective

    @backup-holds-an-older-shape
    Scenario: a backup of a store one change behind holds no objectives
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn backup new`
      Then it says the backup holds 1 task and 0 objectives

    @backup-holds-json
    Scenario: what a backup holds can be read by a machine
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup show 1 --json`
      Then it writes JSON to standard output
      And it carries 1 task and 1 objective held

    @backup-add-not-a-store
    Scenario: backing up what is not a cairn store is refused
      Given the store holds something that is not a cairn store
      When I run `cairn backup new`
      Then it exits with code 3
      And it names the store
      And `cairn backup list` says there is no backup

    @backup-holds-what-was-there
    Scenario: a backup holds what the store held when it was made
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup show 1`
      Then it says the backup holds 1 task and 0 objectives




  Rule: Backups can be listed and read

    @backup-list-empty
    Scenario: listing when no backup was ever made says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup list`
      Then it succeeds
      And it says there is no backup

    @backup-list
    Scenario: every backup is listed with its number
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I made a backup of the store
      When I run `cairn backup list`
      Then it lists 2 backups
      And it lists backup 1
      And it lists backup 2

    @backup-show
    Scenario: a backup says when it was made and what it holds
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup show 1`
      Then it succeeds
      And it says when the backup was made
      And it says the backup holds 1 task and 0 objectives

    @backup-show-unknown
    Scenario: showing a backup that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup show 99`
      Then it fails
      And it says there is no such backup

  Rule: A backup can be deleted

    @backup-delete
    Scenario: deleting a backup removes it
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup delete 1 --yes`
      Then it succeeds
      And `cairn backup list` says there is no backup

    @backup-delete-unconfirmed
    Scenario: deleting a backup without confirming refuses
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup delete 1`
      Then it fails
      And it says how to confirm

    @backup-delete-json
    Scenario: deleting a backup can be read by a machine
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup delete 1 --yes --json`
      Then it writes JSON to standard output
      And it carries the number 1

    @backup-delete-unknown
    Scenario: deleting a backup that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup delete 99 --yes`
      Then it fails
      And it says there is no such backup

    @backup-number-not-reused
    Scenario: a number freed by a deletion is not handed out again
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I deleted backup 1
      When I run `cairn backup new`
      Then it numbers the backup 2




  Rule: A backup can be put back

    @backup-restore
    Scenario: putting a backup back returns the store to what it held
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I added a task titled "Ship it"
      When I run `cairn backup restore 1 --yes`
      Then it succeeds
      And `cairn task list` lists 1 task

    @backup-restore-under-an-open-board
    Scenario: a board open while a backup is put back writes to what was put back
      A board holds the store open for as long as it is on screen. Putting a
      backup back replaces the file under it, and a board still writing to the
      file that was replaced loses what it writes without a word.

      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And a board is open on the store
      When I run `cairn backup restore 1 --yes`
      And the board adds a task titled "Added on the board"
      Then `cairn task list` lists 2 tasks
      And `cairn task list` lists task 2 titled "Added on the board"

    @backup-restore-latest
    Scenario: with no number, the most recent backup comes back
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I added a task titled "Ship it"
      And I made a backup of the store
      And I added a task titled "And another"
      When I run `cairn backup restore --yes`
      Then it succeeds
      And `cairn task list` lists 2 tasks

    @backup-restore-backups-first
    Scenario: what is about to be replaced is copied first
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      And I added a task titled "Ship it"
      When I run `cairn backup restore 1 --yes`
      Then it succeeds
      And `cairn backup list` lists 2 backups

    @backup-restore-json
    Scenario: putting a backup back can be read by a machine
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup restore 1 --yes --json`
      Then it writes JSON to standard output
      And it carries the number 1

    @backup-restore-unconfirmed
    Scenario: putting a backup back without confirming refuses
      Given I added a task titled "Write the spec"
      And I made a backup of the store
      When I run `cairn backup restore 1`
      Then it fails
      And it says how to confirm

    @backup-restore-unknown
    Scenario: putting back a backup that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup restore 99 --yes`
      Then it fails
      And it says there is no such backup

    @backup-restore-none
    Scenario: putting back when no backup was ever made says so
      Given I added a task titled "Write the spec"
      When I run `cairn backup restore --yes`
      Then it fails
      And it says there is no backup to put back

  Rule: A store travels as a file

    A store moves to another machine, or comes back from one, as a file put
    wherever a person wants it — not only into the backups folder beside it.
    Bringing one in replaces the store, so it is checked first, and what the
    store held is backed up before it goes.

    @backup-export
    Scenario: the store is written to a file of my choosing
      Given I added a task titled "Write the spec"
      When I export the store to a file of my own
      Then it succeeds
      And it says the file written holds 1 task and 0 objectives
      And that file is a store holding the task "Write the spec"

    @backup-export-never-overwrites
    Scenario: exporting onto a file that is already there is refused
      Given I added a task titled "Write the spec"
      And a file of my own that is not a cairn store
      When I export the store to that file
      Then it fails
      And that file is as it was

    @backup-import
    Scenario: a store brought in from a file replaces this one, backed up first
      Given I added a task titled "Write the spec"
      And a file of my own holding a store with a task titled "Ship it"
      When I import that file
      Then it succeeds
      And `cairn task list` lists 1 task
      And `cairn task list` lists task 1 titled "Ship it"
      And `cairn backup list` lists 2 backups

    @backup-import-not-a-store
    Scenario: bringing in what is not a cairn store is refused, and changes nothing
      Given I added a task titled "Write the spec"
      And a file of my own that is not a cairn store
      When I import that file
      Then it fails
      And `cairn task list` lists task 1 titled "Write the spec"
      And `cairn backup list` says there is no backup

    @backup-import-unconfirmed
    Scenario: bringing in a store without confirming refuses
      Given I added a task titled "Write the spec"
      And a file of my own holding a store with a task titled "Ship it"
      When I run the import of that file without confirming
      Then it fails
      And it says how to confirm
