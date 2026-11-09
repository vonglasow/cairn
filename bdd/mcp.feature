Feature: Answering a machine over MCP

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: The server is started like any other command

    @mcp-serve-speaks-the-protocol
    Scenario: the server answers a client that greets it
      When I start `cairn mcp serve` and greet it
      Then it answers with its name and its version

    @mcp-client-hangs-up
    Scenario: nothing is owed to a client that has gone
      When I start `cairn mcp serve` and greet it without waiting
      Then it stops, and says nothing went wrong

    @mcp-needs-a-verb
    Scenario: the group on its own says how it is used
      When I run `cairn mcp`
      Then it fails
      And it says how to use it

    @mcp-access-refused
    Scenario: an access it cannot make sense of is refused
      When I run `cairn mcp serve --access sideways`
      Then it fails
      And it says what it expects instead of "sideways"

  Rule: The server holds nothing between two calls

    @mcp-a-backup-taken-through-it-holds-everything
    Scenario: a backup taken through the server holds what was just added
      Given a cairn server started with `--access full`
      When I call `task_new` with the title "Added just now"
      And I call `backup_new`
      And I call `backup_restore` with the number 1
      Then `cairn task list` lists 1 task

    @mcp-sees-what-the-terminal-did
    Scenario: what was done at the terminal is there on the next call
      Given a cairn server started with `--access write`
      And I asked it to add a task titled "From the agent"
      When I run `cairn task new "From the terminal"`
      And I call `task_list`
      Then it answers with 2 tasks

    @mcp-what-it-did-is-on-disk
    Scenario: what the server did is there for the terminal
      Given a cairn server started with `--access write`
      When I call `task_new` with the title "From the agent"
      Then `cairn task list` lists 1 task

  Rule: What the server offers is what it was allowed to do

    @mcp-read-offers-reading
    Scenario: reading only, which is the default
      Given a cairn server started with no access given
      When I ask it what it can do
      Then it offers `task_list`
      And it does not offer `task_new`
      And it does not offer `task_delete`

    @mcp-write-offers-changing
    Scenario: reading and changing
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then it offers `task_list`
      And it offers `task_new`
      And it does not offer `task_delete`

    @mcp-full-offers-everything
    Scenario: everything, including what loses work
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it offers `task_delete`
      And it offers `backup_restore`
      And it offers `migration_rollback`

    @mcp-offers-every-command
    Scenario: every command is offered, under the name it has on the command line
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it offers a tool for every command cairn has
      And each one is named for the command it runs

    @mcp-offers-nothing-of-its-own
    Scenario: the server is not something an agent can reach
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `mcp_serve`

    @workspace-an-agent-cannot-add-or-switch
    Scenario: an agent can read the workspaces, and neither add one nor switch to another
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it offers `workspace_list`
      And it offers `workspace_show`
      And it does not offer `workspace_add`
      And it does not offer `workspace_use`

    @workspace-an-agent-cannot-remove-one
    Scenario: an agent cannot remove a workspace
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `workspace_remove`

  Rule: A tool says what calling it would cost

    @mcp-marks-what-loses-work
    Scenario Outline: what loses work is marked as such
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then `<tool>` is marked as losing work

      Examples: the five that take something away
        | tool               |
        | task_delete        |
        | objective_delete   |
        | backup_delete      |
        | backup_restore     |
        | migration_rollback |

    @mcp-marks-what-only-reads
    Scenario Outline: what only reads is marked as such
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then `<tool>` is marked as read-only

      Examples: looking changes nothing
        | tool           |
        | task_list      |
        | task_show      |
        | objective_list |
        | migration_show |

    @mcp-marks-what-it-never-reaches
    Scenario: nothing here goes looking outside
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then every tool is marked as reaching nothing outside cairn

  Rule: A tool does what the command of the same name does

    @mcp-add-a-task
    Scenario: adding a task
      Given a cairn server started with `--access write`
      When I call `task_new` with the title "Write the spec"
      Then it answers without an error
      And `cairn task list` lists 1 task

    @mcp-a-task-opened-with-a-link
    Scenario: a task opened through the MCP with a link is linked
      Given a cairn server started with `--access write`
      And I added a task titled "Ship it"
      When I call `task_new` with the title "It does not hold" and the link 1
      Then it answers without an error
      And `cairn task show 1` says "linked       #2"

    @mcp-what-it-answers
    Scenario: what a tool answers is what the command prints
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_show` with the number 1
      Then it answers with the task titled "Write the spec"

    @mcp-answers-both-ways
    Scenario: an agent and a script are told the same thing
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_show` with the number 1
      Then it answers with the card a person would have read
      And it answers with the shape `--json` would have printed

    @mcp-a-listing-fits-an-agent
    Scenario: a listing leaves the descriptions out, so an agent can read it whole
      Given I added a task titled "Write the spec" with the description "A long account of everything done so far"
      And a cairn server started with `--access read`
      When I call `task_list`
      Then it answers with 1 task
      And no row of it carries a description
      And its answer holds to the shape the tool described

    @mcp-a-listing-carries-the-fields-asked-for
    Scenario: a listing carries only the fields an agent asked for
      Given I added a task titled "Write the spec" with the description "A long account of everything done so far"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the fields "number", "state" and "title"
      Then it answers with 1 task
      And every row of it carries "number", "state" and "title", and nothing else
      And its answer holds to the shape the tool described

    @mcp-a-search-carries-the-fields-asked-for
    Scenario: a search carries only the fields an agent asked for
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec", asking for the fields "number", "state" and "title"
      Then it answers with 1 task
      And every row of it carries "number", "state" and "title", and nothing else
      And its answer holds to the shape the tool described

    @mcp-a-search-finds-any-of-several-terms
    Scenario: a search asked for any of several terms finds each task carrying one, and says which term matched which
      Given I added a task titled "Port T010 to the new board"
      And I added a task titled "Merge T013 and T022"
      And I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for any of "T010", "T013", "T022" and "T999"
      Then it answers with 2 tasks
      And it says "T010" matched task 1
      And it says "T013" matched task 2
      And it says "T022" matched task 2
      And it says "T999" matched no task
      And its answer holds to the shape the tool described

    @mcp-which-term-matched-whatever-the-fields
    Scenario: which term matched is said whatever fields a row was asked to carry
      Given I added a task titled "Port T010 to the new board"
      And a cairn server started with `--access read`
      When I call `task_search` looking for any of "T010" and "T013", asking for the fields "number", "state" and "title"
      Then it answers with 1 task
      And every row of it carries "number", "state" and "title", and nothing else
      And it says "T010" matched task 1
      And it says "T013" matched no task
      And its answer holds to the shape the tool described

    @mcp-a-search-for-everything-named-says-no-term
    Scenario: a search for everything named does not say which term matched
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec"
      Then it answers with 1 task
      And it says nothing of which term matched

    @mcp-a-search-answers-a-page
    Scenario: a search asked for a page answers those of what it kept in the board's order, and says how many it kept in all
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Review the spec"
      And I added a task titled "Print the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec", asking for 1 task after the first 1
      Then it answers with the task 3
      And it says there are 3 tasks in all
      And its answer holds to the shape the tool described

    @mcp-a-search-page-says-what-matched-on-it
    Scenario: a page of a search for any of several terms says which term matched only of the tasks on the page
      Given I added a task titled "Port T010 to the new board"
      And I added a task titled "Merge T013"
      And I added a task titled "Close T010"
      And a cairn server started with `--access read`
      When I call `task_search` looking for any of "T010" and "T013", asking for 1 task after the first 1
      Then it answers with the task 2
      And it says "T013" matched task 2
      And it says "T010" matched no task
      And it says there are 3 tasks in all
      And its answer holds to the shape the tool described

    @mcp-a-search-page-past-the-last-task-is-empty
    Scenario: a page of a search past the last task it kept answers none, and still says how many it kept
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec", asking for 2 tasks after the first 5
      Then it answers with 0 tasks
      And it says there is 1 task in all

    @mcp-a-search-asked-for-no-page-says-no-total
    Scenario: a search asked for no page says nothing of how many it kept in all
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec"
      Then it answers with 1 task
      And it says nothing of how many there are in all

    @mcp-a-search-page-that-is-none-is-refused
    Scenario Outline: a page of a search of no task, or one starting before the first, is refused
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec", asking for <limit> tasks after the first <offset>
      Then it answers that it failed
      And it says the refusal is "usage", because "not-understood"

      Examples:
        | limit | offset |
        | 0     | 0      |
        | -1    | 0      |
        | 2     | -1     |

    @mcp-a-listing-answers-the-numbers-asked-for
    Scenario: a listing asked for several numbers answers those tasks, and says which number is no task
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Tell the team"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 3, 1 and 9
      Then it answers with the tasks 1 and 3
      And it says there is no task 9
      And its answer holds to the shape the tool described

    @mcp-a-listing-by-number-reaches-an-archived-task
    Scenario: a number asked for answers its task, archived or not
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I archived task 1
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 1, 2 and 2
      Then it answers with the tasks 1 and 2
      And it says every number asked for is a task

    @mcp-a-listing-by-number-is-narrowed-like-every-listing
    Scenario: a listing by number still keeps only what everything else named holds for
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I archived task 2
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 1, 2 and 7, archived only
      Then it answers with the task 2
      And it says there is no task 7

    @mcp-a-listing-by-number-carries-the-fields-asked-for
    Scenario: which number is no task is said whatever fields a row was asked to carry
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 1 and 4, and for the fields "number", "state" and "title"
      Then it answers with the task 1
      And every row of it carries "number", "state" and "title", and nothing else
      And it says there is no task 4
      And its answer holds to the shape the tool described

    @mcp-asking-for-no-number-is-refused
    Scenario: asking for an empty list of numbers is refused
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for no number
      Then it answers that it failed
      And the answer says a task number

    @mcp-a-listing-takes-numbers-written-as-numbers
    Scenario: a listing takes the numbers it is asked for written as numbers, not only as text
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Tell the team"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 3, 1 and 9, written as numbers
      Then it answers with the tasks 1 and 3
      And it says there is no task 9

    @mcp-a-blocker-and-a-link-take-numbers-written-as-numbers
    Scenario: a task's blocker and link are taken written as numbers
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `task_new` with the title "Ship it", blocked by 1 and linked to 1, written as numbers
      Then it answers with a task blocked by task 1 and linked to task 1

    @mcp-tasks-taken-together-take-numbers-written-as-numbers
    Scenario: tasks taken together are named written as numbers
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And a cairn server started with `--access write`
      When I call `task_archive` with the number 1 and also 2, written as a number
      Then it answers without an error
      And `cairn task list --archived` lists 2 tasks

    @mcp-a-list-of-numbers-says-it-takes-numbers
    Scenario: every list of numbers a tool takes says it takes whole numbers
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then every list of task or objective numbers it takes says it takes whole numbers

    @mcp-what-is-no-number-is-refused-as-it-was-sent
    Scenario Outline: what is no task number among the numbers is refused, saying what was sent
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the numbers 1 and <sent>
      Then it answers that it failed
      And the answer says <said>

      Examples:
        | sent  | said                                               |
        | 1.5   | "1.5": a task number is expected                   |
        | 1.0   | "1.0": a task number is expected                   |
        | -2    | "-2": a task number is expected                    |
        | true  | a task number is expected among the list, got true |
        | "one" | "one": a task number is expected                   |

    @mcp-a-number-among-labels-is-refused-as-a-number
    Scenario: a number sent among labels is refused as the number it is, not as text
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `task_edit` with the number 1 and the label 3, written as a number
      Then it answers that it failed
      And the answer says a string is expected among the list, got 3

    @mcp-a-listing-answers-a-page
    Scenario: a listing asked for a page answers those tasks in the board's order, and says how many there are in all
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Tell the team"
      And a cairn server started with `--access read`
      When I call `task_list` asking for 2 tasks after the first 1
      Then it answers with the tasks 2 and 3
      And it says there are 3 tasks in all
      And its answer holds to the shape the tool described

    @mcp-a-page-is-cut-from-what-is-kept
    Scenario: a page is cut from what everything else named keeps
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it"
      And I added a task titled "Tell the team" with the state "todo"
      And a cairn server started with `--access read`
      When I call `task_list` in "todo", asking for 1 task after the first 1
      Then it answers with the task 3
      And it says there are 2 tasks in all

    @mcp-a-page-past-the-last-task-is-empty
    Scenario: a page past the last task answers none, and still says how many there are
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for 2 tasks after the first 5
      Then it answers with 0 tasks
      And it says there is 1 task in all

    @mcp-a-listing-asked-for-no-page-says-no-total
    Scenario: a listing asked for no page says nothing of how many there are in all
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list`
      Then it answers with 1 task
      And it says nothing of how many there are in all

    @mcp-a-page-that-is-none-is-refused
    Scenario Outline: a page of no task, or one starting before the first, is refused
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for <limit> tasks after the first <offset>
      Then it answers that it failed
      And it says the refusal is "usage", because "not-understood"

      Examples:
        | limit | offset |
        | 0     | 0      |
        | -1    | 0      |
        | 2     | -1     |

    @mcp-a-listing-counts-by-state
    Scenario: a listing asked to count by state answers how many tasks sit in each column, and no row
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it" with the state "todo"
      And I added a task titled "Tell the team" with the state "done"
      And a cairn server started with `--access read`
      When I call `task_list` counting by "state"
      Then it answers with 0 tasks
      And it counts 2 under "todo"
      And it counts 1 under "done"
      And it counts nothing under "current"
      And it says there are 3 tasks in all
      And its answer holds to the shape the tool described

    @mcp-a-listing-counts-by-label
    Scenario: a listing asked to count by label answers how many of what is kept carry each one
      Given I added a task titled "Write the spec" with the labels "team" and "spec"
      And I added a task titled "Ship it" with the labels "team" and "release"
      And I added a task titled "Tell the team"
      And I archived task 2
      And a cairn server started with `--access read`
      When I call `task_list` counting by "label"
      Then it answers with 0 tasks
      And it counts 1 under "team"
      And it counts 1 under "spec"
      And it counts nothing under "release"
      And it says there are 2 tasks in all

    @mcp-counting-takes-no-row-argument
    Scenario Outline: counting, which answers no row, refuses what shapes a row or a page
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` counting by "state", with a <argument> as well
      Then it answers that it failed
      And it says the refusal is "usage", because "not-understood"

      Examples:
        | argument |
        | field    |
        | limit    |
        | offset   |

    @mcp-counting-by-what-it-cannot-is-refused
    Scenario: counting by anything but state or label is refused, naming both
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` counting by "priority"
      Then it answers that it failed
      And it says the refusal is "refused", because "not-one-of"
      And the answer says state, label

    @mcp-a-listing-asked-for-no-field-carries-every-one
    Scenario: a listing that names no field carries every one, as it always did
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list`
      Then every row of it carries every field a row can carry

    @mcp-a-field-a-row-does-not-carry-is-refused
    Scenario: a field a row does not carry is refused, naming the ones it does
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for the fields "number", "state" and "description"
      Then it answers that it failed
      And it says the refusal is "refused", because "not-a-field"
      And the answer names every field a row can carry

    @mcp-asking-for-no-field-is-refused
    Scenario: asking for an empty list of fields is refused
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list` asking for no field
      Then it answers that it failed
      And the answer names every field a row can carry

    @mcp-names-the-fields-it-can-answer
    Scenario Outline: a listing tool names the fields an agent can ask for
      Given a cairn server started with `--access read`
      When I ask it what it can do
      Then `<tool>` names every field a row can carry

      Examples: the two that answer rows
        | tool        |
        | task_list   |
        | task_search |

    @mcp-what-an-agent-does-is-written-down-as-its-own
    Scenario: what an agent changes says it came through the server, and which agent it was
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `task_edit` with the number 1 and the priority "high"
      Then the trace of task 1 says the change came through the server
      And the trace of task 1 says the change was made by the agent that asked

    @mcp-says-what-it-answers
    Scenario: every tool says what shape its answer takes
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then every tool describes the shape of its answer

    @mcp-every-tool-answers-an-object
    Scenario: every tool answers something a machine reads as one thing
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then every tool says it answers one object

    @mcp-says-what-one-thing-it-answers
    Scenario: a tool says which fields its one answer carries
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then `task_list` says its answer carries the tasks
      And `task_show` says its answer carries a title

    @mcp-what-it-answers-has-the-shape-it-said
    Scenario: what a tool answers is measured against what it said
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `task_list`
      Then its answer holds to the shape the tool described

    @mcp-a-refusal-is-not-measured-against-that-shape
    Scenario: a refusal is not an answer of that shape
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `task_show` with the number 99
      Then it answers that it failed
      And it carries nothing for a machine to measure

    @mcp-a-refusal-says-its-kind
    Scenario: a refusal says what kind it is, for a machine to branch on
      Given a cairn server started with `--access read`
      When I call `task_show` with the number 99
      Then it answers that it failed
      And it says the refusal is "refused", because "no-such-task"

    @mcp-labels-arrive-whole
    Scenario: labels given to a tool arrive as labels
      Given a cairn server started with `--access write`
      When I call `task_new` with the title "From the agent" and the labels "spec" and "team"
      Then it answers with a task carrying both labels

    @mcp-a-label-joins-the-others
    Scenario: a label an agent gives a task joins the ones it carries, and the answer says so
      Given I added a task titled "Write the spec" with the labels "spec" and "team"
      And a cairn server started with `--access write`
      When I call `task_edit` with the number 1 and the label "urgent"
      Then it answers with a task carrying 3 labels
      And its answer says the labels added are "urgent"

    @mcp-a-piece-of-a-description-is-replaced
    Scenario: an agent changes one piece of a description in one call
      Given I added a task titled "Write the spec" with the description "built on main at abc123"
      And a cairn server started with `--access write`
      When I call `task_edit` with the number 1, replacing "abc123" by "def456"
      Then it answers without an error
      And `cairn task show 1` says "built on main at def456"

    @mcp-a-key-it-does-not-know-does-not-stop-it
    Scenario: a configuration holding a key this cairn does not know does not stop the server
      Given a configuration that turns the page "calendar" off
      And a cairn server started with `--access read`
      When I call `task_list`
      Then it answers without an error

    @mcp-looks-for-what-it-was-told
    Scenario: what a tool is told to look for arrives whole
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for "spec"
      Then it answers with 1 task

    @mcp-what-it-needs-is-asked-for
    Scenario: a tool asked for nothing it needs says so
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` with nothing at all
      Then it answers that it failed
      And the answer says what it needed

    @mcp-an-empty-list-is-nothing-to-look-for
    Scenario: a tool handed an empty list to look for says so
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_search` looking for nothing
      Then it answers that it failed
      And the answer says what it needed

    @mcp-an-argument-it-does-not-take
    Scenario: a tool is told the name it did not recognise
      Given a cairn server started with `--access read`
      When I call `task_show` naming an argument it does not take
      Then it answers that it failed
      And the answer says which name it could not read

    @mcp-says-what-it-accepts
    Scenario: a tool names what it takes
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then `task_new` names every state cairn accepts
      And it says what each thing it takes is for

    @mcp-empties-a-field
    Scenario: handed nothing a field is emptied, left out it is untouched
      Given I added a task titled "Write the spec" with the deadline "2026-09-01"
      And a cairn server started with `--access write`
      When I call `task_edit` with the number 1 and the deadline as nothing
      Then the task has no deadline
      And it is still titled "Write the spec"

    @objective-sub-mcp-parent-and-children
    Scenario: a listing of objectives says which sits under which
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And a cairn server started with `--access read`
      When I call `objective_list`
      Then it answers objective 1 under nothing, holding 2
      And it answers objective 2 under 1, holding nothing
      And its answer holds to the shape the tool described

    @objective-sub-mcp-made-under-a-parent
    Scenario: an objective is made under another, named by its number
      Given I added an objective titled "Home automation"
      And a cairn server started with `--access write`
      When I call `objective_new` with {"title": "Enclosure", "parent": 1}
      Then it answers objective 2 under 1, holding nothing

    @objective-sub-mcp-task-list-itself
    Scenario: a listing by a parent reaches its sub-objectives, or keeps to it alone when asked
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Print the case" with the objective "2"
      And I added a task titled "Wiring diagram" with the objective "1"
      And a cairn server started with `--access read`
      When I call `task_list` with {"objective": [1]}
      Then it answers with the tasks 1 and 2
      When I call `task_list` with {"objective": [1], "itself": true}
      Then it answers with the task 2

    @objective-sub-mcp-delete-asks
    Scenario: deleting a parent lists what it touches until it is told what goes with it
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And a cairn server started with `--access full`
      When I call `objective_delete` with {"number": 1}
      Then it answers that it failed
      And the answer says #2  Enclosure
      When I call `objective_delete` with {"number": 1, "sub-objectives": "keep"}
      Then it answers without an error
      And `cairn objective list` lists 1 objective

    @label-scope-mcp
    Scenario: an agent makes a label for objectives only, and is refused putting it on a task
      Given I added a task titled "Write the spec"
      And a cairn server started with `--access write`
      When I call `label_new` with {"label": "area", "scope": "objectives"}
      Then it answers without an error
      And the answer says for objectives
      And its answer holds to the shape the tool described
      When I call `task_edit` with the number 1 and the label "area"
      Then it answers that it failed
      And the answer says "area" is a label for objectives only
      When I call `label_scope` with {"label": "area", "scope": "both"}
      Then it answers without an error
      When I call `label_list` with {"for": "tasks"}
      Then the answer says area

    @objective-waits-on-mcp
    Scenario: an agent says an objective waits on another, and is warned when work starts on it
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And I added a task titled "Dimmer scenes" with the objective "2"
      And a cairn server started with `--access write`
      When I call `objective_edit` with {"number": 2, "waits-on": [1]}
      Then it answers without an error
      And the answer says #1 Firmware
      And its answer holds to the shape the tool described
      When I call `task_move` with {"number": 1, "state": "current"}
      Then it answers without an error
      And the answer says #2 Features waits on #1 Firmware, not reached
      And its answer holds to the shape the tool described

  Rule: A refusal is a refusal, not a broken tool

    @mcp-refusal-is-answered
    Scenario: a refusal comes back as the reason it was refused
      Given a cairn server started with `--access write`
      When I call `task_new` with the title ""
      Then it answers that it failed
      And the answer says a task needs a title

    @mcp-unknown-tool
    Scenario: asking for a tool that was not offered
      Given a cairn server started with `--access read`
      When I call `task_delete` with the number 1
      Then the call itself fails

    @mcp-store-behind-is-answered
    Scenario: a store this cairn cannot use comes back as the reason
      Given a store from an older cairn, holding a task titled "Write the spec"
      And a cairn server started with `--access read`
      When I call `task_list`
      Then it answers that it failed
      And the answer says the store is behind

  Rule: An agent is told where its work on a task goes

    @mcp-greeting-says-where-the-work-goes
    Scenario: an agent is told as it greets the server
      When I start `cairn mcp serve` and greet it
      Then it tells the agent that what was done or decided on a task goes in its notes, never in its description

    @mcp-description-points-at-the-notes
    Scenario: the tool that edits a task says it where the description is
      Given a cairn server started with `--access write`
      When I ask it what it can do
      Then the description `task_edit` takes says what was done or decided on a task goes in its notes

  Rule: An agent is asked to renew a focus session past its length, whatever it calls

    @mcp-any-answer-asks-to-renew-a-focus-session
    Scenario: any answer asks the agent to renew a focus session past its length
      Given a cairn server started with no access given
      And focus sessions last "1h"
      And I added a task titled "Ship it"
      And I put task 1 in focus
      And it is 2 hours later
      When I call `task_list`
      Then the answer says this focus session has been up 2h: renew it by focusing task 1 again, or end it by unfocusing task 1

    @mcp-no-answer-asks-within-a-focus-sessions-length
    Scenario: no answer asks anything of a focus session within its length
      Given a cairn server started with no access given
      And focus sessions last "1h"
      And I added a task titled "Ship it"
      And I put task 1 in focus
      And it is 30 minutes later
      When I call `task_list`
      Then the answer does not say "renew"

  Rule: A page turned off in the configuration offers no tool

    @mcp-a-page-off-offers-none-of-its-tools
    Scenario: the time page turned off takes its tools away, and leaves the rest
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `timesheet_show`
      And it does not offer `timesheet_add`
      And it offers `task_list`

    @mcp-every-page-off-offers-none-of-theirs
    Scenario: every page turned off leaves none of their tools, and the rest
      Given a configuration that turns every page off
      And a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `task_list`
      And it does not offer `objective_list`
      And it does not offer `timesheet_show`
      And it offers `config_show`

    @settings-the-server-reads-the-settings-of-its-store
    Scenario: the settings beside the store a server answers from turn a page off
      Given the settings beside the store turn the page "time" off
      And a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `timesheet_show`
      And it offers `task_list`

  Rule: A page turned on or off while the server runs is offered or taken away at the next call, with no reconnect

    @mcp-a-page-turned-back-on-is-offered-at-the-next-call
    Scenario: the time page turned back on offers its tools at the next call, and the client is told
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access full`
      And the page "time" turned back on
      When I call `task_list`
      Then the client is told the tools changed
      And I ask it what it can do
      And it offers `timesheet_show`
      And it offers `timesheet_add`

    @mcp-a-page-turned-off-is-taken-away-at-the-next-call
    Scenario: the time page turned off takes its tools away at the next call, and leaves the rest
      Given a cairn server started with `--access full`
      And the page "time" turned off
      When I call `task_list`
      Then the client is told the tools changed
      And I ask it what it can do
      And it does not offer `timesheet_show`
      And it offers `task_list`

    @mcp-a-page-turned-back-on-stays-within-the-access
    Scenario: a page turned back on offers only the tools the server was allowed
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access read`
      And the page "time" turned back on
      When I call `task_list`
      Then the client is told the tools changed
      And I ask it what it can do
      And it offers `timesheet_show`
      And it does not offer `timesheet_add`

    @mcp-pages-unchanged-are-not-announced
    Scenario: a configuration whose pages did not change says nothing about the tools
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access full`
      And the configuration written again, the page "time" still off and the weekends hidden
      When I call `task_list`
      And I call `task_list`
      Then the client is not told the tools changed

    @mcp-a-configuration-that-cannot-be-read-mid-session-turns-no-page-off
    Scenario: a configuration broken while the server runs turns no page off, as it does at the start
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access full`
      And a configuration that is not TOML
      When I call `task_list`
      Then the client is told the tools changed
      And I ask it what it can do
      And it offers `timesheet_show`

  Rule: A cairn installed while the server runs takes over from it, with no reconnect

    @mcp-a-newer-cairn-takes-over
    Scenario: the cairn now on disk answers the next call, and the client greets nobody again
      Given a configuration that turns the page "time" off
      And a cairn server started from a cairn on disk
      And the page "time" turned back on
      And a newer cairn installed in its place
      When I call `timesheet_show`
      Then it answers without an error
      And the client is told the tools changed

    @mcp-the-same-tools-are-not-announced
    Scenario: a newer cairn offering the same tools says nothing about them
      Given a cairn server started from a cairn on disk
      And a newer cairn installed in its place
      When I call `task_list`
      Then it answers without an error
      And the client is not told the tools changed

    @mcp-what-cannot-run-leaves-the-running-cairn
    Scenario: something installed that cannot be run leaves the running cairn answering
      Given a cairn server started from a cairn on disk
      And something that cannot be run installed in its place
      When I call `task_list`
      Then it answers without an error
      And I call `task_list`
      And it answers without an error
