Feature: Time on, hidden, or off

  The time is recorded and shown, recorded and not shown, or neither. Off,
  the focus stays without a timer; turned back on, the timer starts then,
  nothing caught up. Nothing already recorded is ever taken away.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And I added a task titled "Write the spec"

  Rule: The time says one of three words, and what was written before still reads

    @time-config-show-writes-the-word
    Scenario Outline: what the pages say of the time is shown as its word
      Given the pages say time = <written>
      When I run `cairn config show`
      Then it says "time" is "<word>"

      Examples:
        | written  | word   |
        | true     | on     |
        | false    | hidden |
        | "on"     | on     |
        | "hidden" | hidden |
        | "off"    | off    |

    @time-on-by-default
    Scenario: nothing said of the time is on, as it always was
      When I run `cairn config show`
      Then it says "time" is "on"
      When I run `cairn --help`
      Then the help offers "timesheet"

    @time-a-word-it-does-not-know-is-refused
    Scenario: a word the time does not take is refused, naming the three it does
      Given the pages say time = "paused"
      When I run `cairn task list`
      Then it fails
      And it says on standard error "pages.time"
      And it says on standard error that the time is "on", "hidden" or "off"

  Rule: Hidden, the time is recorded and not shown

    @time-hidden-records-what-it-does-not-show
    Scenario: the time spent while hidden is there once it is shown again
      Given the pages say time = "hidden"
      And it is 09:00 on 2026-09-29
      When I run `cairn task focus 1`
      And it is 10:30 on 2026-09-29
      And I run `cairn task unfocus 1`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then it fails
      When the pages say time = "on"
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h30"

    @time-not-shown-says-no-time-spent
    Scenario Outline: hidden or off, neither a task nor its objective says the time spent on it
      Given I added an objective titled "Ship the spec"
      And I edited task 1 with the objective "1"
      And I edited task 1 with the time spent "2h"
      And the pages say time = "<time>"
      When I run `cairn task show 1`
      Then the answer does not say "time spent"
      When I run `cairn task show 1 --json`
      Then the answer does not say "2h"
      When I run `cairn objective show 1`
      Then the answer does not say "time spent"
      When I run `cairn objective list`
      Then the answer does not say "2h"
      When the pages say time = "on"
      And I run `cairn task show 1`
      Then the answer says "2h"

      Examples:
        | time   |
        | hidden |
        | off    |

  Rule: Off, the focus stays and no time is recorded

    @time-off-records-nothing
    Scenario: a task in focus while the time is off counts no time
      Given the pages say time = "off"
      And it is 09:00 on 2026-09-29
      When I run `cairn task focus 1`
      Then `cairn task focus` says "Write the spec"
      And `cairn trace show 1` says "without a timer"
      When it is 11:00 on 2026-09-29
      And I run `cairn task unfocus 1`
      And the pages say time = "on"
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" says nothing of task 1

    @time-on-from-off-starts-the-timer-then
    Scenario: time turned on from off starts the timer then, catching nothing up
      Given the pages say time = "off"
      And it is 09:00 on 2026-09-29
      When I run `cairn task focus 1`
      And the pages say time = "on"
      And it is 10:00 on 2026-09-29
      And I run `cairn task list`
      And it is 10:30 on 2026-09-29
      And I run `cairn task unfocus 1`
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "30m"
      And `cairn trace show 1` says "timer started"

    @time-off-keeps-what-was-recorded
    Scenario: time recorded while on stays once the time is off
      Given it is 09:00 on 2026-09-29
      When I run `cairn task focus 1`
      And the pages say time = "off"
      And it is 10:00 on 2026-09-29
      And I run `cairn task list`
      And it is 12:00 on 2026-09-29
      And I run `cairn task unfocus 1`
      And the pages say time = "on"
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" gives task 1 "1h"
      And `cairn trace show 1` says "timer stopped"

    @time-off-in-the-store-settings
    Scenario: the store's settings turn the time off while the global configuration has it on
      Given the pages say time = "on"
      And the settings beside the store say time = "off"
      And it is 09:00 on 2026-09-29
      When I run `cairn task focus 1`
      And it is 10:00 on 2026-09-29
      And I run `cairn task unfocus 1`
      And the settings beside the store say time = "on"
      And I run `cairn timesheet show --since 2026-09-29 --until 2026-09-29`
      Then the day "2026-09-29" says nothing of task 1

    @time-off-asks-no-renewal
    Scenario: no focus session asks to be renewed while the time is off
      Given the pages say time = "off", and focus sessions last "1h"
      When I run `cairn task focus 1`
      And it is 2 hours later
      And I run `cairn task focus --json`
      Then the focus session of task 1 asks nothing
