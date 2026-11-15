Feature: Answering a machine over MCP

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: The server is started like any other command

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

  Rule: What the server offers is what it was allowed to do

    @mcp-read-offers-reading
    Scenario: reading only, which is the default
      Given a cairn server started with no access given
      When I ask it what it can do
      Then it offers `task_list`
      And it does not offer `task_new`
      And it does not offer `task_delete`

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

  Rule: A tool does what the command of the same name does

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

  Rule: An agent reads the views, and keeps, changes or deletes none

    @view-mcp-only-reads-the-views
    Scenario: an agent is offered the views to read, and nothing that changes one
      Given a cairn server started with `--access full`
      When I ask it what it can do
      Then it offers `view_list`
      And `view_list` is marked as read-only
      And it does not offer `view_new`
      And it does not offer `view_edit`
      And it does not offer `view_delete`

  Rule: An agent handles a series the way a person does, at the level it was given

    @series-mcp-offers-the-series
    Scenario: with recurrence on, an agent is offered the series at its level
      Given recurrence is on for this store
      And a cairn server started with `--access full`
      When I ask it what it can do
      Then it offers `series_list`
      And `series_list` is marked as read-only
      And it offers `series_show`
      And it offers `series_edit`
      And it offers `series_stop`
      And it offers `series_resume`
      And it offers `task_skip`
      And `series_delete` is marked as losing work

  Rule: A refusal is a refusal, not a broken tool

    @mcp-refusal-is-answered
    Scenario: a refusal comes back as the reason it was refused
      Given a cairn server started with `--access write`
      When I call `task_new` with the title ""
      Then it answers that it failed
      And the answer says a task needs a title

  Rule: An agent is told where its work on a task goes

    @mcp-greeting-says-where-the-work-goes
    Scenario: an agent is told as it greets the server
      When I start `cairn mcp serve` and greet it
      Then it tells the agent that what was done or decided on a task goes in its notes, never in its description

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

  Rule: A page turned off in the configuration offers no tool

    @mcp-a-page-off-offers-none-of-its-tools
    Scenario: the time page turned off takes its tools away, and leaves the rest
      Given a configuration that turns the page "time" off
      And a cairn server started with `--access full`
      When I ask it what it can do
      Then it does not offer `timesheet_show`
      And it does not offer `timesheet_add`
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
