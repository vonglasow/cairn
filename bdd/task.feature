Feature: Tasks
  A task is the unit of work. It carries a title, and an identity to refer to
  it by afterwards. Everything else — a description, a priority, a column,
  labels, a deadline, an estimate — is optional and has a stated default.

  A task sits in one of seven columns — backlog, todo, current, blocked,
  watching, paused, done — and carries one of four priorities — low, normal,
  high, critical. Nothing else is a column, nothing else is a priority. A new
  task starts in the backlog at normal priority.

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
        | --estimate | 30s        | estimate    | 30s        |

    @task-attribute-refused
    Scenario Outline: an attribute it cannot make sense of is refused
      When I run `cairn task add "Write the spec" <flag> <value>`
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

    @task-list-leaves-sub-tasks-out
    Scenario: a listing names the work, and a sub-task is part of one
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task list`
      Then it lists 1 task
      And it lists task 1 titled "Ship it"
      And it says nothing of task 2

    @task-list-json-carries-sub-tasks
    Scenario: a machine reading the listing is given them all
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task list --json`
      Then it writes JSON to standard output
      And it carries 2 tasks

    @task-show-names-what-hangs-under-it
    Scenario: what hangs under a task is named where the task is shown
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I added a task titled "Write the docs" with the parent "1"
      And I edited task 2 with the state "done"
      When I run `cairn task show 1`
      Then it names task 2 as done
      And it names task 3 as not done

    @task-show-json-carries-what-hangs-under-it
    Scenario: a machine asking about a task is told what hangs under it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I edited task 2 with the state "done"
      When I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries task 2 under it, done

  Rule: A listing can be narrowed to what is wanted

    A list of everything stops being readable once there is a lot of it, and
    stops being usable to an agent, which pays for every line it is handed. So
    a listing takes the facts a task carries from a closed set — the column it
    sits in, the priority it carries, a label, the objective it serves — and
    answers only the tasks carrying them.

    They are named the way they are named everywhere else. A filter that spelled
    a fact differently from the command that sets it would be a second
    vocabulary for the same thing.

    @task-list-narrowed-by-state
    Scenario: only the tasks in one column
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      When I run `cairn task list --state todo`
      Then it lists 1 task
      And it lists task 1 titled "Write the spec"

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

    @task-list-narrowed-is-flat
    Scenario: a narrowed listing answers what matched, not where it sits
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I edited task 2 with the state "todo"
      When I run `cairn task list --state todo`
      Then it lists 1 task
      And it lists task 2 titled "Write the tests"
      And task 2 hangs under nothing in the listing

    @task-list-narrowed-keeps-the-parent-out
    Scenario: a parent that does not carry it is not brought along
      Given I added a task titled "Ship it" with the state "todo"
      # Done, because a sub-task takes done or todo and the parent is in todo:
      # the two have to differ for this to be about the narrowing at all.
      And I added a task titled "Write the tests" with the parent "1"
      And I edited task 2 with the state "done"
      When I run `cairn task list --state todo`
      Then it lists 1 task
      And it lists task 1 titled "Ship it"

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
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      And I put task 1 in focus
      When I run `cairn task list --state todo --json`
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

    Every other command asks for a number, and a number is the one thing
    nobody remembers. So the way back to one is to look for what a person does
    remember — a word from the title, the day it is due — and be told which
    task carries it.

    Looking reaches what a task says in its own words, and reads it the way a
    person does: a duration is looked for as it is shown and not as the count
    of nanoseconds it is kept as. So the title, the description, the deadline,
    the estimate and the time spent.

    It reaches neither the column, the priority nor a label, though it used to.
    Those come from closed sets, and a listing narrows by them — so looking for
    them too would be a second way to ask one question, and a worse one: `search
    high` answered every task whose description merely contained the word. What
    a listing filters, looking leaves alone; what somebody typed themselves,
    looking searches. Nothing is lost and the false answer is gone.

    A task's own number it does reach, and it used not to. The reason it was
    left out held for the command line, where `task show 228` is right there;
    it does not hold on a board, where the search box is the only way in and
    there is no `show` to type. Somebody who knows the number had no way to use
    it. Found by using it, which is where that kind of thing is found.

    It is looked for the way every other thing is — as a substring, so 228
    answers 228 and 1228 both. Weighed and kept: the alternative is a second
    kind of matching in a membrane that has one, and being handed 1228 as well
    costs a glance, while being handed nothing costs the way back.

    What it still does not reach is what a task hangs under or serves. Those
    are numbers a task carries about *other* things, and reaching them would
    answer every sub-task of 228 to somebody looking for 228 — which is the
    half of the old reason that was always true.

    Each thing named has to be found, and any field it still reaches may be
    where it is found. A thing with a space in it is one thing: the shell hands
    `"the spec"` over whole.

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

    @task-search-by-a-number-leaves-what-hangs-under-it-alone
    Scenario: looking for a number does not answer what hangs under it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      When I run `cairn task search 1`
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

    @task-search-does-not-reach-the-column
    Scenario: a column is narrowed by, not looked for
      Given I added a task titled "Write the spec" with the state "current"
      When I run `cairn task search current`
      Then it succeeds
      And it says there is no task

    @task-search-does-not-reach-the-priority
    Scenario: a priority is narrowed by, not looked for
      Given I added a task titled "Write the spec" with the priority "high"
      When I run `cairn task search high`
      Then it succeeds
      And it says there is no task

    @task-search-does-not-reach-a-label
    Scenario: a label is narrowed by, not looked for
      Given I added a task titled "Write the spec" with the labels "team" and "urgent"
      When I run `cairn task search urgent`
      Then it succeeds
      And it says there is no task

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

    @task-search-reaches-a-sub-task
    Scenario: a sub-task is looked for like any other, and listed like one
      Given I added a task titled "Write the spec"
      And I added a task titled "Read it first" with the parent "1"
      When I run `cairn task search first`
      Then it lists task 2 titled "Read it first"
      And task 2 hangs under nothing in the listing

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

    A label given plainly says what the set becomes, so adding one means typing
    them all again. A label given with a leading + is added to what is there,
    one with a leading - is taken away, and the rest is left alone. The two
    cannot be mixed: one says what the set becomes, the other says how it
    changes, and a command that did both would have to guess which won.

    @task-label-list
    Scenario: several labels can be given in one flag
      When I run `cairn task add "Write the spec" --label "spec,team"`
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
      When I run `cairn task add "Write the spec" --label "spec,,team,"`
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
      When I run `cairn task add "Write the spec" --label +urgent`
      Then it fails
      And it says a label is added or taken away only when changing a task

  Rule: A task can say what blocks it

    A task stops for one of two reasons: another task, or something that is not
    in cairn at all. The first has a number, so it is named by one. The second
    has none, so it is named in words. Both land in the same place — what blocks
    this — and the card shows them together.

    Two flags rather than one, because a single flag taking either would have to
    guess whether "3" is task 3 or the word "3", and nothing here guesses.

    A blocker can sit on a task in any column: it may block part of the work and
    not the rest. When it blocks the whole thing, the column is what says so —
    and splitting the part that is stopped into a sub-task of its own is the
    better move, which cairn suggests nowhere and refuses never.

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

    @task-blocker-a-sub-task
    Scenario: a sub-task blocks like anything else
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      And I added a task titled "Announce it"
      When I run `cairn task edit 3 --state blocked --blocker 2`
      Then `cairn task show 3` says task 2 blocks it

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

    A task in the blocked column with something recorded against it cannot be
    put back in todo or current. Not a confirmation — a refusal: those two
    columns say the work is yours to do, and a task still waiting on somebody
    else is not.

    Taking the last thing away is what puts it back in todo, because that is
    the one moment somebody said out loud that nothing is in the way. Taking one
    of two away changes nothing: something is still in the way.

    The other columns stay reachable. Deciding a blocked task is done, or
    putting it back in the backlog, or pausing it, or handing it over to watch
    are all things you can say about work that is waiting — only claiming it is
    yours to do now is refused.

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

    A blocker going done does not move the task that was waiting on it. Cairn
    proposes, it does not decide — and a column changed by nobody is a change
    the trace cannot attribute. Nothing ever leaves the blocked column on its
    own: taking the last thing away is what does that, and that is somebody
    saying so.

    But it says so, because otherwise a blocked task waits forever on somebody
    remembering to look. The listing marks a blocked task whose every named
    blocker is done: that mark is what says it is your turn.

    Only when blockers were named and all of them are done. A task blocked by
    words has words cairn cannot check — "the client's answer" is either in or
    out, and cairn has no way to know which — and a task blocked by nothing
    named is blocked by something cairn was never told about. Neither gets the
    mark, because in both cases claiming nothing is in the way would be a guess.

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

  Rule: Time spent on a task can be written down

    A task carries the time already spent on it, and that time accumulates.
    Given plainly it says what the total becomes; with a leading + or - it
    moves what is there, the way a label joins or leaves a set. That is why a
    minus is a value here and a refusal on --estimate: an estimate is a guess
    about the whole, time spent is a tally that grows.

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

  Rule: A task can hang under another one

    A task given a parent becomes one of its sub-tasks, and stays a task in
    every other way: its own number, its own title, its own description, its
    own labels, changed and read back like any other. Only one level — a
    sub-task cannot itself carry sub-tasks, so what hangs under a task is never
    something with more hanging under it.

    What it stops carrying is a column of its own. A sub-task is a part of the
    work its parent names, split off to be sure the whole is finished, and the
    only thing worth knowing about a part is whether it is done. It had all
    seven columns and used two: measured on a store of 219 tasks, the 49
    sub-tasks in it were `done` forty-eight times and `todo` once, never
    `current`, never stopped, never with a deadline or a blocker of their own.

    So `done` and `todo` are what a sub-task takes, and any other column is
    refused. Where the work stands is the parent's to say — a part being
    "blocked" is the parent being blocked, and a part being "current" is the
    parent being worked on.

    A task carrying sub-tasks is not deleted while they are there. What depends
    on it would go with it, and a deletion that takes more than the number
    typed is a deletion nobody asked for.

    @task-sub-task-takes-done
    Scenario Outline: a sub-task is done or it is not, and takes nothing else
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task edit 2 --state <column>`
      Then it succeeds
      And task 2 is in the column "<column>"

      Examples: what a part of the work can say about itself
        | column |
        | done   |
        | todo   |

    @task-sub-task-refuses-a-column-of-its-own
    Scenario Outline: a sub-task refuses the columns that say where work stands
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task edit 2 --state <column>`
      Then it fails
      And it says a sub-task is done or it is not

      Examples: where the work stands is the parent's to say
        | column   |
        | backlog  |
        | current  |
        | blocked  |
        | watching |
        | paused   |

    @task-hanging-a-task-under-another-settles-its-column
    Scenario: a task given a parent keeps done and loses any other column
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the state "current"
      When I run `cairn task edit 2 --parent 1`
      Then it succeeds
      And task 2 is in the column "todo"

    @task-hanging-a-finished-task-under-another-keeps-it-done
    Scenario: a task already done keeps being done when it is hung under one
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the state "done"
      When I run `cairn task edit 2 --parent 1`
      Then it succeeds
      And task 2 is in the column "done"

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
    Scenario: a sub-task keeps a number of its own, and is reached by it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task show 2`
      Then it succeeds
      And it lists task 2 titled "Write the tests"

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

  Rule: A task is not done while something under it is open

    A parent is done because what hangs under it is done. If it can be
    finished while a sub-task is still open, then either the sub-task was not
    part of it — in which case it is another task — or the parent was closed
    over work nobody did.

    Refused rather than cascaded. Marking the children done on the parent's
    behalf would finish work nobody finished, and it would do it silently; the
    refusal says which ones are in the way, which is the useful thing to be
    told. Every other stopped state stays open to a parent: blocked, watching
    and paused say something about the parent itself, and done is the only one
    that claims something about what hangs under it.

    @task-done-refused-while-a-sub-task-is-open
    Scenario: a parent is not finished while a sub-task is still open
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      When I run `cairn task edit 1 --state done`
      Then it fails
      And it says task 2 is in the way
      And task 1 is in the column "backlog"

    @task-done-allowed-once-they-are-all-done
    Scenario: a parent is finished once everything under it is
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      And I edited task 2 with the state "done"
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And task 1 is in the column "done"

    @task-done-with-nothing-under-it
    Scenario: a task nothing hangs under is finished as it always was
      Given I added a task titled "Ship it"
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And task 1 is in the column "done"

    @task-stopped-is-not-done
    Scenario Outline: a parent may still be stopped while a sub-task is open
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      When I run `cairn task edit 1 --state <stopped>`
      Then it succeeds
      And task 1 is in the column "<stopped>"

      Examples: what a parent says about itself
        | stopped  |
        | blocked  |
        | watching |
        | paused   |

  Rule: A task says how many of its sub-tasks are done

    A task counts what hangs under it, never itself: a task with two sub-tasks
    of which one is done is one of two, not one of three. Nobody reads it the
    other way round.

    @task-progress
    Scenario: a task says how many of its sub-tasks are done
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task edit 2 --state done`
      And I run `cairn task show 1`
      Then it says 1 of 2 sub-tasks are done

    @task-progress-counts-only-done
    Scenario: a sub-task not done is not a sub-task done
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      And I added a task titled "Write the tests" with the parent "1"
      And I edited task 2 with the state "done"
      When I run `cairn task show 1`
      Then it says 1 of 2 sub-tasks are done

    @task-progress-of-nothing
    Scenario: a task with nothing under it counts nothing
      Given I added a task titled "Ship it"
      When I run `cairn task show 1`
      Then it does not say how much is done

    @task-progress-percent
    Scenario: how much of a task is done is also given as a share
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      And I added a task titled "Write the tests" with the parent "1"
      And I added a task titled "Ship the thing" with the parent "1"
      When I run `cairn task edit 2 --state done`
      And I run `cairn task show 1`
      Then it says the task is 33% done

    @task-json-progress
    Scenario: how many sub-tasks are done can be read by a machine
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" with the parent "1"
      And I added a task titled "Write the tests" with the parent "1"
      When I run `cairn task edit 2 --state done`
      And I run `cairn task show 1 --json`
      Then it writes JSON to standard output
      And it carries 1 of 2 sub-tasks done
      And it carries the share 50
      And it carries no count of tasks

    @task-json-progress-of-nothing
    Scenario: a task with nothing under it carries no progress
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

    @task-json-list-says-the-focus-once
    Scenario: a listing says which task is in focus, once
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I put task 2 in focus
      When I run `cairn task list --json`
      Then it carries 2 tasks
      And it says task 2 is the one in focus
      And no task in it carries that on its own

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
