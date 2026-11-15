Feature: A series, the tasks that come back

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And recurrence is on for this store
    And today is "2026-11-09"

  Rule: A rule is written in plain words, and read back as a sentence with the next dates

    @series-grammar-after-done
    Scenario: a rule can count from the day the last one is done
      When I run `cairn task new "Water the plants" --every "every 3 days after done"`
      Then it succeeds
      And the answer says "every 3 days, after the last is done"
      And the answer says "then     3 days after it is done"

  Rule: One occurrence is open at a time, and the next is made when it is done or skipped

    @series-made-on-its-day-and-never-ahead
    Scenario: with nothing open, cairn makes the next on its day, and not before
      Given I ran `cairn task new Review --every "every friday"`
      And recurrence is off for this store
      And I ran `cairn task edit 1 --state done`
      And recurrence is on for this store
      When I run `cairn task list --series 1`
      Then `cairn task list --series 1 --with-archived` lists 1 task
      When today is "2026-11-20"
      And I run `cairn task list`
      Then `cairn task list --series 1` lists 2 tasks
      And `cairn task show 2` says "deadline     2026-11-20"
      And `cairn series show 1` says "0 missed"

  Rule: A series is stopped, resumed, changed and deleted on its own

    @series-edit-the-column-occurrences-arrive-in
    Scenario: a series chooses the column its occurrences arrive in
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn series edit 1 --state current`
      Then it succeeds
      When I run `cairn task edit 1 --state done`
      Then `cairn task show 2` says "current"
      And `cairn series show 1` says "changed column"

    @series-edit-a-column-that-needs-a-reason-is-refused
    Scenario Outline: an occurrence arrives in backlog, todo or current, never in a column that needs a reason or in done
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn series edit 1 --state <column>`
      Then it fails

      Examples:
        | column   |
        | blocked  |
        | watching |
        | paused   |
        | done     |

    @series-task-edit-reaches-the-series
    Scenario: an occurrence changed with --series changes the series too
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn task edit 1 --title Daily --priority high --series`
      Then it succeeds
      When I run `cairn task edit 1 --state done`
      Then `cairn task show 2` says "Daily"
      And `cairn task show 2` says "priority     high"

  Rule: A listing keeps the occurrences when asked

    @series-list-repeats
    Scenario: a listing keeps only the occurrences, or those of one series
      Given I added a task titled "Write the spec"
      And I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task new Review --every "every friday"`
      When I run `cairn task list --series`
      Then it succeeds
      And the answer says "↻2 every week on Friday"
      And the answer does not say "Write the spec"
      And `cairn task list --series 2` lists 1 task

  Rule: Recurrence is off unless a store's settings turn it on, and never takes back a task

    @series-off-offers-nothing
    Scenario: with recurrence off there is no rule to give and no series to ask for
      Given recurrence is off for this store
      When I run `cairn task new "Team stand-up" --every "every weekday"`
      Then it fails
      And it says on standard error "unknown flag: --every"
      When I run `cairn series list`
      Then it fails
