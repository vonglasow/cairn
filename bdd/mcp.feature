Feature: Answering a machine over MCP

  The same cairn, asked by an agent instead of by a person. It parses a call,
  hands it to the engine and answers — it decides nothing the command line does
  not already decide, and the rules live where they have always lived.

  What an agent may do is settled when the server starts, not per call:
  `--access read` looks and never writes, `--access write` is the default and
  also adds and changes, `--access full` also does what loses work.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: The server is started like any other command

    `cairn mcp serve`, because everything else here reads
    `cairn <thing> <verb>` and a server is no reason to break that. It speaks
    over standard input and output, since a client starts it itself as a child
    process — and because no open port is the right default for something one
    person runs on one machine.

    A client owns the process it started, and ends it by closing the pipe. That
    is how every session ends rather than a way one fails: once the asker has
    gone there is nobody left to answer, so a request still in flight is not
    work lost, and an exit code nobody is there to read is not worth failing
    over.

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

    A command opens the store when it needs it and closes it on the way out,
    and that is what makes "added by one, found by the next" a claim about the
    disk rather than about a cache. A server is only a command that is asked
    more often, so it does the same: opening costs a fraction of the work of
    the call it serves, and holding the file would make the server the one
    place where the rule stops being true.

    It also settles two things that would otherwise need settling. Nothing
    holds a write-ahead log open, so a backup still copies a whole store. And
    nothing holds the file while `backup_restore` or `migration_rollback`
    swaps it underneath.

    @mcp-a-backup-taken-through-it-holds-everything
    Scenario: a backup taken through the server holds what was just added
      Given a cairn server started with `--access full`
      When I call `task_add` with the title "Added just now"
      And I call `backup_add`
      And I call `backup_restore` with the number 1
      Then `cairn task list` lists 1 task

    @mcp-sees-what-the-terminal-did
    Scenario: what was done at the terminal is there on the next call
      Given a cairn server started with no access given
      And I asked it to add a task titled "From the agent"
      When I run `cairn task add "From the terminal"`
      And I call `task_list`
      Then it answers with 2 tasks

    @mcp-what-it-did-is-on-disk
    Scenario: what the server did is there for the terminal
      Given a cairn server started with no access given
      When I call `task_add` with the title "From the agent"
      Then `cairn task list` lists 1 task

  Rule: What the server offers is what it was allowed to do

    An agent cannot ask for a tool it was never shown, so a level is kept by
    leaving tools out rather than by turning them down. A refusal invites
    another attempt; an absence does not.

    @mcp-read-offers-reading
    Scenario: reading only
      Given a cairn server started with `--access read`
      When I ask it what it can do
      Then it offers `task_list`
      And it does not offer `task_add`
      And it does not offer `task_delete`

    @mcp-write-offers-changing
    Scenario: reading and changing, which is the default
      Given a cairn server started with no access given
      When I ask it what it can do
      Then it offers `task_list`
      And it offers `task_add`
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

    The level settles what an agent may reach; the annotations tell the client
    what it is about to do, so a person is asked before the work goes. One is
    ours to enforce, the other is theirs to honour — a client that ignores the
    hint still cannot reach a tool the level left out.

    A third thing is said, and it is true of all of them: cairn is a store on
    this machine and goes looking nowhere else. The protocol assumes the
    opposite until told otherwise, so saying nothing would claim a reach cairn
    does not have.

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

    The same work, and the same words for it: an agent names a state the way a
    person types it, and is turned down for the same reasons. Three things are
    said out loud rather than left to be worked out — what a tool takes, what
    handing it nothing means, and what handing it several things means. That
    last one is here because it was got wrong first: a list arrived as one
    string nobody read, so labels an agent gave were dropped without a word,
    and nothing on either side said so.

    @mcp-add-a-task
    Scenario: adding a task
      Given a cairn server started with no access given
      When I call `task_add` with the title "Write the spec"
      Then it answers without an error
      And `cairn task list` lists 1 task

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
    Scenario: what an agent changes says it was not the terminal
      Given I added a task titled "Write the spec"
      And a cairn server started with no access given
      When I call `task_edit` with the number 1 and the priority "high"
      Then the history of task 1 says the change came through the server

    @mcp-says-what-it-answers
    Scenario: every tool says what shape its answer takes
      Given a cairn server started with no access given
      When I ask it what it can do
      Then every tool describes the shape of its answer

    @mcp-every-tool-answers-an-object
    Scenario: every tool answers something a machine reads as one thing
      Given a cairn server started with no access given
      When I ask it what it can do
      Then every tool says it answers one object

    @mcp-says-what-one-thing-it-answers
    Scenario: a tool says which fields its one answer carries
      Given a cairn server started with no access given
      When I ask it what it can do
      Then `task_list` says its answer carries the tasks
      And `task_show` says its answer carries a title

    @mcp-what-it-answers-has-the-shape-it-said
    Scenario: what a tool answers is measured against what it said
      Given I added a task titled "Write the spec"
      And a cairn server started with no access given
      When I call `task_list`
      Then its answer holds to the shape the tool described

    @mcp-a-refusal-is-not-measured-against-that-shape
    Scenario: a refusal is not an answer of that shape
      Given I added a task titled "Write the spec"
      And a cairn server started with no access given
      When I call `task_show` with the number 99
      Then it answers that it failed
      And it carries nothing for a machine to measure

    @mcp-labels-arrive-whole
    Scenario: labels given to a tool arrive as labels
      Given a cairn server started with no access given
      When I call `task_add` with the title "From the agent" and the labels "spec" and "team"
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
      Given a cairn server started with no access given
      When I ask it what it can do
      Then `task_add` names every state cairn accepts
      And it says what each thing it takes is for

    @mcp-empties-a-field
    Scenario: handed nothing a field is emptied, left out it is untouched
      Given I added a task titled "Write the spec" with the deadline "2026-09-01"
      And a cairn server started with no access given
      When I call `task_edit` with the number 1 and the deadline as nothing
      Then the task has no deadline
      And it is still titled "Write the spec"

  Rule: A refusal is a refusal, not a broken tool

    A tool that reports an error means cairn could not do the work. Something
    cairn understood and turned down — an empty title, a state that is not one
    — is an answer, and the agent has to be able to read it as one.

    A store this cairn cannot use belongs with those, not with the broken ones.
    Nothing is broken: the store is simply not the one this cairn knows. An
    agent that reads why can put it right, and one told only that the call
    failed cannot.

    @mcp-refusal-is-answered
    Scenario: a refusal comes back as the reason it was refused
      Given a cairn server started with no access given
      When I call `task_add` with the title ""
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
