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
