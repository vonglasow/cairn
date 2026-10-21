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

    @task-title-spaces
    Scenario: a title of spaces is no title
      When I run `cairn task new "   "`
      Then it fails
      And it says a task needs a title
      And it shows the title I typed

    @task-title-trimmed
    Scenario: what surrounds a title is not part of it
      When I run `cairn task new "  Write the spec  "`
      Then it reports a task titled "Write the spec"

    @task-title-inner-spaces
    Scenario: spaces inside a title are kept
      When I run `cairn task new "Write  the   spec"`
      Then it reports a task titled "Write  the   spec"

    @task-title-content
    Scenario Outline: a title is taken as typed
      When I run `cairn task new "<title>"`
      Then it reports a task titled "<title>"

      Examples: nothing about the text is special
        | title                    | what it checks  |
        | Écrire la spéc — enfin ! | accents, dashes |
        | 90% done, right?         | punctuation     |
        | 🧗 climb                 | an emoji        |

    @task-title-dash
    Scenario: a title starting with a dash is still a title
      When I run `cairn task new -- "--not-a-flag"`
      Then it reports a task titled "--not-a-flag"

    @task-title-one-line
    Scenario Outline: a title is one line
      Given a title with a <character> in it
      When I add a task with it
      Then it fails
      And it says a title is one line

      Examples: what a card and a listing cannot show
        | character  |
        | line break |
        | tab        |

    @task-title-long
    Scenario: a long title is not truncated
      Given a title of 5000 characters
      When I add a task with it
      Then it reports a task with that exact title

  Rule: A task is numbered so it can be referred to later

    @task-number-first
    Scenario: the first task is numbered 1
      When I run `cairn task new "Write the spec"`
      Then it numbers the task 1

    @task-number-next
    Scenario: the next task gets the next number
      Given I added a task titled "First"
      When I run `cairn task new "Second"`
      Then it numbers the task 2

    @task-number-same-title
    Scenario: two tasks with the same title are still two tasks
      Given I added a task titled "Write the spec"
      When I run `cairn task new "Write the spec"`
      Then it numbers the task 2

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

    @task-columns
    Scenario Outline: every column a task can sit in
      When I run `cairn task new "Write the spec" --state <column>`
      Then the task is in the "<column>" column

      Examples: the four, and no others
        | column  |
        | backlog |
        | todo    |
        | current |
        | done    |

    @task-priorities
    Scenario Outline: every priority a task can carry
      When I run `cairn task new "Write the spec" --priority <priority>`
      Then the task carries the priority "<priority>"

      Examples: the four, and no others
        | priority |
        | low      |
        | normal   |
        | high     |
        | critical |

    @task-attribute-accepted
    Scenario Outline: an attribute given is an attribute carried
      When I run `cairn task new "Write the spec" <flag> <value>`
      Then the task carries the <attribute> "<stored>"

      Examples: each attribute, and how it comes back
        | flag       | value      | attribute   | stored     |
        | --desc     | the one    | description | the one    |
        | --label    | spec       | label       | spec       |
        | --deadline | 2026-09-01 | deadline    | 2026-09-01 |
        | --estimate | 90m        | estimate    | 1h30m      |
        | --estimate | 30s        | estimate    | 30s        |

    @task-attribute-refused
    Scenario Outline: an attribute it cannot make sense of is refused
      When I run `cairn task new "Write the spec" <flag> <value>`
      Then it fails
      And it says what it expects instead of "<value>"

      Examples: nothing is guessed
        | flag       | value      | what is wrong             |
        | --priority | urgent     | not one of the priorities |
        | --state    | someday    | not one of the columns    |
        | --deadline | soon       | not a date                |
        | --deadline | 2026-13-45 | a date that cannot exist  |
        | --estimate | ages       | not a duration            |
        | --estimate | -30m       | a negative duration       |
        | --estimate | 500ms      | shorter than a second     |
        | --spent    | 500ms      | shorter than a second     |

    @task-label-blank
    Scenario: a label with no name is refused
      When I run `cairn task new "Write the spec" --label "  "`
      Then it fails
      And it says a label needs a name

    @task-labels-several
    Scenario: a task can carry several labels
      When I run `cairn task new "Write the spec" --label spec --label team`
      Then the task carries the labels "spec" and "team"

    @task-labels-sorted
    Scenario: labels come back in alphabetical order, not the order they were given
      When I run `cairn task new "Write the spec" --label team --label spec`
      Then the task carries the labels "spec" and "team"

    @task-labels-lowercased
    Scenario: a label is kept in lower case, whatever case it was given in
      When I run `cairn task new "Write the spec" --label Spec`
      Then the task carries the label "spec"

    @task-labels-repeated
    Scenario: the same label twice counts once
      When I run `cairn task new "Write the spec" --label spec --label spec`
      Then the task carries exactly 1 label

    @task-labels-same-in-any-case
    Scenario: the same label in two cases is still one label
      When I run `cairn task new "Write the spec" --label Spec --label spec`
      Then the task carries exactly 1 label
      And the task carries the label "spec"

    @task-labels-written-one-way
    Scenario Outline: a label is written one way, whatever way it was typed
      When I run `cairn task new "Write the spec" --label "<typed>"`
      Then the task carries exactly 1 label
      And the task carries the label "<kept>"

      Examples:
        | typed             | kept            |
        | Sp26 20           | sp26-20         |
        | Modularité/Helm   | modularite-helm |
        | dette__config     | dette-config    |
        | Release__26.1.6-- | release-26.1.6  |

    @task-labels-no-family
    Scenario: a label with a family is refused, and says to write two
      When I run `cairn task new "Write the spec" --label "ado:77594"`
      Then it fails
      And it says a label has no family

    @task-labels-what-a-label-holds
    Scenario: a label holding what no label may hold is refused, saying what one may hold
      When I run `cairn task new "Write the spec" --label "#urgent"`
      Then it fails
      And it says what a label may hold

    @task-attributes-together
    Scenario: attributes do not interfere with each other
      When I run `cairn task new "Write the spec" --priority high --state todo --label spec`
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

    @task-list-carries-what-a-decision-is-made-on
    Scenario: a listed task carries what a decision is made on
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --state current --priority high --deadline 2026-10-09 --label backend`
      And I run `cairn task list`
      Then it succeeds
      And the line for task 1 says "Write the spec"
      And the line for task 1 says "current"
      And the line for task 1 says "high"
      And the line for task 1 says "2026-10-09"
      And the line for task 1 says "backend"

    @task-list-order
    Scenario: tasks are listed in the order they were added
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task list`
      Then task 1 is listed before task 2

    @task-list-leaves-steps-out
    Scenario: a listing names the tasks, never their steps
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      When I run `cairn task list`
      Then it lists 1 task
      And it lists task 1 titled "Ship it"
      And it does not mention "Write the tests"

    @task-show-names-its-steps
    Scenario: a task's steps are named where the task is shown, done or not
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      And I added the step "Write the docs" to task 1
      And I ticked step 1 of task 1
      When I run `cairn task show 1`
      Then it names the step "Write the tests" as done
      And it names the step "Write the docs" as not done

    @task-show-json-carries-its-steps
    Scenario: a machine asking about a task is given its steps
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      And I ticked step 1 of task 1
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries the step "Write the tests", done

  Rule: A listing can be narrowed to what is wanted

    @task-list-narrowed-by-state
    Scenario: only the tasks in one column
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      When I run `cairn task list --state todo`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-a-deadline
    Scenario: only the tasks due on or before a day
      Given I added a task titled "Write the spec" with the deadline "2026-09-10"
      And I added a task titled "Ship it" with the deadline "2026-12-01"
      And I added a task titled "Someday"
      When I run `cairn task list --due-before 2026-10-01`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-a-deadline-takes-the-day-itself
    Scenario: the day named is inside the range, not beside it
      Given I added a task titled "Write the spec" with the deadline "2026-10-01"
      When I run `cairn task list --due-before 2026-10-01`
      Then it lists 1 task

    @task-list-narrowed-by-a-deadline-refuses-what-is-not-a-day
    Scenario: a day that is not one is refused rather than read as none
      Given I added a task titled "Write the spec" with the deadline "2026-10-01"
      When I run `cairn task list --due-before soon`
      Then it fails

    @task-list-narrowed-by-priority
    Scenario: only the tasks carrying one priority
      Given I added a task titled "Write the spec" with the priority "high"
      And I added a task titled "Ship it"
      When I run `cairn task list --priority high`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-a-label
    Scenario: only the tasks carrying one label
      Given I added a task titled "Write the spec" with the labels "team" and "urgent"
      And I added a task titled "Ship it" with the labels "team" and "later"
      When I run `cairn task list --label urgent`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-a-label-however-typed
    Scenario: a label is looked for written the one way, however it was typed
      Given I added a task titled "Write the spec" with the labels "sp26-20" and "team"
      And I added a task titled "Ship it" with the labels "team" and "later"
      When I run `cairn task list --label "SP26_20"`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-a-column-that-is-not-one
    Scenario: narrowing by something no task can be in is refused
      Given I added a task titled "Write the spec"
      When I run `cairn task list --state nowhere`
      Then it fails
      And it says what it expects instead of "nowhere"

    @task-list-narrowed-by-an-objective-nobody-added
    Scenario: narrowing by an objective that is not there is answered, not refused
      Given I added a task titled "Write the spec"
      When I run `cairn task list --objective 9`
      Then it succeeds
      And it says there is no task

    @task-list-narrowed-still-says-the-focus
    Scenario: a task in focus that matched is shown as in focus
      Given I added a task titled "Write the spec" with the state "current"
      And I added a task titled "Ship it"
      And I put task 1 in focus
      When I run `cairn task list --state current --json`
      Then it carries 1 task
      And it says task 1 is the one in focus

    @task-list-narrowed-past-the-focus-says-none
    Scenario: a focus narrowed out of the answer is not spoken for
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      And I put task 2 in focus
      When I run `cairn task list --state todo --json`
      Then it carries 1 task
      And it says no task in it is in focus

    @task-list-narrowed-by-two-things
    Scenario: everything named has to hold
      Given I added a task titled "Write the spec" with the state "todo"
      And I edited task 1 with the priority "high"
      And I added a task titled "Ship it" with the state "todo"
      When I run `cairn task list --state todo --priority high`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-by-two-labels
    Scenario: a task named twice over carries both labels
      Given I added a task titled "Write the spec" with the labels "team" and "urgent"
      And I added a task titled "Ship it" with the labels "team" and "later"
      When I run `cairn task list --label team --label urgent`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-list-narrowed-to-nothing
    Scenario: nothing carrying it is said, not answered as everything
      Given I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task list --state done`
      Then it succeeds
      And it says there is no task

    @task-list-narrowed-by-the-objective-served
    Scenario: only the tasks serving one objective
      Given I added an objective titled "Ship cairn 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Ship it"
      When I run `cairn task list --objective 1`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

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

    @task-description-from-standard-input
    Scenario: a description too long to type is read from standard input

      When I run `cairn task new "Write the spec" --desc -` with "first paragraph" and "second paragraph" on standard input
      Then it succeeds
      And `cairn task show 1` shows "second paragraph"

    @task-description-not-text
    Scenario: a description a terminal would obey is refused

      When I add a task whose description carries an escape
      Then it fails
      And it says only text, line breaks and tabs can be written there

    @task-show-since
    Scenario: showing a task says since when it has stood in its column

      Given I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task show 1`
      Then it says the task has stood in "todo" since today

    @task-show-since-json
    Scenario: a machine is told since when a task has stood in its column
      Given I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task show 1 --json`
      Then it says in JSON the task has stood in its column since today

    @task-number-negative
    Scenario: a number that starts with a minus is still a number
      When I run `cairn task show -1`
      Then it fails
      And it says a task number is expected

    @task-show-unknown
    Scenario: showing a task that does not exist says so
      When I run `cairn task show 99`
      Then it fails
      And it says there is no such task

    @task-show-unknown-as-json
    Scenario: a refusal asked for as JSON says what kind it is, and which

      When I run `cairn task show 99 --json`
      Then it fails
      And it says in JSON the refusal is "refused", because "no-such-task"

    @task-refusal-quotes-as-typed
    Scenario: a refusal quotes what was typed exactly as it was typed

      When I run `cairn task show abc`
      Then it fails
      And it quotes "abc" back as it was typed

    @task-list-refuses-a-column-like-an-edit
    Scenario: a column that does not exist is refused for the same reason everywhere
      When I run `cairn task list --state bogus --json`
      Then it fails
      And it says in JSON the refusal is "refused", because "not-a-state"

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

  Rule: A task is looked for by what somebody remembers of it

    @task-search-by-a-word
    Scenario: a word from the title is the way back to a number
      Given I added a task titled "Write the spec"
      When I run `cairn task search spec`
      Then it lists task 1 titled "Write the spec"

    @task-search-by-its-number
    Scenario: the number is the way back when it is what somebody has
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I run `cairn task search 2`
      Then it lists task 2 titled "Ship it"

    @task-search-narrowed
    Scenario: a search can be narrowed by what a listing narrows by
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Rewrite the spec" with the state "done"
      When I run `cairn task search spec --state todo`
      Then it succeeds
      And it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-search-narrowed-reaching-nothing
    Scenario: a search narrowed to where the word is not answers nothing
      Given I added a task titled "Write the spec" with the state "todo"
      When I run `cairn task search spec --state done`
      Then it succeeds
      And it says there is no task

    @task-search-leaves-archived-out-until-asked-for
    Scenario: what was archived is not looked in until it is asked for
      Given I added a task titled "Write the spec"
      And I archived task 1
      When I run `cairn task search spec`
      Then it says there is no task
      When I run `cairn task search spec --archived`
      Then it lists 1 task

    @task-search-ignores-case
    Scenario: looking does not care how it was typed
      Given I added a task titled "Write the spec"
      When I run `cairn task search SPEC`
      Then it lists 1 task

    @task-search-reaches-every-field
    Scenario Outline: anything a task carries is looked in
      Given I added a task titled "Write the spec" with the <field> "<carried>"
      When I run `cairn task search <named>`
      Then it lists 1 task

      Examples: read the way a person reads them
        | field       | carried    | named   |
        | description | the first  | first   |
        | deadline    | 2026-09-01 | 2026-09 |
        | time spent  | 90m        | 1h30m   |

    @task-search-reaches-the-words-that-block
    Scenario: what somebody typed is looked in, blockers included
      Given I added a task titled "Ship it"
      And I edited task 1 with the because "the client's answer"
      When I run `cairn task search client`
      Then it lists 1 task

    @task-search-reaches-the-column
    Scenario: the column is looked in like anything else it carries
      Given I added a task titled "Write the spec" with the state "current"
      And I added a task titled "Ship it"
      When I run `cairn task search current`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-search-reaches-the-priority
    Scenario: the priority is looked in like anything else it carries
      Given I added a task titled "Write the spec" with the priority "high"
      And I added a task titled "Ship it"
      When I run `cairn task search high`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-search-reaches-a-label
    Scenario: a label is looked in like anything else it carries
      Given I added a task titled "Write the spec" with the labels "team" and "urgent"
      And I added a task titled "Ship it"
      When I run `cairn task search urgent`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-search-everything-named-is-found
    Scenario: everything named has to be there, wherever it is
      Given I added a task titled "Write the spec" with the description "the first"
      And I added a task titled "Read the spec"
      When I run `cairn task search first spec`
      Then it lists task 1 titled "Write the spec"
      And it lists 1 task

    @task-search-a-thing-with-a-space
    Scenario: a thing with a space in it is one thing
      Given I added a task titled "Write the spec"
      And I added a task titled "Spec what to write"
      When I run `cairn task search "the spec"`
      Then it lists 1 task

    @task-search-reaches-a-step
    Scenario: a word found only in a step answers the task it belongs to
      Given I added a task titled "Write the spec"
      And I added the step "Read it first" to task 1
      When I run `cairn task search first`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

    @task-search-finds-nothing
    Scenario: looking for what is not there says so
      Given I added a task titled "Write the spec"
      When I run `cairn task search nowhere`
      Then it succeeds
      And it says there is no task

    @task-search-needs-something-to-look-for
    Scenario: looking for nothing at all is refused
      When I run `cairn task search`
      Then it fails
      And it says how to use it

  Rule: A task can be changed

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
        | --estimate | 30s        | estimate    | 30s        |

    @task-edit-attribute-refused
    Scenario Outline: what was refused at creation is refused at edit
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 <flag> <value>`
      Then it fails
      And it says what it expects instead of "<value>"

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

    @task-label-list
    Scenario: several labels can be given in one flag
      When I run `cairn task new "Write the spec" --label "spec,team"`
      Then the task carries exactly 2 labels
      And the task carries the labels "spec" and "team"

    @task-label-list-signed
    Scenario: a list joins the others when it is signed
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label "+urgent,+draft"`
      Then the task carries exactly 4 labels

    @task-label-list-mixed
    Scenario: a list cannot mix the two forms either
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      When I run `cairn task edit 1 --label "+urgent,spec"`
      Then it fails
      And it says the labels are either replaced or changed, not both

    @task-label-list-gaps
    Scenario: an empty place in a list is not a label
      When I run `cairn task new "Write the spec" --label "spec,,team,"`
      Then the task carries exactly 2 labels

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
      When I run `cairn task new "Write the spec" --label +urgent`
      Then it fails
      And it says a label is added or taken away only when changing a task

  Rule: A task can say what blocks it

    @task-blocker-named
    Scenario: a task can say which task blocks it
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I run `cairn task edit 2 --state blocked --blocker 1`
      Then it succeeds
      And `cairn task show 2` says task 1 blocks it

    @task-blocker-in-words
    Scenario: what is not in cairn is named in words
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --state blocked --because "the client's answer"`
      Then it succeeds
      And `cairn task show 1` says "the client's answer" blocks it

    @task-blocker-several
    Scenario: more than one thing can be in the way
      Given I added a task titled "Write the spec"
      And I added a task titled "Write the tests"
      And I added a task titled "Ship it"
      When I run `cairn task edit 3 --state blocked --blocker 1 --blocker 2 --because "the client's answer"`
      Then `cairn task show 3` says task 1 blocks it
      And `cairn task show 3` says task 2 blocks it
      And `cairn task show 3` says "the client's answer" blocks it

    @task-blocker-itself
    Scenario: a task cannot be what is blocking it
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --blocker 1`
      Then it fails
      And it says a task cannot block itself

    @task-blocker-a-cycle
    Scenario: two tasks cannot each be waiting on the other
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      When I run `cairn task edit 1 --blocker 2`
      Then it fails
      And it says nothing would ever be able to start

    @task-blocker-a-longer-cycle
    Scenario: a cycle is refused however long the way round
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      And I edited task 2 with the blocker "1"
      And I edited task 3 with the blocker "2"
      When I run `cairn task edit 1 --blocker 3`
      Then it fails
      And it says nothing would ever be able to start

    @task-blocker-a-task-nobody-added
    Scenario: a blocker has to be a task that exists
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --blocker 9`
      Then it fails
      And it says there is no such task

    @task-blocker-on-a-task-that-is-not-blocked
    Scenario: a blocker is a note until the column says otherwise
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it" with the state "todo"
      When I run `cairn task edit 2 --blocker 1`
      Then it succeeds
      And the task is in the "todo" column
      And `cairn task show 2` says task 1 blocks it

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

    @task-blocked-can-go-anywhere-else
    Scenario Outline: work that is waiting can still be filed
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 2 --state <column>`
      Then it succeeds
      And the task is in the "<column>" column

      Examples: nothing here claims the work is yours
        | column   |
        | done     |
        | backlog  |
        | paused   |
        | watching |

    @task-blocked-with-nothing-recorded-is-free
    Scenario: a task blocked by nothing named has nothing to take away
      Given I added a task titled "Ship it" with the state "blocked"
      When I run `cairn task edit 1 --state current`
      Then it succeeds
      And the task is in the "current" column

    @task-unblocked-comes-back-to-todo
    Scenario: taking the last thing away puts it back in todo
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 2 --blocker -1`
      Then it succeeds
      And the task is in the "todo" column

    @task-still-blocked-stays-blocked
    Scenario: taking one of two away leaves it where it is
      Given I added a task titled "Write the spec"
      And I added a task titled "Write the tests"
      And I added a task titled "Ship it"
      And I edited task 3 with the blocker "1"
      And I edited task 3 with the state "blocked"
      When I run `cairn task edit 3 --blocker +2 --blocker -1`
      Then it succeeds
      And the task is in the "blocked" column

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

    @task-blocked-stays-where-it-is
    Scenario: the mark is a reminder, not a move
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 1 --state done`
      And I run `cairn task show 2`
      Then the task is in the "blocked" column

    @task-blocked-one-still-open
    Scenario: one of two still open is still in the way
      Given I added a task titled "Write the spec"
      And I added a task titled "Write the tests"
      And I added a task titled "Ship it"
      And I edited task 3 with the blocker "1"
      And I edited task 3 with the blocker "+2"
      And I edited task 3 with the state "blocked"
      When I run `cairn task edit 1 --state done`
      And I run `cairn task list`
      Then it marks nothing as having nothing in its way

    @task-blocked-by-words-is-never-cleared
    Scenario Outline: words cairn cannot check keep it in the way
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the because "the client's answer"
      And I edited task 2 with the blocker "<also>"
      And I edited task 2 with the state "blocked"
      When I run `cairn task edit 1 --state done`
      And I run `cairn task list`
      Then it marks nothing as having nothing in its way

      Examples: cairn has no way to know whether the client answered
        | also |
        | 1    |
        |      |

    @task-blocked-by-nothing-named
    Scenario: nothing named is not nothing in the way
      Given I added a task titled "Ship it" with the state "blocked"
      When I run `cairn task list`
      Then it marks nothing as having nothing in its way

    @task-blocker-read-by-a-machine
    Scenario: what blocks a task crosses to a machine
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the because "the client's answer"
      When I run `cairn task show 2 --json`
      Then it writes JSON to standard output
      And it says task 1 blocks it and "the client's answer" does too

    @task-blocker-goes-with-the-task-deleted
    Scenario: deleting what blocked it takes the link away
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      When I run `cairn task delete 1 --yes`
      And I run `cairn task show 2`
      Then it says nothing blocks it

    @task-delete-names-what-it-was-blocking
    Scenario: deleting a task says which tasks it was in the way of

      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      When I run `cairn task delete 1 --yes`
      Then it succeeds
      And it says task 2 was waiting on it

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

    @task-delete-unconfirmed-names-it
    Scenario: a deletion refused for want of confirming names what it would take
      Given I added a task titled "Write the spec"
      When I run `cairn task delete 1`
      Then it fails
      And it names task 1

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
      When I run `cairn task new "Third"`
      Then it numbers the task 3

    @task-delete-several
    Scenario: several tasks are deleted in one act
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I run `cairn task delete 1 --also 3 --yes`
      Then it succeeds
      And `cairn task list` lists 1 tasks
      And `cairn task list` lists task 2 titled "Second"

    @task-delete-several-or-none
    Scenario: several tasks are deleted all together, or none of them is
      Given I added a task titled "First"
      When I run `cairn task delete 1 --also 99 --yes`
      Then it fails
      And it says there is no such task
      And `cairn task show 1` still reports a task titled "First"

    @task-delete-several-unconfirmed-names-them
    Scenario: a deletion of several refused for want of confirming names every one
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task delete 1 --also 2`
      Then it fails
      And it names task 1
      And it names task 2

  Rule: Time spent on a task is the time it was in focus and the time added by hand

    @task-spent-counts-focus
    Scenario: the time a task was in focus is time spent on it
      Given a store whose task 1 was in focus from 07:00 to 08:30 UTC on 2026-09-29
      When I run `cairn task show 1`
      Then the task carries the time spent "1h30m"

    @task-spent-by-hand-completes-focus
    Scenario: time added by hand completes the time in focus
      Given a store whose task 1 was in focus from 07:00 to 08:30 UTC on 2026-09-29
      When I run `cairn task edit 1 --spent +30m`
      Then the task carries the time spent "2h"

    @task-spent-written-is-the-whole
    Scenario: a time given plainly is the whole, the time in focus included
      Given a store whose task 1 was in focus from 07:00 to 08:30 UTC on 2026-09-29
      When I run `cairn task edit 1 --spent 2h`
      Then the task carries the time spent "2h"

    @task-spent-by-hand-is-on-the-timesheet
    Scenario: time added by hand to a task is on today's timesheet
      Given I added a task titled "Write the spec"
      And I ran `cairn task edit 1 --spent +30m`
      When I run `cairn timesheet show`
      Then today gives task 1 "30m"

    @task-spent-from-the-timesheet
    Scenario: time added on the timesheet is time spent on the task
      Given I added a task titled "Sprint review"
      And I ran `cairn timesheet add 1 1h --on 2026-09-29`
      When I run `cairn task show 1`
      Then the task carries the time spent "1h"

    @task-spent-written
    Scenario: time spent can be written down
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --spent 2h`
      Then the task carries the time spent "2h"

    @task-spent-added
    Scenario: time given with a plus is added to what is there
      Given I added a task titled "Write the spec" with the time spent "2h"
      When I run `cairn task edit 1 --spent +30m`
      Then the task carries the time spent "2h30m"

    @task-spent-taken-off
    Scenario: time given with a minus is taken off what is there
      Given I added a task titled "Write the spec" with the time spent "2h"
      When I run `cairn task edit 1 --spent -15m`
      Then the task carries the time spent "1h45m"

    @task-spent-taken-below-zero
    Scenario: taking off more than was there leaves nothing, not less
      Given I added a task titled "Write the spec" with the time spent "2h"
      When I run `cairn task edit 1 --spent -3h`
      Then the task has no time spent

    @task-spent-cleared
    Scenario: time spent given empty is cleared
      Given I added a task titled "Write the spec" with the time spent "2h"
      When I run `cairn task edit 1 --spent ""`
      Then the task has no time spent

    @task-spent-refused
    Scenario Outline: a time it cannot make sense of is refused
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --spent <value>`
      Then it fails
      And it says what it expects instead of "<value>"

      Examples: nothing is guessed
        | value | what is wrong           |
        | ages  | not a duration          |
        | +ages | not a duration to add   |
        | -ages | not a duration to take  |
        | +     | a sign with no duration |

  Rule: A label is put on everything a narrowing reaches

    @task-label-what-a-narrowing-reaches
    Scenario: a label is put on everything a narrowing reaches
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I added a task titled "Dépôt APP" with the state "done"
      And I added a task titled "Spike Argo"
      When I run `cairn task label sprint-41 --state done`
      Then it succeeds
      And it says 2 tasks were labelled
      And task 1 carries the label "sprint-41"
      And task 3 does not carry the label "sprint-41"

    @task-label-keeps-what-is-already-there
    Scenario: a label is added to the ones a task already carries
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I edited task 1 with the label "backend"
      When I run `cairn task label sprint-41 --state done`
      Then it succeeds
      And task 1 carries the label "backend"
      And task 1 carries the label "sprint-41"

    @task-label-keeps-by-the-labels-named
    Scenario: labelling reaches only what carries the label it was narrowed by
      Given I added a task titled "Migration Jenkins"
      And I edited task 1 with the label "backend"
      And I added a task titled "Spike Argo"
      When I run `cairn task label sprint-41 --label backend`
      Then it succeeds
      And it says 1 task was labelled
      And task 1 carries the label "sprint-41"
      And task 2 does not carry the label "sprint-41"

    @task-label-needs-something-to-go-on
    Scenario: labelling with nothing to narrow by is refused
      Given I added a task titled "Spike Argo"
      When I run `cairn task label sprint-41`
      Then it fails
      And it says what it needs to go on

  Rule: A task is archived rather than deleted

    @task-list-leaves-archived-out
    Scenario: a listing is the work still on the board
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      And I archived task 1
      When I run `cairn task list`
      Then it lists 1 task
      And it lists task 2 titled "Spike Argo"
      And it says nothing of task 1

    @task-list-archived-on-their-own
    Scenario: what was archived is asked for by name
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      And I archived task 1
      When I run `cairn task list --archived`
      Then it lists 1 task
      And it lists task 1 titled "Migration Jenkins"

    @task-show-reaches-an-archived-task
    Scenario: an archived task is still reached by its number
      Given I added a task titled "Migration Jenkins"
      And I archived task 1
      When I run `cairn task show 1`
      Then it succeeds
      And it says the task is archived

    @task-archive
    Scenario: a task is archived and stays what it was
      Given I added a task titled "Migration Jenkins"
      And I edited task 1 with the state "done"
      When I run `cairn task archive 1`
      Then it succeeds
      And task 1 is archived
      And task 1 is in the column "done"

    @task-archive-what-was-never-done
    Scenario: work abandoned is archived without being called finished
      Given I added a task titled "Spike Argo"
      When I run `cairn task archive 1`
      Then it succeeds
      And task 1 is archived
      And task 1 is in the column "backlog"

    @task-unarchive
    Scenario: an archived task is brought back
      Given I added a task titled "Migration Jenkins"
      And I archived task 1
      When I run `cairn task unarchive 1`
      Then it succeeds
      And task 1 is not archived

    @task-archive-what-a-narrowing-reaches
    Scenario: everything a narrowing reaches is archived at once
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I added a task titled "Dépôt APP" with the state "done"
      And I added a task titled "Spike Argo"
      When I run `cairn task archive --state done --yes`
      Then it succeeds
      And it says 2 tasks were archived
      And task 1 is archived
      And task 2 is archived
      And task 3 is not archived

    @task-archive-a-sprint
    Scenario: a sprint is cleared by what its tasks carry
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I edited task 1 with the label "sprint-41"
      And I added a task titled "Dépôt APP" with the state "done"
      And I edited task 2 with the label "sprint-42"
      When I run `cairn task archive --state done --label sprint-41 --yes`
      Then it succeeds
      And task 1 is archived
      And task 2 is not archived

    @task-archive-reaching-nothing
    Scenario: a narrowing reaching nothing says so rather than claiming a clean sweep
      Given I added a task titled "Spike Argo"
      When I run `cairn task archive --state done --yes`
      Then it succeeds
      And it says 0 tasks were archived

    @task-archive-a-sweep-needs-confirming
    Scenario: archiving everything a narrowing reaches asks to be confirmed

      Given I added a task titled "Migration Jenkins" with the state "done"
      When I run `cairn task archive --state done`
      Then it fails
      And it says how to confirm
      And task 1 is not archived

    @task-archive-needs-something-to-go-on
    Scenario: archiving with neither a number nor a narrowing is refused
      Given I added a task titled "Spike Argo"
      When I run `cairn task archive`
      Then it fails
      And it says what it needs to go on

    @task-archive-twice
    Scenario: archiving what is already archived is refused rather than counted twice
      Given I added a task titled "Migration Jenkins"
      And I archived task 1
      When I run `cairn task archive 1`
      Then it fails
      And it says the task is already archived

    @task-archive-several
    Scenario: several tasks are archived in one act
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I run `cairn task archive 1 --also 2`
      Then it succeeds
      And `cairn task list` says there is no task

    @task-archive-several-or-none
    Scenario: several tasks are archived all together, or none of them is
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I archived task 2
      When I run `cairn task archive 1 --also 2`
      Then it fails
      And it says the task is already archived
      And `cairn task list` lists task 1 titled "First"

    @task-unarchive-what-is-on-the-board
    Scenario: bringing back what never went away is refused, the way archiving twice is
      Given I added a task titled "Migration Jenkins"
      When I run `cairn task unarchive 1`
      Then it fails
      And it says the task is on the board


  Rule: A sprint archives what was done before it began

    @sprint-archives-what-was-done-before-it
    Scenario: what went done before the sprint began is archived once the tasks are listed
      Given I added a task titled "Migration Jenkins" with the state "done"
      And I added a task titled "Spike Argo" with the state "todo"
      And sprints of 14 days, the first one begun 10 days ago
      And it is 10 days from now
      When I run `cairn task list`
      Then it lists 1 task
      And task 1 is archived
      And task 2 is not archived

    @sprint-keeps-what-went-done-in-it
    Scenario: what went done in the sprint under way stays on the board
      Given I added a task titled "Migration Jenkins" with the state "done"
      And sprints of 14 days, the first one begun 10 days ago
      And it is 3 days from now
      When I run `cairn task list`
      Then it lists 1 task
      And task 1 is not archived

    @sprint-archive-is-said-in-the-trace
    Scenario: the trace says a sprint archived the task
      Given I added a task titled "Migration Jenkins" with the state "done"
      And sprints of 14 days, the first one begun 10 days ago
      And it is 10 days from now
      When I run `cairn task list`
      And I run `cairn trace show 1`
      Then an act says it was archived through "sprint"

    @sprint-lasts-14-days-unless-told
    Scenario: a sprint whose length is not said lasts 14 days
      Given I added a task titled "Migration Jenkins" with the state "done"
      And a sprint begun 10 days ago, its length not said
      And it is 4 days from now
      When I run `cairn task list`
      Then task 1 is archived

    @sprint-leaves-the-rest-alone
    Scenario: only what is done is archived by a sprint
      Given I added a task titled "Migration Jenkins" with the state "todo"
      And I added a task titled "Spike Argo" with the state "done"
      And I edited task 2 with the state "current"
      And sprints of 14 days, the first one begun 10 days ago
      And it is 10 days from now
      When I run `cairn task list`
      Then it lists 2 tasks
      And task 1 is not archived
      And task 2 is not archived

    @sprint-none-archives-nothing
    Scenario: with no sprint configured, nothing is archived on its own
      Given I added a task titled "Migration Jenkins" with the state "done"
      And it is 40 days from now
      When I run `cairn task list`
      Then task 1 is not archived

    @sprint-not-begun-archives-nothing
    Scenario: before the first sprint begins, nothing is archived on its own
      Given I added a task titled "Migration Jenkins" with the state "done"
      And sprints of 14 days, the first one beginning in 3 days
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

    @task-delete-takes-its-steps
    Scenario: deleting a task takes its steps, and leaves the others' alone
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      And I added a task titled "Buy a lamp"
      And I added the step "Measure the desk" to task 2
      When I run `cairn task delete 1 --yes`
      Then it succeeds
      And `cairn step list 2` shows "Measure the desk"

    @task-steps-survive
    Scenario: a task's steps outlive the command that wrote them
      Given I added a task titled "Ship it"
      And I added the step "Write the tests" to task 1
      And cairn has exited
      When I run `cairn task show 1`
      Then it names the step "Write the tests" as not done

  Rule: A task is not done while one of its steps is open

    @task-done-refused-while-a-step-is-open
    Scenario: a task is not finished while a step of it is still open
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --state done`
      Then it fails
      And it says step 1 is in the way
      And task 1 is in the column "backlog"

    @task-done-with-open-steps-says-how-to-confirm
    Scenario: refusing a task with a step open says how to finish both
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --state done`
      Then it fails
      And it says how to confirm

    @task-done-confirmed-ticks-its-steps
    Scenario: a task put in done, confirmed, ticks every step still open
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the docs" to task 1
      And I ticked step 2 of task 1
      When I run `cairn task edit 1 --state done --yes`
      Then it succeeds
      And task 1 is in the column "done"
      And every step of task 1 is done

    @task-moved-to-done-confirmed-ticks-its-steps
    Scenario: a task moved into done, confirmed, ticks every step still open
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task move 1 --state done --first --yes`
      Then it succeeds
      And task 1 is in the column "done"
      And every step of task 1 is done

    @task-done-allowed-once-every-step-is-done
    Scenario: a task is finished once every step of it is
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I ticked step 1 of task 1
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And task 1 is in the column "done"

    @task-done-with-no-steps
    Scenario: a task with no steps is finished as it always was
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And task 1 is in the column "done"

    @task-stopped-is-not-done
    Scenario Outline: a task may still be stopped while a step of it is open
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --state <stopped>`
      Then it succeeds
      And task 1 is in the column "<stopped>"

      Examples: what a task says about itself
        | stopped  |
        | blocked  |
        | watching |
        | paused   |

  Rule: A task's steps are written with it, in one act

    @task-new-with-its-steps
    Scenario: a task is made with its steps
      When I run `cairn task new "Ship it" --steps -` with "[ ] Write the spec" and "[x] Plan it" on standard input
      Then it succeeds
      And `cairn task show 1` names the step "Write the spec" as not done
      And `cairn task show 1` names the step "Plan it" as done

    @task-edit-with-its-steps
    Scenario: a task's steps are written again with it, each line a step
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --steps -` with "[x] Write it" and "[ ] Ship the docs" on standard input
      Then it succeeds
      And `cairn task show 1` names the step "Write it" as done
      And `cairn task show 1` names the step "Ship the docs" as not done
      And `cairn task show 1` does not mention "Write the spec"

    @task-edit-with-no-steps
    Scenario: a task written with no steps keeps none
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      When I run `cairn task edit 1 --steps ""`
      Then it succeeds
      And `cairn task show 1` does not mention "Write the spec"

  Rule: A task says how many of its steps are done

    @task-progress
    Scenario: a task says how many of its steps are done
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the tests" to task 1
      When I run `cairn step edit 1 1 --done`
      And I run `cairn task show 1`
      Then it says 1 of 2 steps are done

    @task-progress-of-nothing
    Scenario: a task with no steps counts nothing
      Given I added a task titled "Ship it"
      When I run `cairn task show 1`
      Then it does not say how much is done

    @task-progress-percent
    Scenario: how much of a task is done is also given as a share
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the tests" to task 1
      And I added the step "Ship the thing" to task 1
      When I run `cairn step edit 1 1 --done`
      And I run `cairn task show 1`
      Then it says the task is 33% done

    @task-json-progress
    Scenario: how many steps are done can be read by a machine
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added the step "Write the tests" to task 1
      When I run `cairn step edit 1 1 --done`
      And I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries 1 of 2 steps done
      And it carries the share 50
      And it carries no count of tasks

    @task-json-progress-of-nothing
    Scenario: a task with no steps carries no progress
      Given I added a task titled "Ship it"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries no progress

    @task-json-progress-is-the-key
    Scenario: the progress read back is the key, never the word
      Given I added a task titled "Report on progress" with the description "no progress yet"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries no progress
      And it carries the title "Report on progress"

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
      When I run `cairn task new "Second"`
      Then it numbers the task 2

    @task-store-created
    Scenario: the first command creates the store
      When I run `cairn task new "Write the spec"`
      Then it succeeds
      And the store exists

    @task-store-elsewhere
    Scenario: told where to live, it does not touch the default place
      When I run `cairn task new "Write the spec"`
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
        | cairn task new                | no title at all           |
        | cairn task new Write the spec | quotes forgotten          |
        | cairn task new "a" "b"        | two titles                |
        | cairn task addd "Write"       | a typo in the subcommand  |
        | cairn task                    | the group with no command |
        | cairn task show               | no number given           |
        | cairn task edit               | no number given           |
        | cairn task delete             | no number given           |

    @task-usage-unknown-flag
    Scenario: an unknown flag is refused rather than ignored
      When I run `cairn task new "Write the spec" --colour red`
      Then it fails
      And it says the flag is unknown

  Rule: It answers the same way every time

    @task-exit-success
    Scenario: what worked says nothing on the error stream
      When I run `cairn task new "Write the spec"`
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
        | cairn task new ""                     | 1    | no title to work with    |
        | cairn task show 99                    | 1    | nothing there to show    |
        | cairn task delete 99 --yes            | 1    | nothing there to delete  |
        | cairn task new                        | 2    | no title given at all    |
        | cairn task new "Spec" --colour red    | 2    | a flag it does not know  |
        | cairn task show spec                  | 2    | a number was expected    |

    @task-exit-store
    Scenario: a store it cannot use has a code of its own
      Given the store holds something that is not a cairn store
      When I run `cairn task list`
      Then it exits with code 3
      And it writes the reason to standard error

  Rule: The output can be read by a machine

    @task-json
    Scenario: asking for machine-readable output
      When I run `cairn task new "Write the spec" --json`
      Then it writes JSON to standard output
      And it carries the title "Write the spec"
      And it carries an identifier

    @task-json-list
    Scenario: listing can be read by a machine
      Given I added a task titled "First"
      When I run `cairn task list --json`
      Then it writes JSON to standard output
      And it carries 1 task

    @task-json-list-says-the-focus-once
    Scenario: a listing says which task is in focus, once
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 2 in focus
      When I run `cairn task list --json`
      Then it carries 2 tasks
      And it says task 2 is the one in focus
      And no task in it carries that on its own

    @task-json-list-leaves-the-description
    Scenario: a listing names the work and leaves what it says to be shown
      Given I added a task titled "First" with the description "A long account of everything done so far"
      And I added a task titled "Second"
      When I run `cairn task list --json`
      Then it carries 2 tasks
      And no task in it carries its description
      And it says task 1 has a description
      And it says task 2 has none

    @task-json-search-leaves-the-description
    Scenario: a search names what it found the same way
      Given I added a task titled "First" with the description "everything done so far"
      When I run `cairn task search First --json`
      Then it carries 1 task
      And no task in it carries its description
      And it says task 1 has a description

    @task-json-show-carries-the-description
    Scenario: showing a task is where its description is read
      Given I added a task titled "First" with the description "everything done so far"
      When I run `cairn task show 1 --json`
      Then it carries the description "everything done so far"

    @task-json-show
    Scenario: showing can be read by a machine
      Given I added a task titled "Write the spec"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries the title "Write the spec"

    @task-json-failure
    Scenario: a refusal is reported the same way it is asked for
      When I run `cairn task new "" --json`
      Then it exits with code 1
      And it writes JSON to standard output
      And the JSON carries the reason
