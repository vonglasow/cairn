Feature: Tasks

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A task needs a title

    @task-title-empty
    Scenario: an empty title is no title
      When I run `cairn task new ""`
      Then it fails
      And it says a task needs a title
      And it shows the title I typed

  Rule: A task is numbered so it can be referred to later

    @task-number-first
    Scenario: the first task is numbered 1
      When I run `cairn task new "Write the spec"`
      Then it numbers the task 1

  Rule: A task carries only what it was given

    @task-defaults
    Scenario: a new task carries nothing else
      When I run `cairn task new "Write the spec"`
      Then the task has no description
      And the task has no labels
      And the task has no deadline
      And the task has no estimate
      And the task has no time spent
      And the task carries the priority "normal"
      And the task is in the "backlog" column

  Rule: Tasks can be listed

    @task-new-description-spelled-whole
    Scenario: --description is --desc spelled whole
      When I run `cairn task new "Write the spec" --description "the one with the diagrams"`
      Then `cairn task show 1` says "the one with the diagrams"

  Rule: A listing can be narrowed to what is wanted

    @task-list-narrowed-by-state
    Scenario: only the tasks in one column
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      When I run `cairn task list --state todo`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

  Rule: A task can be read back by its number

    @task-show
    Scenario: showing a task gives back what it carries
      Given I added a task titled "Write the spec" with the description "the one"
      When I run `cairn task show 1`
      Then it reports a task titled "Write the spec"
      And the task carries the description "the one"

  Rule: A task is looked for by what somebody remembers of it

    @task-search-leaves-archived-out-until-asked-for
    Scenario: what was archived is not looked in until it is asked for
      Given I added a task titled "Write the spec"
      And I archived task 1
      When I run `cairn task search spec`
      Then it says there is no task
      When I run `cairn task search spec --archived`
      Then it lists 1 task

    @task-search-needs-something-to-look-for
    Scenario: looking for nothing at all is refused
      When I run `cairn task search`
      Then it fails
      And it says how to use it

  Rule: A task can be changed

    @task-edit-nothing
    Scenario: editing without naming anything to change says so
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1`
      Then it fails
      And it says what it can change

  Rule: A label can be added or taken away without naming the others

    @task-label-list
    Scenario: several labels can be given in one flag
      When I run `cairn task new "Write the spec" --label "spec,team"`
      Then the task carries exactly 2 labels
      And the task carries the labels "spec" and "team"

  Rule: A task can say what blocks it

    @task-blocker-itself
    Scenario: a task cannot be what is blocking it
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --blocker 1`
      Then it fails
      And it says a task cannot block itself

    @task-blocker-a-task-nobody-added
    Scenario: a blocker has to be a task that exists
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --blocker 9`
      Then it fails
      And it says there is no such task

  Rule: Two tasks can be linked

    @task-link-at-creation
    Scenario: a task opened with a link says so, and so does the other
      Given I added a task titled "Ship it"
      When I run `cairn task new "Ship it again" --link 1`
      Then it succeeds
      And `cairn task show 2` says "linked       #1"
      And `cairn task show 1` says "linked       #2"

  Rule: A blocked task is not picked back up by accident

    @task-blocked-cannot-go-back-to-work
    Scenario Outline: a task still blocked cannot be claimed as yours to do
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 2 --state <column>`
      Then it fails
      And it says what is in the way and how to take it away

      Examples: the two columns that claim the work is yours
        | column  |
        | todo    |
        | current |

  Rule: A blocked task says when nothing is in its way any more

    But it says so, because otherwise a blocked task waits forever on somebody

    @task-blocked-nothing-in-the-way
    Scenario: every task blocking it is done, and the listing says so
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 1 --state done`
      And I run `cairn task list`
      Then it marks "Ship it" as having nothing in its way

  Rule: A task gives the facts that help choose it, and ranks nothing

    @task-says-how-many-it-blocks
    Scenario: a task says how many open tasks it is in the way of
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Announce it"
      And I added a task titled "Thank the reviewers"
      And I edited task 2 with the blocker "1"
      And I edited task 3 with the blocker "1"
      And I edited task 4 with the blocker "1"
      And I edited task 4 with the state "done"
      When I run `cairn task list --json --fields number --fields blocks`
      Then the JSON row of task 1 gives "blocks" as 2
      And the JSON row of task 2 gives "blocks" as 0

  Rule: A task can be deleted

    @task-delete
    Scenario: deleting a task removes it
      Given I added a task titled "Write the spec"
      When I run `cairn task delete 1 --yes`
      Then it succeeds
      And it says task 1 was deleted
      And `cairn task list` says there is no task

  Rule: Time spent on a task is the time it was in focus and the time added by hand

    @task-spent-counts-focus
    Scenario: the time a task was in focus is time spent on it
      Given a store whose task 1 was in focus from 07:00 to 08:30 UTC on 2026-09-29
      When I run `cairn task show 1`
      Then the task carries the time spent "1h30m"

  Rule: A label is put on everything a narrowing reaches

    @task-label-needs-something-to-go-on
    Scenario: labelling with nothing to narrow by is refused
      Given I added a task titled "Spike Argo"
      When I run `cairn task label sprint-41`
      Then it fails
      And it says what it needs to go on

  Rule: A task is archived rather than deleted

    @task-unarchive-what-is-on-the-board
    Scenario: bringing back what never went away is refused, the way archiving twice is
      Given I added a task titled "Migration Jenkins"
      When I run `cairn task unarchive 1`
      Then it fails
      And it says the task is on the board

  Rule: What went done before the period under way is archived on its own

    @archive-what-went-done-before-the-period
    Scenario: what went done before the period under way is archived once the tasks are listed
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I added a task titled "Spike Argo" with the state "todo"
      And archiving every 14 days, counted from 10 days ago
      And it is 10 days from now
      When I run `cairn task list`
      Then it lists 1 task
      And task 1 is archived
      And task 2 is not archived

    @archive-periods-last-14-days-unless-told
    Scenario: a period whose length is not said lasts 14 days
      Given I added a task titled "Migration Jenkins" with the state "done"
      And archiving counted from 10 days ago, every so many days not said
      And it is 4 days from now
      When I run `cairn task list`
      Then task 1 is archived

    @archive-nothing-before-the-count-starts
    Scenario: before the day the count starts, nothing is archived on its own
      Given I added a task titled "Migration Jenkins" with the state "done"
      And archiving every 14 days, counted from 3 days from now
      When I run `cairn task list`
      Then it lists 1 task
      And task 1 is not archived

  Rule: A task is split into steps, which are not tasks

    @task-step-has-no-number
    Scenario: a step takes no number of its own
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      When I run `cairn task new "Announce it"`
      Then it numbers the task 2

  Rule: A task is not done while one of its steps is open

    @task-done-refused-while-a-step-is-open
    Scenario: a task is not finished while a step of it is still open
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --state done`
      Then it fails
      And it says step 1 is in the way
      And task 1 is in the column "backlog"

  Rule: A task's steps are written with it, in one act

    @task-new-with-its-steps
    Scenario: a task is made with its steps
      When I run `cairn task new "Ship it" --steps -` with "[ ] Write the spec" and "[x] Plan it" on standard input
      Then it succeeds
      And `cairn task show 1` names the step "Write the spec" as not done
      And `cairn task show 1` names the step "Plan it" as done

  Rule: A task says how many of its steps are done

    @task-progress
    Scenario: a task says how many of its steps are done
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the tests" to task 1
      When I run `cairn step edit 1 1 --done`
      And I run `cairn task show 1`
      Then it says 1 of 2 steps are done

  Rule: A task outlives the command that created it

    @task-survives
    Scenario: a task added by one command is found by the next
      Given I added a task titled "Write the spec"
      And cairn has exited
      When I run `cairn task list`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

  Rule: What goes into the store comes back out unchanged

    @task-roundtrip-text
    Scenario Outline: text survives the store
      Given I added a task titled "<title>"
      And cairn has exited
      When I run `cairn task show 1`
      Then it reports a task titled "<title>"

      Examples: encoding is not the place to lose data
        | title                    | what it checks  |
        | Écrire la spéc — enfin ! | accents, dashes |
        | 🧗 climb                 | an emoji        |
        | 90% done, right?         | punctuation     |

  Rule: Getting the command wrong is told, never guessed

    @task-usage-refused
    Scenario Outline: the command says what it needed
      When I run `<command>`
      Then it fails
      And it does not crash
      And it says how to use it

      Examples: what someone types the first time
        | command                       | what is wrong             |
        | cairn task new                | no title at all           |
        | cairn task new Write the spec | quotes forgotten          |
        | cairn task new "a" "b"        | two titles                |
        | cairn task addd "Write"       | a typo in the subcommand  |
        | cairn task                    | the group with no command |
        | cairn task show               | no number given           |
        | cairn task edit               | no number given           |
        | cairn task delete             | no number given           |

  Rule: It answers the same way every time

    @task-exit-success
    Scenario: what worked says nothing on the error stream
      When I run `cairn task new "Write the spec"`
      Then it exits with code 0
      And it writes the task to standard output
      And it writes nothing to standard error

  Rule: The output can be read by a machine

    @task-json
    Scenario: asking for machine-readable output
      When I run `cairn task new "Write the spec" --json`
      Then it writes JSON to standard output
      And it carries the title "Write the spec"
      And it carries an identifier
