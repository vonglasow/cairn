Feature: Objectives

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: An objective needs a title

    @objective-add
    Scenario: an objective is added and numbered
      When I run `cairn objective new "Ship cairn 1.0"`
      Then it succeeds
      And it numbers the objective 1

    @objective-title-one-line
    Scenario Outline: a title is one line
      Given a title with a <character> in it
      When I add an objective with it
      Then it fails
      And it says a title is one line

      Examples: what a card and a listing cannot show
        | character  |
        | line break |
        | tab        |

    @objective-title-empty
    Scenario: an empty title is no title
      When I run `cairn objective new ""`
      Then it fails
      And it says an objective needs a title
      And it shows the title I typed

  Rule: An objective is numbered so it can be referred to later

    @objective-number-next
    Scenario: the next objective gets the next number
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective new "Learn Go properly"`
      Then it numbers the objective 2

    @objective-number-apart-from-tasks
    Scenario: objectives and tasks are numbered apart
      Given I added a task titled "Write the spec"
      When I run `cairn objective new "Ship cairn 1.0"`
      Then it numbers the objective 1

  Rule: A task can be put under an objective

    @objective-task-added-under
    Scenario: a task can be added under an objective
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn task new "Write the spec" --objective 1`
      Then the task carries the objective "1"

    @objective-task-none
    Scenario: a task belongs to no objective unless it is told to
      When I run `cairn task new "Write the spec"`
      Then the task has no objective

    @objective-task-attached-later
    Scenario: a task can be put under an objective after the fact
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec"
      When I run `cairn task edit 1 --objective 1`
      Then the task carries the objective "1"

    @objective-task-detached
    Scenario: an objective given empty sets the task loose
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn task edit 1 --objective ""`
      Then the task has no objective
      And `cairn objective show 1` says nothing is under it

  Rule: A task can serve more than one objective

    @objective-task-serves-two
    Scenario: a task can be given two objectives at once
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      When I run `cairn task new "Write the spec" --objective 1 --objective 2`
      Then it succeeds
      And `cairn task show 1` says it serves objectives 1 and 2

    @objective-task-serves-one-more
    Scenario: another objective can be added without naming the first
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn task edit 1 --objective +2`
      Then `cairn task show 1` says it serves objectives 1 and 2

    @objective-task-serves-one-less
    Scenario: one objective can be taken away without naming the other
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      When I run `cairn task edit 1 --objective -1`
      Then `cairn task show 1` says it serves objective 2 alone

    @objective-task-serves-these-instead
    Scenario: a plain number says what the set becomes, not what joins it
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      When I run `cairn task edit 1 --objective 2`
      Then `cairn task show 1` says it serves objective 2 alone

    @objective-task-serves-it-once
    Scenario: the same objective named twice is served once
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn task new "Write the spec" --objective 1 --objective 1`
      Then `cairn task show 1` says it serves objective 1 alone

    @objective-counts-a-task-it-shares
    Scenario: an objective counts a task that also serves another
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      When I run `cairn objective show 1`
      Then it says 1 task is under it
      And `cairn objective show 2` says 1 task is under it

    @objective-progress-counts-a-shared-task-on-both-sides
    Scenario: work done on a shared task advances both objectives
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      And I edited task 1 with the state "done"
      When I run `cairn objective show 1`
      Then it says the objective is 100% done
      And `cairn objective show 2` says the objective is 100% done

    @objective-task-detached-from-one-stays-under-the-other
    Scenario: taking one away leaves the other holding it
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      When I run `cairn task edit 1 --objective -1`
      Then `cairn objective show 1` says nothing is under it
      And `cairn objective show 2` says 1 task is under it

    @objective-task-json-carries-every-objective
    Scenario: what a task serves crosses to a machine as a list
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objectives "1" and "2"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it says the task serves objectives 1 and 2

  Rule: An objective knows what was put under it

    @objective-carries-the-task
    Scenario: an objective names the tasks put under it
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective show 1`
      Then it lists task 1 titled "Write the spec"

    @objective-shows-the-task-under-it
    Scenario: a task is shown under the objective it was put under
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective show 1`
      Then it reports an objective titled "Ship cairn 1.0"
      And task 1 is shown under objective 1

    @objective-carries-nothing
    Scenario: an objective with nothing under it says so
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective show 1`
      Then it says nothing is under it

  Rule: An objective says how much of it is done

    @objective-progress
    Scenario: an objective says how much of it is done
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      When I run `cairn objective show 1`
      Then it says 0 of 2 tasks are done

    @objective-progress-follows
    Scenario: what is done is counted as it is done
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      When I run `cairn task edit 1 --state done`
      And I run `cairn objective show 1`
      Then it says 1 of 2 tasks are done

    @objective-progress-keeps-what-was-archived
    Scenario: archived work still counts towards what it served

      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      And I edited task 1 with the state "done"
      When I run `cairn task archive --state done --yes`
      And I run `cairn objective show 1`
      Then it says 1 of 2 tasks are done

    @objective-progress-counts-steps
    Scenario: a step is counted apart, and still counts towards the share
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added the step "Write the tests" to task 1
      When I run `cairn step edit 1 1 --done`
      And I run `cairn objective show 1`
      Then it says 0 of 1 tasks and 1 of 1 steps are done
      And it says the objective is 50% done

    @objective-progress-percent
    Scenario: how much is done is also given as a share
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      And I added a task titled "Ship it" with the objective "1"
      When I run `cairn task edit 1 --state done`
      And I run `cairn objective show 1`
      Then it says the objective is 33% done

    @objective-progress-never-rounds-to-done
    Scenario: almost everything done is not everything done
      Given I added an objective titled "Ship cairn 1.0"
      And I added 300 tasks under it, of which 299 are done
      When I run `cairn objective show 1`
      Then it says the objective is 99% done

    @objective-progress-never-rounds-to-nothing
    Scenario: something done is not nothing done
      Given I added an objective titled "Ship cairn 1.0"
      And I added 300 tasks under it, of which 1 is done
      When I run `cairn objective show 1`
      Then it says the objective is 1% done

    @objective-progress-of-nothing
    Scenario: an objective with nothing under it counts nothing
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective show 1`
      Then it does not say how much is done

  Rule: An objective given to a task is one that exists

    @objective-unknown
    Scenario: putting a task under an objective that does not exist is refused
      When I run `cairn task new "Write the spec" --objective 99`
      Then it fails
      And it says there is no such objective

    @objective-not-a-number
    Scenario Outline: an objective that is not a number says so
      Given I added an objective titled "Ship cairn 1.0"
      When I run `<command>`
      Then it fails
      And it says an objective number is expected

      Examples: everywhere a number is read
        | command                                          |
        | cairn task new "Write the spec" --objective spec |
        | cairn objective show spec                        |
        | cairn objective edit spec --title "Nope"         |
        | cairn objective delete spec --yes                |

    @objective-one-of-several-unknown
    Scenario: one objective that does not exist turns down the whole set
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn task edit 1 --objective +99`
      Then it fails
      And it says there is no such objective
      And `cairn task show 1` says it serves objective 1 alone

    @objective-taken-away-that-was-never-there
    Scenario: taking away one it never served changes nothing
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn task edit 1 --objective -99`
      Then it succeeds
      And `cairn task show 1` says it serves objective 1 alone

    @objective-both-ways-at-once-refused
    Scenario: naming what the set becomes and how it moves at once is refused
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec"
      When I run `cairn task edit 1 --objective 1 --objective +2`
      Then it fails
      And it says the objectives are either replaced or changed, not both

    @objective-sign-at-creation-refused
    Scenario: a sign says how a set moves, and a new task has none
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn task new "Write the spec" --objective +1`
      Then it fails
      And it says an objective is added or taken away only when changing a task

  Rule: An objective can be listed, read back and changed

    @objective-list
    Scenario: listing shows every objective with its number
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Learn Go properly"
      When I run `cairn objective list`
      Then it lists 2 objectives
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-edit-title
    Scenario: changing the title
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective edit 1 --title "Ship cairn 1.0, for real"`
      Then it reports an objective titled "Ship cairn 1.0, for real"

    @objective-edit-blank-title
    Scenario: a title cannot be emptied by editing
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective edit 1 --title ""`
      Then it fails
      And it says an objective needs a title

    @objective-label-list
    Scenario: several labels can be given in one flag
      When I run `cairn objective new "Ship cairn 1.0" --label "perso,spec"`
      Then the objective carries exactly 2 labels
      And the objective carries the label "perso"

    @objective-label-add
    Scenario: a label given with a plus joins the others
      Given I added an objective titled "Ship cairn 1.0" with the labels "perso" and "spec"
      When I run `cairn objective edit 1 --label +urgent`
      Then the objective carries exactly 3 labels
      And the objective carries the label "urgent"

    @objective-label-remove
    Scenario: a label given with a minus leaves the others alone
      Given I added an objective titled "Ship cairn 1.0" with the labels "perso" and "spec"
      When I run `cairn objective edit 1 --label -spec`
      Then the objective carries exactly 1 label
      And the objective carries the label "perso"

  Rule: A listing of objectives can be narrowed the same way

    @objective-list-narrowed-by-state
    Scenario: only the objectives standing where I asked
      Given I added an objective titled "Ship cairn 1.0" with the state "on-hold"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective list --state on-hold`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-list-narrowed-by-a-label
    Scenario: only the objectives carrying one label
      Given I added an objective titled "Ship cairn 1.0" with the labels "team" and "urgent"
      And I added an objective titled "Ship the docs" with the labels "team" and "later"
      When I run `cairn objective list --label urgent`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

  Rule: An objective is looked for by anything it carries

    @objective-search-by-a-word
    Scenario: a word from the title is the way back to a number
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective search cairn`
      Then it lists objective 1 titled "Ship cairn 1.0"

    @objective-search-everything-named-is-found
    Scenario: everything named has to be there, wherever it is
      Given I added an objective titled "Ship cairn 1.0" with the description "the first"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective search first ship`
      Then it lists 1 objective

    @objective-search-by-its-number
    Scenario: the number is the way back when it is what somebody has
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective search 2`
      Then it lists 1 objective
      And it lists objective 2 titled "Ship the docs"

    @objective-search-reaches-where-it-stands
    Scenario: where it stands is looked in like anything else it carries
      Given I added an objective titled "Ship cairn 1.0" with the state "on-hold"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective search hold`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-search-reaches-the-priority
    Scenario: the priority is looked in like anything else it carries
      Given I added an objective titled "Ship cairn 1.0" with the priority "high"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective search high`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-search-reaches-a-label
    Scenario: a label is looked in like anything else it carries
      Given I added an objective titled "Ship cairn 1.0" with the labels "team" and "urgent"
      And I added an objective titled "Ship the docs"
      When I run `cairn objective search urgent`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-search-finds-nothing
    Scenario: looking for what is not there says so
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective search nowhere`
      Then it succeeds
      And it says there is no objective

  Rule: An objective carries only what it was given

    @objective-defaults
    Scenario: a new objective carries nothing else
      When I run `cairn objective new "Ship cairn 1.0"`
      Then the objective has no description
      And the objective has no labels
      And the objective has no deadline
      And the objective has no estimate
      And the objective carries the priority "normal"
      And the objective is active

    @objective-attribute-accepted
    Scenario Outline: an attribute given is an attribute carried
      When I run `cairn objective new "Ship cairn 1.0" <flag> <value>`
      Then the objective carries the <attribute> "<stored>"

      Examples: each attribute, and how it comes back
        | flag       | value      | attribute   | stored     |
        | --desc     | the one    | description | the one    |
        | --label    | perso      | label       | perso      |
        | --deadline | 2026-12-31 | deadline    | 2026-12-31 |
        | --estimate | 90m        | estimate    | 1h30m      |

    @objective-attribute-refused
    Scenario Outline: an attribute it cannot make sense of is refused
      When I run `cairn objective new "Ship cairn 1.0" <flag> <value>`
      Then it fails
      And it says what it expects instead of "<value>"

      Examples: nothing is guessed
        | flag       | value   | what is wrong                      |
        | --priority | urgent  | not one of the priorities          |
        | --state    | todo    | a task's state, not an objective's |
        | --deadline | soon    | not a date                         |
        | --estimate | ages    | not a duration                     |
        | --estimate | -30m    | a negative duration                |
        | --estimate | 500ms   | shorter than a second              |

    @objective-priorities
    Scenario Outline: every priority an objective can carry
      When I run `cairn objective new "Ship cairn 1.0" --priority <priority>`
      Then the objective carries the priority "<priority>"

      Examples: the four, and no others
        | priority |
        | low      |
        | normal   |
        | high     |
        | critical |

    @objective-states
    Scenario Outline: every state an objective can be in
      When I run `cairn objective new "Ship cairn 1.0" --state <state>`
      Then the objective is <state>

      Examples: the four, and no others
        | state   |
        | active  |
        | on-hold |
        | reached |
        | dropped |

  Rule: An objective can be deleted

    @objective-delete
    Scenario: deleting an objective removes it
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective delete 1 --yes`
      Then it succeeds
      And `cairn objective list` says there is no objective

    @objective-delete-unconfirmed
    Scenario: deleting without confirming refuses
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective delete 1`
      Then it fails
      And it says how to confirm
      And `cairn objective show 1` still reports an objective titled "Ship cairn 1.0"

    @objective-delete-with-tasks
    Scenario: an objective with tasks under it is not deleted
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective delete 1 --yes`
      Then it fails
      And it says the objective carries tasks
      And `cairn objective show 1` still reports an objective titled "Ship cairn 1.0"

    @objective-delete-several
    Scenario: several objectives are deleted in one act
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Keep it honest"
      When I run `cairn objective delete 1 --also 2 --yes`
      Then it succeeds
      And `cairn objective list` says there is no objective

    @objective-delete-several-or-none
    Scenario: several objectives are deleted all together, or none of them is
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec" with the objective "2"
      When I run `cairn objective delete 1 --also 2 --yes`
      Then it fails
      And it says the objective carries tasks
      And `cairn objective show 1` still reports an objective titled "Ship cairn 1.0"

  Rule: An objective says the time spent on it and what is left of it

    @objective-says-the-time-spent-on-it
    Scenario: the time spent on an objective is the time spent on every task serving it
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I edited task 1 with the time spent "1h30m"
      And I edited task 2 with the time spent "45m"
      When I run `cairn objective show 1`
      Then the objective carries the time spent "2h15m"

    @objective-says-what-is-left
    Scenario: what is left is every task not done, the next one due named
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I added a task titled "Ship it" with the objective "1"
      And I edited task 1 with the state "done"
      And I edited task 2 with the deadline "2026-11-20"
      And I edited task 3 with the deadline "2026-11-05"
      When I run `cairn objective show 1`
      Then the objective carries the left "2 tasks, the next due 2026-11-05 (#3)"

    @objective-with-nothing-left-says-so
    Scenario: an objective whose every task is done has nothing left
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I edited task 1 with the state "done"
      When I run `cairn objective show 1`
      Then the objective carries the left "nothing"

    @objective-list-says-how-far-along
    Scenario: the list says how far along each objective is
      Given I added an objective titled "Ship cairn 1.0"
      And I added an objective titled "Ship the docs"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I edited task 1 with the state "done"
      When I run `cairn objective list`
      Then the row for objective 1 says "50% done"
      And the row for objective 2 says nothing of how far along it is

    @objective-json-says-the-time-and-what-is-left
    Scenario: a machine reads the time spent on an objective, what is left of it, and how far along the list is
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I edited task 1 with the time spent "1h30m"
      And I edited task 1 with the state "done"
      When I run `cairn objective show 1 --json`
      Then the JSON objective carries the time spent "1h30m" and leaves the tasks "2"
      When I run `cairn objective list --json`
      Then the JSON list gives objective 1 at 50% done

  Rule: An objective is reviewed over a period

    @objective-report-says-the-time-month-by-month
    Scenario: the time spent on an objective's tasks in a period is said month by month
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I added a task titled "Elsewhere"
      And I ran `cairn timesheet add 1 2h --on 2026-07-30`
      And I ran `cairn timesheet add 1 1h30m --on 2026-08-04`
      And I ran `cairn timesheet add 2 45m --on 2026-09-10`
      And I ran `cairn timesheet add 3 3h --on 2026-09-10`
      When I run `cairn objective report 1 --since 2026-08-01 --until 2026-09-30`
      Then it succeeds
      And the objective carries the time spent "2h15m in the period, 4h15m in all"
      And the month "2026-08" says "1h30m"
      And the month "2026-09" says "45m"
      And it says nothing of the month "2026-07"

    @objective-report-names-what-was-done-in-the-period
    Scenario: the tasks done in the period are named, and what is left is said
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Draw the board" with the objective "1"
      And I edited task 1 with the state "done"
      When I run `cairn objective report 1 --since 2026-10-01`
      Then the report names task 1 as done in the period
      And the objective carries the left "1 task, none due"

    @objective-report-before-the-work-names-nothing-done
    Scenario: a period that ends before anything was done names nothing done in it
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I edited task 1 with the state "done"
      When I run `cairn objective report 1 --since 2026-08-01 --until 2026-08-31`
      Then the report names nothing as done in the period

    @objective-report-needs-the-first-day
    Scenario: a review needs the day its period starts
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective report 1`
      Then it fails

    @objective-report-reads-by-a-machine
    Scenario: a review can be read by a machine
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I ran `cairn timesheet add 1 1h30m --on 2026-08-04`
      And I edited task 1 with the state "done"
      When I run `cairn objective report 1 --since 2026-08-01 --json`
      Then the JSON report gives the month "2026-08" at "1h30m" and task 1 done

  Rule: An objective outlives the command that created it

    @objective-survives
    Scenario: an objective added by one command is found by the next
      Given I added an objective titled "Ship cairn 1.0"
      And cairn has exited
      When I run `cairn objective list`
      Then it lists 1 objective
      And it lists objective 1 titled "Ship cairn 1.0"

    @objective-link-survives
    Scenario: what was put under an objective is still there afterwards
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And cairn has exited
      When I run `cairn objective show 1`
      Then it lists task 1 titled "Write the spec"

    @objective-attributes-survive
    Scenario: what an objective carries is still there afterwards
      Given I added an objective titled "Ship cairn 1.0" with the priority "high"
      And cairn has exited
      When I run `cairn objective show 1`
      Then the objective carries the priority "high"

    @objective-state-survives
    Scenario: the state an objective was put in is still there afterwards
      Given I added an objective titled "Ship cairn 1.0" with the state "on-hold"
      And cairn has exited
      When I run `cairn objective show 1`
      Then the objective is on-hold

  Rule: An objective can be read by a machine

    @objective-json
    Scenario: asking for machine-readable output
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      When I run `cairn objective show 1 --json`
      Then it writes JSON to standard output
      And it carries the title "Ship cairn 1.0"
      And it carries the tasks under it

    @objective-json-list
    Scenario: listing can be read by a machine
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective list --json`
      Then it writes JSON to standard output
      And it carries 1 objective

    @objective-json-progress
    Scenario: how much is done can be read by a machine
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Write the tests" with the objective "1"
      And I added the step "Ship it" to task 2
      When I run `cairn task edit 1 --state done`
      And I run `cairn objective show 1 --json`
      Then it writes JSON to standard output
      And it carries 1 of 2 tasks done
      And it carries 0 of 1 steps done
      And it carries the share 33

    @objective-json-progress-of-nothing
    Scenario: an objective with nothing under it carries no progress
      Given I added an objective titled "Ship cairn 1.0"
      When I run `cairn objective show 1 --json`
      Then it writes JSON to standard output
      And it carries no progress

    @objective-json-failure
    Scenario: a refusal is reported the same way it is asked for
      When I run `cairn objective new "" --json`
      Then it exits with code 1
      And it writes JSON to standard output
      And the JSON carries the reason

  Rule: Getting the command wrong is told, never guessed

    @objective-show-unknown
    Scenario: showing an objective that does not exist says so
      When I run `cairn objective show 99`
      Then it fails
      And it says there is no such objective

    @objective-usage-refused
    Scenario Outline: the command says what it needed
      When I run `<command>`
      Then it fails
      And it does not crash
      And it says how to use it

      Examples: what someone types the first time
        | command                    | what is wrong             |
        | cairn objective new        | no title at all           |
        | cairn objective show       | no number given           |
        | cairn objective            | the group with no command |
        | cairn objective addd "x"   | a typo in the subcommand  |
