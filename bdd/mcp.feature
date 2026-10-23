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
