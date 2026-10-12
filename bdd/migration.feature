Feature: Changes to the shape of the store
  What a task can carry grows over time, so the shape of the store has to grow
  with it. A store already on disk knows nothing of the latest addition, and
  changing it on the way past — while listing tasks, say — would rewrite
  somebody's data as a side effect. So cairn says what is missing and waits to
  be told, sets a backup aside before it touches anything, and can restore what
  it changed.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A store behind the cairn that opens it is not touched

    @store-behind-refused
    Scenario: a command on a store in need of a change says what to run
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn task list`
      Then it exits with code 3
      And it says the store is behind
      And it says to run `cairn migration apply`

    @store-behind-refused-by-a-surface-that-holds-the-store
    Scenario Outline: the boards say what to run too
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn <command>`
      Then it exits with code 3
      And it says the store is behind
      And it says to run `cairn migration apply`

      Examples:
        | command |
        | tui     |
        | gui     |

    @store-behind-untouched
    Scenario: refusing changes nothing about the store
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I ran a command that was refused
      When I run `cairn migration list`
      Then it says 1 change is waiting



  Rule: A change is applied when it is asked for

    @migration-apply
    Scenario: applying what is waiting brings the store up to date
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn migration apply`
      Then it succeeds
      And it names the change it applied
      And `cairn task list` lists 1 task

    @migration-apply-one
    Scenario: a single change can be named
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I apply the change that is waiting, by its number
      Then it succeeds
      And it names the change it applied

    @migration-apply-in-order
    Scenario: a change is never made before one it was written after
      Each change was written on the shape the ones before it leave, so one
      applied ahead of an earlier one gives a shape no sequence describes.

      Given a store from an older cairn, holding a task titled "Write the spec"
      And I rolled the last change back
      When I apply the last change waiting, past the one before it
      Then it fails
      And it says an earlier change is still waiting

    @migration-apply-keeps
    Scenario: what the store held is still there afterwards
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn migration apply`
      Then `cairn task list` lists task 1 titled "Write the spec"

    @migration-apply-backups
    Scenario: a backup is set aside before anything is applied
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn migration apply`
      Then `cairn backup list` lists 1 backup

    @migration-apply-nothing
    Scenario: applying when nothing is waiting says so
      Given I added a task titled "Write the spec"
      When I run `cairn migration apply`
      Then it succeeds
      And it says there is nothing waiting

    @migration-apply-nothing-json
    Scenario: having nothing to apply is answered the way it was asked
      Given I added a task titled "Write the spec"
      When I run `cairn migration apply --json`
      Then it writes JSON to standard output

    @migration-apply-already
    Scenario: applying a change that is already applied says so
      Given I added a task titled "Write the spec"
      When I run `cairn migration apply 1`
      Then it fails
      And it says the change is already applied

    @migration-apply-unknown
    Scenario: applying a change that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn migration apply 99`
      Then it fails
      And it says there is no such change

    @migration-apply-new-store
    Scenario: a store made today needs nothing applied to it
      When I run `cairn task new "Write the spec"`
      Then it succeeds
      And it writes nothing to standard error



  Rule: A label is written one way, and has no family

    @migration-labels-split-at-their-family
    Scenario: the change splits every label with a family into labels of its own
      Given a store from before labels were written one way
      When I run `cairn migration apply`
      Then task 1 carries exactly the labels "77594, ado, chantier, proxmox"
      And objective 1 carries exactly the labels "chantier, doc-infra"

    @migration-labels-written-one-way
    Scenario: the change writes every label the one way, and once
      Given a store from before labels were written one way
      When I run `cairn migration apply`
      Then task 2 carries exactly the labels "a, ado, b, c, dette-config, obj-modularite-helm, sp26-20"
      And objective 2 carries exactly the labels "ete-2026"

  Rule: What a change will do can be read before it is made

    @migration-show
    Scenario: a change says what it does, and how it is undone
      Given I added a task titled "Write the spec"
      When I run `cairn migration show 2`
      Then it succeeds
      And it names the change
      And it shows what the change does
      And it shows what undoing it does

    @migration-show-loses
    Scenario: a change says what rolling it back would lose
      Given I added a task titled "Write the spec"
      When I run `cairn migration show 2`
      Then it succeeds
      And it says what rolling the change back loses

    @migration-show-unknown
    Scenario: showing a change that does not exist says so
      Given I added a task titled "Write the spec"
      When I run `cairn migration show 99`
      Then it fails
      And it says there is no such change



  Rule: What was done to a store can be read back

    @migration-list
    Scenario: the store says what has been done to it, and when
      Given I added a task titled "Write the spec"
      When I run `cairn migration list`
      Then it succeeds
      And it names every change made to the store
      And it says when each was made

    @migration-list-waiting
    Scenario: what has not been applied is listed too, and marked
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I run `cairn migration list`
      Then it succeeds
      And it says 1 change is waiting

    @migration-list-json
    Scenario: what was done can be read by a machine
      Given I added a task titled "Write the spec"
      When I run `cairn migration list --json`
      Then it writes JSON to standard output
      And it carries every change
      And it says the store holds 1 task and 0 objectives



  Rule: The last change applied can be rolled back

    @migration-rollback
    Scenario: rolling back puts the store back the way it was
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      When I run `cairn migration rollback --yes`
      Then it succeeds
      And it names the change it undid

    @migration-holds-after-rollback
    Scenario: the store still says what it holds once rolled back
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec"
      And I rolled back the change named "objectives, and the tasks that serve them"
      When I run `cairn migration list`
      Then it succeeds
      And `cairn migration list` says the store holds 1 task and 0 objectives

    @migration-rollback-says-what-it-lost
    Scenario: rolling back names what it dropped on the way
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      When I run `cairn migration rollback --yes`
      Then it succeeds
      And it says what it lost
      And it says a backup can restore it

    @migration-rollback-keeps-what-fits
    Scenario: rolling back keeps what the older shape can still hold
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      And I added a task titled "Ship it"
      When I run `cairn migration rollback --yes`
      Then it succeeds
      And `cairn migration list` says the store holds 2 tasks and 0 objectives

    @migration-rollback-backups-first
    Scenario: a backup is set aside before anything is rolled back
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      And I deleted every backup
      When I run `cairn migration rollback --yes`
      Then `cairn backup list` lists 1 backup

    @migration-rollback-json
    Scenario: rolling back can be read by a machine
      Given I added a task titled "Write the spec"
      When I run `cairn migration rollback --yes --json`
      Then it writes JSON to standard output
      And it carries the change it rolled back

    @migration-rollback-unconfirmed
    Scenario: rolling back without confirming refuses
      Given a store from an older cairn, holding a task titled "Write the spec"
      And I applied what was waiting
      When I run `cairn migration rollback`
      Then it fails
      And it says how to confirm

    @migration-carries-what-a-task-already-served
    Scenario: the objective a task already served is not lost on the way
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I rolled back the change named "a task can serve more than one objective"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn task show 1` says it serves objective 1 alone

    @migration-rolling-back-keeps-the-smallest-objective
    Scenario: rolled back, a task keeps the first objective it served
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "2" and "1"
      And I rolled back the change named "a task can serve more than one objective"
      When I run `cairn migration apply`
      Then it succeeds
      And `cairn task show 1` says it serves objective 1 alone
