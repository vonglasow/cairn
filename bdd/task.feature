Feature: Tasks
  A task is the unit of work. It carries a title, and an identity to refer to
  it by afterwards. Everything else — a description, a priority, a column,
  labels, a deadline, an estimate — is optional and has a stated default.

  A task sits in one of four columns — backlog, todo, current, done — and
  carries one of four priorities — low, normal, high, critical. Nothing else is
  a column, nothing else is a priority. A new task starts in the backlog at
  normal priority.

  Labels are a set. They are kept in lower case, so Spec and spec are one and
  the same. A task never carries the same one twice, and they always come back
  in alphabetical order, whatever order they were given in.

  A task outlives the command that created it: what one invocation writes, the
  next one finds, unchanged.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A task needs a title

    @task-title-empty
    Scenario: an empty title is no title
      When I run `cairn task add ""`
      Then it fails
      And it says a task needs a title
      And it shows the title I typed

    @task-title-spaces
    Scenario: a title of spaces is no title
      When I run `cairn task add "   "`
      Then it fails
      And it says a task needs a title
      And it shows the title I typed

    @task-title-trimmed
    Scenario: what surrounds a title is not part of it
      When I run `cairn task add "  Write the spec  "`
      Then it reports a task titled "Write the spec"

    @task-title-inner-spaces
    Scenario: spaces inside a title are kept
      When I run `cairn task add "Write  the   spec"`
      Then it reports a task titled "Write  the   spec"

    @task-title-content
    Scenario Outline: a title is taken as typed
      When I run `cairn task add "<title>"`
      Then it reports a task titled "<title>"

      Examples: nothing about the text is special
        | title                    | what it checks  |
        | Écrire la spéc — enfin ! | accents, dashes |
        | 90% done, right?         | punctuation     |
        | 🧗 climb                 | an emoji        |

    @task-title-dash
    Scenario: a title starting with a dash is still a title
      When I run `cairn task add -- "--not-a-flag"`
      Then it reports a task titled "--not-a-flag"

    @task-title-long
    Scenario: a long title is not truncated
      Given a title of 5000 characters
      When I add a task with it
      Then it reports a task with that exact title

  Rule: A task is numbered so it can be referred to later

    @task-number-first
    Scenario: the first task is numbered 1
      When I run `cairn task add "Write the spec"`
      Then it numbers the task 1

    @task-number-next
    Scenario: the next task gets the next number
      Given I added a task titled "First"
      When I run `cairn task add "Second"`
      Then it numbers the task 2

    @task-number-same-title
    Scenario: two tasks with the same title are still two tasks
      Given I added a task titled "Write the spec"
      When I run `cairn task add "Write the spec"`
      Then it numbers the task 2

  Rule: A task carries only what it was given

    @task-defaults
    Scenario: a new task carries nothing else
      When I run `cairn task add "Write the spec"`
      Then the task has no description
      And the task has no labels
      And the task has no deadline
      And the task has no estimate
      And the task has no time spent
      And the task carries the priority "normal"
      And the task is in the "backlog" column

    @task-columns
    Scenario Outline: every column a task can sit in
      When I run `cairn task add "Write the spec" --state <column>`
      Then the task is in the "<column>" column

      Examples: the four, and no others
        | column  |
        | backlog |
        | todo    |
        | current |
        | done    |

    @task-priorities
    Scenario Outline: every priority a task can carry
      When I run `cairn task add "Write the spec" --priority <priority>`
      Then the task carries the priority "<priority>"

      Examples: the four, and no others
        | priority |
        | low      |
        | normal   |
        | high     |
        | critical |

    @task-attribute-accepted
    Scenario Outline: an attribute given is an attribute carried
      When I run `cairn task add "Write the spec" <flag> <value>`
      Then the task carries the <attribute> "<stored>"

      Examples: each attribute, and how it comes back
        | flag       | value      | attribute   | stored     |
        | --desc     | the one    | description | the one    |
        | --label    | spec       | label       | spec       |
        | --deadline | 2026-09-01 | deadline    | 2026-09-01 |
        | --estimate | 90m        | estimate    | 1h30m      |

    @task-attribute-refused
    Scenario Outline: an attribute it cannot make sense of is refused
      When I run `cairn task add "Write the spec" <flag> <value>`
      Then it fails
      And it says what it expects instead

      Examples: nothing is guessed
        | flag       | value      | what is wrong             |
        | --priority | urgent     | not one of the priorities |
        | --state    | someday    | not one of the columns    |
        | --deadline | soon       | not a date                |
        | --deadline | 2026-13-45 | a date that cannot exist  |
        | --estimate | ages       | not a duration            |
        | --estimate | -30m       | a negative duration       |

    @task-label-blank
    Scenario: a label with no name is refused
      When I run `cairn task add "Write the spec" --label "  "`
      Then it fails
      And it says a label needs a name

    @task-labels-several
    Scenario: a task can carry several labels
      When I run `cairn task add "Write the spec" --label spec --label team`
      Then the task carries the labels "spec" and "team"

    @task-labels-sorted
    Scenario: labels come back in alphabetical order, not the order they were given
      When I run `cairn task add "Write the spec" --label team --label spec`
      Then the task carries the labels "spec" and "team"

    @task-labels-lowercased
    Scenario: a label is kept in lower case, whatever case it was given in
      When I run `cairn task add "Write the spec" --label Spec`
      Then the task carries the label "spec"

    @task-labels-repeated
    Scenario: the same label twice counts once
      When I run `cairn task add "Write the spec" --label spec --label spec`
      Then the task carries exactly 1 label

    @task-labels-same-in-any-case
    Scenario: the same label in two cases is still one label
      When I run `cairn task add "Write the spec" --label Spec --label spec`
      Then the task carries exactly 1 label
      And the task carries the label "spec"

    @task-attributes-together
    Scenario: attributes do not interfere with each other
      When I run `cairn task add "Write the spec" --priority high --state todo --label spec`
      Then the task carries the priority "high"
      And the task is in the "todo" column
      And the task carries the label "spec"

  Rule: Tasks can be listed

    @task-list-empty
    Scenario: listing when there is nothing says so
      When I run `cairn task list`
      Then it succeeds
      And it says there is no task

    @task-list
    Scenario: listing shows every task with its number
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task list`
      Then it lists 2 tasks
      And it lists task 1 titled "First"
      And it lists task 2 titled "Second"

    @task-list-order
    Scenario: tasks are listed in the order they were added
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task list`
      Then task 1 is listed before task 2

    @task-list-under
    Scenario: a sub-task is listed under the task it hangs from
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task list`
      Then it lists 2 tasks
      And task 2 is listed under task 1

    @task-list-under-several
    Scenario: everything hanging under a task is listed under it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I added a task titled "Write the docs" with the parent "1"
      When I run `cairn task list`
      Then it lists 3 tasks
      And task 2 is listed under task 1
      And task 3 is listed under task 1

    @task-list-under-then-next
    Scenario: a task hanging under nothing follows the whole of the one before it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I added a task titled "Something else"
      When I run `cairn task list`
      Then task 2 is listed before task 3
      And task 2 is listed under task 1
      And task 3 hangs under nothing in the listing

  Rule: A task can be read back by its number

    @task-show
    Scenario: showing a task gives back what it carries
      Given I added a task titled "Write the spec" with the description "the one"
      When I run `cairn task show 1`
      Then it reports a task titled "Write the spec"
      And the task carries the description "the one"

    @task-show-labels
    Scenario: showing gives back every label the task carries
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task show 1`
      Then the task carries the labels "spec" and "team"
      And the task carries exactly 2 labels

    @task-show-state
    Scenario: showing gives back the column the task is in
      Given I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task show 1`
      Then the task is in the "todo" column

    @task-show-unknown
    Scenario: showing a task that does not exist says so
      When I run `cairn task show 99`
      Then it fails
      And it says there is no such task

    @task-show-not-a-number
    Scenario: showing something that is not a number says so
      When I run `cairn task show spec`
      Then it fails
      And it says a task number is expected

    @task-show-after-edit
    Scenario: showing gives back the changes, not the original
      Given I added a task titled "Write the spec"
      And I edited task 1 with the title "Write the other spec"
      And cairn has exited
      When I run `cairn task show 1`
      Then it reports a task titled "Write the other spec"

  Rule: A task can be changed

    Everything a task can be given when it is created can be changed
    afterwards, read back, and taken away again. Labels are the odd one out:
    a task carries a set of them, so what is given replaces what was there —
    a label left out of the new set is a label removed.

    @task-edit-title
    Scenario: changing the title
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --title "Write the other spec"`
      Then it reports a task titled "Write the other spec"

    @task-edit-attribute
    Scenario Outline: changing an attribute
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 <flag> <value>`
      Then the task carries the <attribute> "<stored>"

      Examples: each attribute can be changed after the fact
        | flag       | value      | attribute   | stored     |
        | --desc     | the one    | description | the one    |
        | --priority | high       | priority    | high       |
        | --state    | todo       | column      | todo       |
        | --label    | spec       | label       | spec       |
        | --deadline | 2026-09-01 | deadline    | 2026-09-01 |
        | --estimate | 90m        | estimate    | 1h30m      |

    @task-edit-attribute-refused
    Scenario Outline: what was refused at creation is refused at edit
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 <flag> <value>`
      Then it fails
      And it says what it expects instead

      Examples: the rules do not soften after the fact
        | flag       | value   | what is wrong             |
        | --priority | urgent  | not one of the priorities |
        | --state    | someday | not one of the columns    |
        | --deadline | soon    | not a date                |
        | --estimate | ages    | not a duration            |

    @task-edit-label-blank
    Scenario: a label with no name is refused when editing too
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --label "  "`
      Then it fails
      And it says a label needs a name

    @task-edit-labels-replace
    Scenario: the labels given replace the ones that were there
      Given I added a task titled "Write the spec" with the labels "spec" and "draft"
      When I run `cairn task edit 1 --label spec --label team`
      Then the task carries the labels "spec" and "team"
      And the task carries exactly 2 labels

    @task-edit-label-dropped
    Scenario: a label left out is a label removed
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label spec`
      Then the task carries exactly 1 label
      And the task carries the label "spec"

    @task-edit-clear
    Scenario: an attribute given empty is cleared
      Given I added a task titled "Write the spec" with the deadline "2026-09-01"
      When I run `cairn task edit 1 --deadline ""`
      Then the task has no deadline

    @task-edit-labels-cleared
    Scenario: labels given empty are all taken away
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label ""`
      Then the task has no labels

    @task-edit-state-cleared
    Scenario: a column given empty returns the task to the backlog
      Given I added a task titled "Write the spec" with the state "current"
      When I run `cairn task edit 1 --state ""`
      Then the task is in the "backlog" column

    @task-edit-priority-cleared
    Scenario: a priority given empty returns the task to normal
      Given I added a task titled "Write the spec" with the priority "critical"
      When I run `cairn task edit 1 --priority ""`
      Then the task carries the priority "normal"

    @task-edit-untouched
    Scenario: what was not named is left alone
      Given I added a task titled "Write the spec" with the description "the one"
      When I run `cairn task edit 1 --priority high`
      Then the task carries the description "the one"
      And it reports a task titled "Write the spec"

    @task-edit-blank-title
    Scenario: a title cannot be emptied by editing
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --title ""`
      Then it fails
      And it says a task needs a title

    @task-edit-unknown
    Scenario: editing a task that does not exist says so
      When I run `cairn task edit 99 --title "Nope"`
      Then it fails
      And it says there is no such task

    @task-edit-nothing
    Scenario: editing without naming anything to change says so
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1`
      Then it fails
      And it says what it can change

    @task-edit-keeps-number
    Scenario: changing a task does not renumber it
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task edit 1 --title "First, revised"`
      Then it numbers the task 1
      And `cairn task list` lists 2 tasks

    @task-edit-twice
    Scenario: the last change wins
      Given I added a task titled "Write the spec"
      And I edited task 1 with the title "Second thoughts"
      When I run `cairn task edit 1 --title "Third thoughts"`
      Then it reports a task titled "Third thoughts"

  Rule: A label can be added or taken away without naming the others

    A label given plainly says what the set becomes, so adding one means typing
    them all again. A label given with a leading + is added to what is there,
    one with a leading - is taken away, and the rest is left alone. The two
    cannot be mixed: one says what the set becomes, the other says how it
    changes, and a command that did both would have to guess which won.

    @task-label-add
    Scenario: a label given with a plus joins the others
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label +urgent`
      Then the task carries exactly 3 labels
      And the task carries the label "urgent"

    @task-label-remove
    Scenario: a label given with a minus leaves the others alone
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label -spec`
      Then the task carries exactly 1 label
      And the task carries the label "team"

    @task-label-add-and-remove
    Scenario: adding and taking away in the same breath
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label +urgent --label -spec`
      Then the task carries the labels "team" and "urgent"
      And the task carries exactly 2 labels

    @task-label-add-known
    Scenario: adding one the task already carries changes nothing
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label +spec`
      Then the task carries exactly 2 labels

    @task-label-remove-unknown
    Scenario: taking away one the task does not carry changes nothing
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label -draft`
      Then the task carries exactly 2 labels

    @task-label-add-case
    Scenario: a label joining keeps the same rule about case
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label +Urgent`
      Then the task carries the label "urgent"

    @task-label-mixed
    Scenario: saying what the set becomes and how it changes at once is refused
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label +urgent --label draft`
      Then it fails
      And it says the labels are either replaced or changed, not both

    @task-label-sign-alone
    Scenario: a sign with no name is no label
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --label +`
      Then it fails
      And it says a label needs a name

    @task-label-sign-at-creation
    Scenario: a new task has nothing to add to
      When I run `cairn task add "Write the spec" --label +urgent`
      Then it fails
      And it says a label is added or taken away only when changing a task

  Rule: A task can be deleted

    @task-delete
    Scenario: deleting a task removes it
      Given I added a task titled "Write the spec"
      When I run `cairn task delete 1 --yes`
      Then it succeeds
      And it says task 1 was deleted
      And `cairn task list` says there is no task

    @task-delete-unconfirmed
    Scenario: deleting without confirming refuses
      Given I added a task titled "Write the spec"
      When I run `cairn task delete 1`
      Then it fails
      And it says how to confirm
      And `cairn task show 1` still reports a task titled "Write the spec"

    @task-delete-unknown
    Scenario: deleting a task that does not exist says so
      When I run `cairn task delete 99 --yes`
      Then it fails
      And it says there is no such task

    @task-delete-twice
    Scenario: deleting the same task twice says so the second time
      Given I added a task titled "Write the spec"
      And I deleted task 1
      When I run `cairn task delete 1 --yes`
      Then it fails
      And it says there is no such task

    @task-delete-leaves-others
    Scenario: deleting one task leaves the others alone
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task delete 2 --yes`
      Then it succeeds
      And `cairn task list` lists 2 tasks
      And `cairn task list` lists task 1 titled "First"
      And `cairn task list` lists task 3 titled "Third"

    @task-number-not-reused
    Scenario: a number freed by a deletion is not handed out again
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I deleted task 2
      When I run `cairn task add "Third"`
      Then it numbers the task 3

  Rule: A task can hang under another one

    A task given a parent becomes one of its sub-tasks, and stays a task in
    every other way: its own number, its own priority, its own labels, changed
    and read back like any other. Only one level — a sub-task cannot itself
    carry sub-tasks, so what hangs under a task is never something with more
    hanging under it.

    A task carrying sub-tasks is not deleted while they are there. What depends
    on it would go with it, and a deletion that takes more than the number
    typed is a deletion nobody asked for.

    @task-parent-none
    Scenario: a task hangs under nothing unless it is told to
      When I run `cairn task add "Ship it"`
      Then the task has no parent

    @task-parent-add
    Scenario: a task can be added under another
      Given I added a task titled "Ship it"
      When I run `cairn task add "Write the tests" --parent 1`
      Then the task carries the parent "1"

    @task-parent-edit
    Scenario: a task can be hung under another after the fact
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests"
      When I run `cairn task edit 2 --parent 1`
      Then the task carries the parent "1"

    @task-parent-cleared
    Scenario: a parent given empty sets the task free
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task edit 2 --parent ""`
      Then the task has no parent

    @task-parent-unknown
    Scenario: hanging a task under one that does not exist is refused
      Given I added a task titled "Ship it"
      When I run `cairn task add "Write the tests" --parent 99`
      Then it fails
      And it says there is no such task

    @task-parent-itself
    Scenario: a task cannot hang under itself
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --parent 1`
      Then it fails
      And it says a task cannot hang under itself

    @task-parent-not-a-number
    Scenario: a parent that is not a number says so
      Given I added a task titled "Ship it"
      When I run `cairn task add "Write the tests" --parent spec`
      Then it fails
      And it says a task number is expected

    @task-parent-depth
    Scenario: a sub-task cannot carry sub-tasks
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task add "Write the first one" --parent 2`
      Then it fails
      And it says a sub-task cannot carry sub-tasks

    @task-parent-shown
    Scenario: showing a task names what hangs under it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task show 1`
      Then it lists task 2 titled "Write the tests"

    @task-parent-listed
    Scenario: a sub-task is listed in its own right
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task list`
      Then it lists 2 tasks

    @task-parent-delete-refused
    Scenario: a task carrying sub-tasks is not deleted
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task delete 1 --yes`
      Then it fails
      And it says the task carries sub-tasks
      And `cairn task show 1` still reports a task titled "Ship it"

    @task-parent-delete-freed
    Scenario: once nothing hangs under it, the task can be deleted
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I deleted task 2
      When I run `cairn task delete 1 --yes`
      Then it succeeds

    @task-parent-survives
    Scenario: the link outlives the command that made it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And cairn has exited
      When I run `cairn task show 2`
      Then the task carries the parent "1"

    @task-parent-json
    Scenario: the link can be read by a machine
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task show 2 --json`
      Then it writes JSON to standard output
      And it carries the parent 1

  Rule: A task outlives the command that created it

    @task-survives
    Scenario: a task added by one command is found by the next
      Given I added a task titled "Write the spec"
      And cairn has exited
      When I run `cairn task list`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-number-survives
    Scenario: numbering picks up where it left off
      Given I added a task titled "First"
      And cairn has exited
      When I run `cairn task add "Second"`
      Then it numbers the task 2

    @task-store-created
    Scenario: the first command creates the store
      When I run `cairn task add "Write the spec"`
      Then it succeeds
      And the store exists

    @task-store-elsewhere
    Scenario: told where to live, it does not touch the default place
      When I run `cairn task add "Write the spec"`
      Then the store is created where I said
      And nothing is written to the default place

    @task-store-unreadable
    Scenario: a store it cannot read is reported, never replaced
      Given the store holds something that is not a cairn store
      When I run `cairn task list`
      Then it fails
      And it says the store cannot be read
      And it names the store
      And the store is left as it was

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

    @task-roundtrip-long
    Scenario: a long title survives the store
      Given a title of 5000 characters
      And I added a task with it
      And cairn has exited
      When I run `cairn task show 1`
      Then it reports a task with that exact title

    @task-roundtrip-deadline
    Scenario: a deadline comes back as the day it was given
      Given I added a task titled "Write the spec" with the deadline "2026-09-01"
      And cairn has exited
      When I run `cairn task show 1`
      Then the task carries the deadline "2026-09-01"

    @task-roundtrip-labels
    Scenario: every label survives the store
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      And cairn has exited
      When I run `cairn task show 1`
      Then the task carries the labels "spec" and "team"
      And the task carries exactly 2 labels

    @task-roundtrip-state
    Scenario: the column survives the store
      Given I added a task titled "Write the spec" with the state "todo"
      And cairn has exited
      When I run `cairn task show 1`
      Then the task is in the "todo" column

  Rule: Getting the command wrong is told, never guessed

    @task-usage-refused
    Scenario Outline: the command says what it needed
      When I run `<command>`
      Then it fails
      And it does not crash
      And it says how to use it

      Examples: what someone types the first time
        | command                       | what is wrong             |
        | cairn task add                | no title at all           |
        | cairn task add Write the spec | quotes forgotten          |
        | cairn task add "a" "b"        | two titles                |
        | cairn task addd "Write"       | a typo in the subcommand  |
        | cairn task                    | the group with no command |
        | cairn task show               | no number given           |
        | cairn task edit               | no number given           |
        | cairn task delete             | no number given           |

    @task-usage-unknown-flag
    Scenario: an unknown flag is refused rather than ignored
      When I run `cairn task add "Write the spec" --colour red`
      Then it fails
      And it says the flag is unknown

  Rule: It answers the same way every time

    A result goes to standard output. A reason meant for a person goes to
    standard error. The exit code says which kind of answer it is: 0 it worked,
    1 the request was understood but could not be satisfied, 2 the command was
    not understood, 3 the store could not be used.

    @task-exit-success
    Scenario: what worked says nothing on the error stream
      When I run `cairn task add "Write the spec"`
      Then it exits with code 0
      And it writes the task to standard output
      And it writes nothing to standard error

    @task-exit-refused
    Scenario Outline: each kind of refusal has its own code
      When I run `<command>`
      Then it exits with code <code>
      And it writes the reason to standard error
      And it writes nothing to standard output

      Examples: understood but impossible, against not understood at all
        | command                               | code | what happened            |
        | cairn task add ""                     | 1    | no title to work with    |
        | cairn task show 99                    | 1    | nothing there to show    |
        | cairn task delete 99 --yes            | 1    | nothing there to delete  |
        | cairn task add                        | 2    | no title given at all    |
        | cairn task add "Spec" --colour red    | 2    | a flag it does not know  |
        | cairn task show spec                  | 2    | a number was expected    |

    @task-exit-store
    Scenario: a store it cannot use has a code of its own
      Given the store holds something that is not a cairn store
      When I run `cairn task list`
      Then it exits with code 3
      And it writes the reason to standard error

  Rule: The output can be read by a machine

    Asked for JSON, the whole answer goes to standard output as JSON — whether
    it worked or not. A machine reads one stream, and a refusal it cannot parse
    is a refusal it cannot act on.

    @task-json
    Scenario: asking for machine-readable output
      When I run `cairn task add "Write the spec" --json`
      Then it writes JSON to standard output
      And it carries the title "Write the spec"
      And it carries an identifier

    @task-json-list
    Scenario: listing can be read by a machine
      Given I added a task titled "First"
      When I run `cairn task list --json`
      Then it writes JSON to standard output
      And it carries 1 task

    @task-json-show
    Scenario: showing can be read by a machine
      Given I added a task titled "Write the spec"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries the title "Write the spec"

    @task-json-failure
    Scenario: a refusal is reported the same way it is asked for
      When I run `cairn task add "" --json`
      Then it exits with code 1
      And it writes JSON to standard output
      And the JSON carries the reason
