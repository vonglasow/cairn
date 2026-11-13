Feature: The board

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a window of 100 by 44

  Rule: The work is laid out in four columns and a band

    @tui-four-columns-left-to-right
    Scenario: the columns are named in the order the work goes through
      When I open the board
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @tui-columns-stand-apart
    Scenario: a rule stands between each two columns
      When I open the board
      Then a rule stands between each two columns

    @tui-archived-are-off-the-board-until-asked-for
    Scenario: what was archived comes back when it is asked for
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      And I archived task 1
      When I open the board
      Then the board carries 1 task
      And the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      Then the board carries 2 tasks
      And "BACKLOG" carries "Migration Jenkins"
      When I press "shift+a"
      Then the board carries 1 task

    @tui-a-task-is-taken-off-the-board-from-the-cursor
    Scenario: a task is archived, and unarchived
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      When I open the board
      And I press "i"
      Then the cursor is on the task "Migration Jenkins"
      When I press "x"
      Then the board carries 1 task
      And the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      Then the board carries 2 tasks
      When I press "x"
      Then the board carries 2 tasks
      And "BACKLOG" carries "Migration Jenkins"

    @tui-a-step-is-not-on-the-board
    Scenario: a step does not stand in a column of its own
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
      And I added a task titled "Something else"
      When I open the board
      Then the board carries 2 tasks
      And "BACKLOG" carries "Ship it"
      And "BACKLOG" carries "Something else"
      And the board says nothing of "Write the spec"

    @tui-the-band-holds-what-is-not-moving
    Scenario: blocked, watching and paused share one band
      Given I added a task titled "Client's answer" with the state "blocked"
      And I added a task titled "Handed over" with the state "watching"
      And I added a task titled "Not now" with the state "paused"
      When I open the board
      Then the band is named "STALLED"
      And the band carries 3 tasks

    @tui-an-empty-column-stays
    Scenario: a column with nothing in it is still drawn
      Given I added a task titled "Write the spec"
      When I open the board
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @tui-a-task-sits-in-the-column-of-its-state
    Scenario: a task is drawn under the column it stands in
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the board
      Then "TODO" carries "Write the spec"

  Rule: A window too narrow is told so, rather than drawn badly

    @tui-too-narrow-says-what-it-needs
    Scenario: a window under sixty says how many it needs
      Given a window of 47 by 30
      When I open the board
      Then it says it needs 60 columns
      And it says the window is 47

    @tui-sixty-is-wide-enough
    Scenario: sixty columns still draws the board
      Given a window of 60 by 26
      When I open the board
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

  Rule: Normal moves between columns, insert between tasks

    @tui-the-board-opens-in-normal
    Scenario: the board opens on a column
      When I open the board
      Then the mode is "NORMAL"
      And the cursor is on the column "BACKLOG"

    @tui-normal-moves-between-columns
    Scenario: l moves to the next column
      When I open the board
      And I press "l"
      Then the cursor is on the column "TODO"

    @tui-normal-stops-at-the-last-column
    Scenario: there is nothing to the right of done
      When I open the board
      And I press "l", "l", "l", "l"
      Then the cursor is on the column "DONE"

    @tui-insert-takes-hold-of-a-task
    Scenario: i puts the cursor on a task
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i"
      Then the mode is "INSERT"
      And the cursor is on the task "Write the spec"

    @tui-insert-moves-between-tasks
    Scenario: j moves down the column
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I open the board
      And I press "i", "j"
      Then the cursor is on the task "Second"

    @tui-escape-goes-back-to-normal
    Scenario: escape lets go of the task
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "esc"
      Then the mode is "NORMAL"

    @tui-insert-does-nothing-on-an-empty-column
    Scenario: there is no task to take hold of
      When I open the board
      And I press "i"
      Then the mode is "NORMAL"

  Rule: Walking sideways keeps the level, and the ends are one key away

    @tui-sideways-keeps-the-level
    Scenario: changing column stays at the same level in it
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Also first" with the state "current"
      And I added a task titled "Also second" with the state "current"
      When I open the board
      And I press "l", "i", "j"
      Then the cursor is on the task "Second"
      When I press "l"
      Then the cursor is on the task "Also second"

    @tui-sideways-stops-at-a-shorter-columns-end
    Scenario: a column with fewer in it stops at its last
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Only one" with the state "current"
      When I open the board
      And I press "l", "i", "j", "l"
      Then the cursor is on the task "Only one"

    @tui-G-and-gg-reach-the-ends
    Scenario: G goes to the last of a column and gg back to the first
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Third" with the state "todo"
      When I open the board
      And I press "l", "i", "shift+g"
      Then the cursor is on the task "Third"
      When I press "g", "g"
      Then the cursor is on the task "First"

  Rule: The band is walked the way it is drawn

    @tui-the-band-is-a-grid
    Scenario: right walks along the band's row, down goes a row lower
      Given I added a task titled "First" with the state "blocked"
      And I added a task titled "Second" with the state "blocked"
      And I added a task titled "Third" with the state "blocked"
      And I added a task titled "Fourth" with the state "blocked"
      When I open the board
      And I press "l", "j", "i"
      Then the cursor is on the task "First"
      When I press "l"
      Then the cursor is on the task "Second"
      When I press "h", "j"
      Then the cursor is on the task "Fourth"

  Rule: A page turned off in the configuration is not drawn

    @tui-a-page-off-has-no-tab
    Scenario: the time page turned off has no tab, and tab walks the others
      Given the pages turned off are "time"
      When I open the board
      Then the board says "TASKS"
      And the board does not say "time"
      When I press "tab", "tab"
      Then the board says "TASKS"

    @tui-a-page-off-takes-its-commands
    Scenario: a command of a page turned off is not run from the command line
      Given the pages turned off are "time"
      When I open the board
      And I press ":"
      And I type "timesheet show"
      And I press "enter"
      Then the board says "not a command"

    @tui-every-page-off-is-an-empty-screen
    Scenario: every page turned off opens a board with nothing in it
      Given the pages turned off are "tasks, objectives, time"
      When I open the board
      Then the screen holds nothing
      When I press "a"
      Then the screen holds nothing

  Rule: Moving a task changes where it stands

    @tui-moving-a-task-across-changes-its-column
    Scenario: pushing a task right puts it in the next column
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "shift+l"
      Then task 1 is in the column "todo"

    @tui-done-asks-about-open-steps
    Scenario: pushing a task into done with steps still open asks to tick them too
      Given I added a task titled "Ship it" with the state "current"
      And I added the step "Write the spec" to task 1
      When I open the board
      And I press "l", "l", "i", "shift+l"
      And the board looks again
      Then the board says "put task 1 in done, ticking its steps?"
      When I press "enter"
      And the board looks again
      Then task 1 is in the column "done"
      And every step of task 1 is done

    @tui-done-asked-about-can-be-kept
    Scenario: saying no leaves the task and its steps as they were
      Given I added a task titled "Ship it" with the state "current"
      And I added the step "Write the spec" to task 1
      When I open the board
      And I press "l", "l", "i", "shift+l"
      And the board looks again
      And I press "esc"
      Then task 1 is in the column "current"

    @tui-done-asked-about-keeps-the-cursor
    Scenario: saying no puts the cursor back on the task, where it still is
      Given I added a task titled "Ship it" with the state "current"
      And I added the step "Write the spec" to task 1
      When I open the board
      And I press "l", "l", "i", "shift+l"
      And the board looks again
      And I press "esc"
      And the board looks again
      Then the cursor is on the column "current"
      And the cursor is on the task "Ship it"

    @tui-moving-a-task-down-a-place
    Scenario: pushing a task down puts it one place lower, and leaves its column
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      When I open the board
      And I press "l", "i", "shift+j"
      Then task 1 is in the column "todo"
      And "TODO" reads "Second", "First"

    @tui-moving-the-last-task-down-asks-which-kind
    Scenario: pushing the last task down asks which kind of stopped it is
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the board
      And I press "l", "i", "shift+j"
      Then it asks which of "blocked", "watching", "paused"
      And task 1 is in the column "todo"

    @tui-choosing-in-the-band-sets-the-state
    Scenario: choosing in the popup is what changes the task
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the board
      And I press "l", "i", "shift+j"
      And I choose "watching"
      Then task 1 is in the column "watching"

    @tui-out-of-the-band-returns-where-the-cursor-came-from
    Scenario: coming back out returns the task to the column it was pushed from
      Given I added a task titled "Write the spec" with the state "current"
      When I open the board
      And I press "l", "l", "i", "shift+j"
      And I choose "blocked"
      And I press "shift+k"
      Then task 1 is in the column "current"

    @tui-moving-a-task-down-among-its-kind-in-the-band
    Scenario: pushing a task down in the band puts it one place lower among the tasks of its kind
      Given I added a task titled "First" with the state "blocked"
      And I added a task titled "Second" with the state "blocked"
      And I added a task titled "Elsewhere" with the state "watching"
      When I open the board
      And I press "l", "j", "i", "shift+j"
      Then task 1 is in the column "blocked"
      And the band reads "Second", "First", "Elsewhere"

    @tui-moving-a-task-up-among-its-kind-in-the-band
    Scenario: pushing a task up in the band puts it one place higher among the tasks of its kind
      Given I added a task titled "Elsewhere" with the state "blocked"
      And I added a task titled "First" with the state "paused"
      And I added a task titled "Second" with the state "paused"
      When I open the board
      And I press "l", "j", "i", "l", "l", "shift+k"
      Then task 3 is in the column "paused"
      And the band reads "Elsewhere", "Second", "First"

    @tui-the-last-of-its-kind-in-the-band-stays
    Scenario: pushing down the last task of its kind in the band moves nothing
      Given I added a task titled "First" with the state "blocked"
      And I added a task titled "Second" with the state "watching"
      When I open the board
      And I press "l", "j", "i", "shift+j"
      Then task 1 is in the column "blocked"
      And the band reads "First", "Second"

    @tui-out-of-the-band-leftmost-goes-to-the-backlog
    Scenario: pushing left out of the band reaches the backlog
      Given I added a task titled "Write the spec" with the state "blocked"
      When I open the board
      And I press "l", "j", "i", "shift+h"
      Then task 1 is in the column "backlog"

    @tui-navigating-into-the-band-asks-nothing
    Scenario: looking at what is stalled does not ask a question
      Given I added a task titled "Client's answer" with the state "blocked"
      When I open the board
      And I press "l", "j"
      Then it asks nothing

  Rule: The bar shows only what works here

    @tui-the-bar-follows-the-mode
    Scenario: the bar changes with the mode
      Given I added a task titled "Write the spec"
      When I open the board
      Then the bar does not offer "move"
      When I press "i"
      Then the bar offers "move"

  Rule: Help answers in every mode

    @tui-help-answers-in-every-mode
    Scenario Outline: the question mark answers wherever it is asked
      Given I added a task titled "Write the spec"
      When I open the board
      And I press <keys>
      And I press "?"
      Then the help is shown

      Examples:
        | keys       |
        | "l"        |
        | "i"        |
        | "i", "e"   |
        | "/"        |

    @tui-leaving-help-returns-where-it-was-asked
    Scenario: closing help puts the board back as it was
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "?", "esc"
      Then the mode is "INSERT"

  Rule: Searching narrows the board in place

    @tui-searching-narrows-in-place
    Scenario: only what matches is drawn, in the columns it sits in
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "spec"
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"
      And "TODO" carries "Write the spec"
      And the board carries 1 task

    @tui-searching-by-number
    Scenario: a number is the way back when it is what somebody has
      Given I added a task titled "Write the spec"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "2"
      Then the board carries 1 task
      And the board says "Buy a lamp"

    @tui-searching-reaches-what-the-command-line-reaches
    Scenario: a search reaches every field the command line looks in
      Given I added a task titled "Write the spec" with the estimate "90m"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "1h30m"
      Then the board carries 1 task

    @tui-searching-raises-the-task-a-step-belongs-to
    Scenario: a word found only in a step answers the task it belongs to
      Given I added a task titled "Ship it"
      And I added the step "Write the pipeline" to task 1
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "pipeline"
      Then the board carries 1 task
      And the board says "Ship it"

    @tui-searching-raises-each-task-once
    Scenario: two steps of one task answer it once
      Given I added a task titled "Ship it"
      And I added the step "Write the pipeline" to task 1
      And I added the step "Fix the pipeline" to task 1
      When I open the board
      And I press "/"
      And I type "pipeline"
      Then the board carries 1 task

    @tui-leaving-a-search-gives-the-board-back
    Scenario: escape ends the search
      Given I added a task titled "Write the spec"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "spec"
      And I press "esc"
      Then the board carries 2 tasks

    @tui-escape-widens-a-search-already-accepted
    Scenario: escape widens a board a search has already narrowed
      Given I added a task titled "Write the spec"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "spec"
      And I press "enter"
      Then the board carries 1 task
      When I press "esc"
      Then the board carries 2 tasks

    @tui-searching-by-label-and-by-column
    Scenario: a search reaches the labels and the column, as the engine's does
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Buy a lamp"
      And I edited task 1 with the labels "backend" and "sprint-172"
      When I open the board
      And I press "/"
      And I type "backend"
      Then the board carries 1 task
      And the board says "Write the spec"

    @tui-searching-keeps-only-what-carries-every-word
    Scenario: every word typed has to be carried, not one of them
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Write the report"
      When I open the board
      And I press "/"
      And I type "write todo"
      Then the board carries 1 task
      And the board says "Write the spec"

  Rule: A card says what its task carries, and no more

    @tui-a-card-carries-the-number-and-the-title
    Scenario: a card is its number and its title
      Given I added a task titled "Write the spec"
      When I open the board
      Then the card for task 1 says "1"
      And the card for task 1 says "Write the spec"

    @tui-a-cut-title-is-read-whole-in-insert
    Scenario: the card under the cursor gives its whole title when the card cuts it
      Given I added a task titled "A title far longer than any column could ever hope to show in full, even over two lines of a card"
      When I open the board
      And I press "i"
      Then the board says "A title far longer than any column could ever hope to show in full, even over two lines of a card"

    @tui-a-title-that-fits-is-not-said-twice
    Scenario: a title the card holds whole is not written again
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i"
      Then the board says "Write the spec" once

    @tui-a-card-marks-a-description-without-showing-it
    Scenario: a description is marked, never drawn
      Given I added a task titled "Write the spec" with the description "The client wants two columns"
      When I open the board
      Then the card for task 1 is marked as described
      And the board does not say "The client wants two columns"

    @tui-a-card-says-what-blocks-it
    Scenario: a stalled card says what is in its way
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I open the board
      Then the card for task 2 says "1"

    @tui-a-card-says-when-nothing-blocks-it-any-more
    Scenario: a stalled card says when every task it waits on is done
      Given I added a task titled "Write the spec" with the state "done"
      And I added a task titled "Ship it"
      And I edited task 2 with the blocker "1"
      And I edited task 2 with the state "blocked"
      When I open the board
      Then the card for task 2 says "nothing blocks it"

    @tui-a-card-says-what-it-blocks
    Scenario: a card says how many tasks it is in the way of
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I added a task titled "Announce it"
      And I edited task 2 with the blocker "1"
      And I edited task 3 with the blocker "1"
      When I open the board
      Then the card for task 1 says "⊳2"

    @tui-an-archived-card-is-marked
    Scenario: what was archived is told apart from what is on the board
      Given I added a task titled "Write the spec"
      And I archived task 1
      When I open the board
      And I press "shift+a"
      Then the card for task 1 is marked as archived

    @tui-the-tasks-in-focus-come-first-in-current
    Scenario: the tasks in focus are drawn first in current, in the order they were focused
      Given I added a task titled "Answer the mail" with the state "current"
      And I added a task titled "Fix the gate" with the state "current"
      And I added a task titled "Call the plumber" with the state "current"
      And I put task 3 in focus
      And I put task 2 in focus
      When I open the board
      Then the board says "Call the plumber" before "Fix the gate"
      And the board says "Fix the gate" before "Answer the mail"

    @tui-the-task-in-focus-is-told-apart
    Scenario: the task in hand is drawn unlike any other
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      When I open the board
      Then the card for task 1 is drawn as the one in focus

  Rule: The command prompt runs what the command line runs

    @tui-the-command-prompt-runs-an-operation
    Scenario: a command typed at the prompt is the command line's own
      Given I added a task titled "Write the spec"
      When I open the board
      And I press ":"
      And I type "task edit 1 --state current"
      And I press "enter"
      And the board looks again
      Then task 1 is in the column "current"

    @tui-the-command-prompt-says-what-it-did-not-understand
    Scenario: a command nobody could read is said rather than swallowed
      When I open the board
      And I press ":"
      And I type "zz"
      And I press "enter"
      Then the board says "zz"

  Rule: Two tabs, and an objective opens a board of its own

    @tui-the-tab-key-changes-tab
    Scenario: tab moves to the objectives
      When I open the board
      And I press "tab"
      Then the tab is "OBJECTIVES"

    @tui-an-objective-says-how-far-along-it-is
    Scenario: an objective carries the count of what serves it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Ship it" with the state "done" serving objective 1
      When I open the board
      And I press "tab"
      Then the card for objective 1 says "1/2"

    @tui-an-objective-card-counts-steps-apart
    Scenario: an objective counts the steps of its tasks apart from its tasks
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added the step "Read the ADR" to task 1, done
      And I added the step "Write it" to task 1
      When I open the board
      And I press "tab"
      Then the card for objective 1 says "0/1 · 1/2 steps"

    @tui-an-objective-card-says-the-time-spent
    Scenario: an objective's card says the time spent on the tasks serving it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Ship it" serving objective 1
      And I spent "1h" on task 1 on "2026-09-29"
      And I spent "30m" on task 2 on "2026-09-30"
      When I open the board
      And I press "tab"
      Then the card for objective 1 says "1h30m"

    @tui-an-objective-opens-its-own-board
    Scenario: opening an objective shows only what serves it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "tab"
      And I press "enter"
      Then the board carries 1 task
      And the board says "Ship the desktop board"

  Rule: Objectives are a tree, and an objective's board walks its sub-objectives

    @tui-objective-sub-objectives-sit-under-their-parent
    Scenario: the objectives tab draws each sub-objective under its parent
      Given I added an objective titled "Home automation"
      And I added an objective titled "Tidy the workshop"
      And I added an objective titled "Firmware" under objective 1
      When I open the board
      And I press "tab"
      Then the objectives read "Home automation", "Firmware", "Tidy the workshop"
      And the board says "└─"

    @tui-objective-a-parent-counts-each-task-once
    Scenario: a parent counts the tasks of its sub-objectives, each once
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Pin map" serving objectives 2 and 3
      And I added a task titled "OTA updates" with the state "done" serving objective 3
      When I open the board
      And I press "tab"
      Then the card for objective 1 says "1/2"

    @tui-objective-a-row-says-what-it-waits-on
    Scenario: an objective's row says what it waits on
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      When I open the board
      And I press "tab"
      Then the card for objective 2 says "waits on Firmware"

    @tui-objective-the-band-carries-the-sub-objectives
    Scenario: an objective's board carries its sub-objectives on its band, each counted
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "OTA updates" serving objective 3
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "tab", "enter"
      Then the board carries 3 tasks
      And the board says "all 3"
      And the board says "Enclosure 1"
      And the board says "itself 1"

    @tui-objective-brackets-walk-the-sub-objectives
    Scenario: ] and [ narrow an objective's board to one sub-objective, or to the objective itself
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "OTA updates" serving objective 3
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "tab", "enter", "]"
      Then the board carries 1 task
      And "BACKLOG" carries "Vent slots"
      When I press "]", "]"
      Then the board carries 1 task
      And "BACKLOG" carries "Buy the relay"
      When I press "]"
      Then the board carries 3 tasks
      When I press "["
      Then the board carries 1 task
      And "BACKLOG" carries "Buy the relay"

    @tui-objective-a-card-names-what-it-serves
    Scenario: a card names the objective it serves
      Given I added an objective titled "Firmware"
      And I added a task titled "OTA updates" serving objective 1
      When I open the board
      Then the card for task 1 says "◇ Firmware"

    @tui-objective-t-shows-its-tasks
    Scenario: t shows on the tasks board the tasks of an objective and of its sub-objectives
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "tab", "t"
      Then the tab is "TASKS"
      And the board carries 1 task
      And the board says "filtered by objective #1 Home automation, with its sub-objectives"

    @tui-objective-brackets-walk-its-tasks-shown
    Scenario: ] and [ narrow the tasks an objective shows to one sub-objective, or to the objective itself
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "OTA updates" serving objective 3
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "tab", "t", "]"
      Then the tab is "TASKS"
      And the board carries 1 task
      And "BACKLOG" carries "Vent slots"
      And the board says "filtered by objective #2 Enclosure"
      When I press "]", "]"
      Then the board carries 1 task
      And the board says "filtered by objective #1 Home automation, itself"
      When I press "]"
      Then the board carries 3 tasks
      When I press "["
      Then the board carries 1 task
      And "BACKLOG" carries "Buy the relay"

    @tui-objective-shift-n-adds-a-sub-objective
    Scenario: N adds a sub-objective under the objective the cursor is on
      Given I added an objective titled "Home automation"
      When I open the board
      And I press "tab", "shift+n"
      And I set the field "title" to "Firmware"
      And I save the form
      Then objective 2 sits under objective 1

    @tui-objective-shift-n-on-a-sub-objective-adds-beside-it
    Scenario: N on a sub-objective adds another under the same parent
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      When I open the board
      And I press "tab", "j", "shift+n"
      And I set the field "title" to "Firmware"
      And I save the form
      Then objective 3 sits under objective 1

    @tui-objective-z-folds-a-parent
    Scenario: z folds a parent's sub-objectives away, and unfolds them again
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      When I open the board
      And I press "tab", "z"
      Then the board does not say "Enclosure"
      And the board says "+2"
      When I press "z"
      Then the board says "Enclosure"

    @tui-objective-shift-jk-reorders
    Scenario: shift with j or k moves the objective under the cursor among its siblings
      Given I added an objective titled "Home automation"
      And I added an objective titled "Tidy the workshop"
      When I open the board
      And I press "tab", "J"
      Then the board says "Tidy the workshop" before "Home automation"
      When I press "K"
      Then the board says "Home automation" before "Tidy the workshop"
      And the board says "JK order"

    @tui-objective-z-on-a-sub-objective-folds-its-parent
    Scenario: z on a sub-objective folds its parent, and the cursor goes up to it
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      When I open the board
      And I press "tab", "j", "z", "e"
      And I set the field "title" to "Domotics"
      And I save the form
      And I press "esc"
      Then the board says "Domotics"
      And the board does not say "Home automation"
      And the board does not say "Enclosure"

    @tui-objective-itself-leaves-its-sub-objectives-out
    Scenario: the objective filter narrows to a parent itself, leaving its sub-objectives out
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "shift+f", "l", "l", "l", "j", "j", "esc"
      Then the board carries 1 task
      And the board says "filtered by objective #1 Home automation, itself"
      And "BACKLOG" carries "Buy the relay"

    @tui-objective-a-blocker-outside-is-a-reference
    Scenario: a task outside an objective that blocks one inside it is shown, and not counted
      Given I added an objective titled "Firmware"
      And I added a task titled "Order the relay"
      And I added a task titled "Wire the relay" serving objective 1
      And I edited task 2 with the blocker "1"
      When I open the board
      And I press "tab", "enter"
      Then the board carries 1 task
      And the board says "Order the relay"

  Rule: An objective's form says where it sits, and a task's what it counts in

    @tui-objective-the-form-puts-it-under-a-parent
    Scenario: the objective's form offers the objectives of level one as its parent
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware"
      And I added an objective titled "Enclosure" under objective 1
      When I open the board
      And I press "tab", "shift+g", "e"
      Then the board says "1 Home automation"
      And the board does not say "3 Enclosure"
      When I set the field "parent" to "1"
      And I save the form
      Then objective 2 sits under objective 1

    @tui-objective-the-form-offers-labels-for-objectives
    Scenario: the objective's form offers only the labels meant for objectives
      Given I made the label "area" for objectives
      And I made the label "backend" for tasks
      And I added an objective titled "Home automation"
      When I open the board
      And I press "tab", "e"
      Then the board says "area"
      And the board does not say "backend"

    @tui-objective-the-form-says-what-it-waits-on
    Scenario: what an objective waits on is written in its form
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      When I open the board
      And I press "tab", "j", "e"
      And I set the field "waits on" to "1"
      And I save the form
      Then objective 2 now waits on objective 1

    @tui-objective-its-board-says-what-it-waits-on
    Scenario: an objective's board says what it waits on
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      When I open the board
      And I press "tab", "j", "enter"
      Then the board says "waits on #1 Firmware"

    @tui-objective-the-task-form-draws-the-tree
    Scenario: a task's form draws the objectives as a tree, a parent counted through what is ticked below it
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Pin map" serving objectives 2 and 3
      When I open the board
      And I press "i", "e"
      Then the board says "counted through 2 below"

    @tui-objective-the-task-form-refuses-a-double-count
    Scenario: a task put on a parent and on one of its sub-objectives is refused at the foot of the form
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Pin map" serving objective 2
      When I open the board
      And I press "i", "e"
      And I set the field "objectives" to "1 2"
      And I save the form
      Then the board says "is already counted in #1 Home automation through #2 Enclosure"
      And the mode is "EDIT"

  Rule: Deleting an objective asks what goes with it, and decides nothing

    @tui-objective-deleting-a-parent-asks-what-goes-with-it
    Scenario: deleting an objective holding sub-objectives and tasks asks about each
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "OTA updates" serving objective 3
      When I open the board
      And I press "tab", "shift+d"
      Then the board says "what goes with objective 1?"
      When I press "enter"
      Then the board says "choose for each first"
      And objective 1 is still there
      When I press "l", "j", "l", "l", "j", "l", "enter"
      Then objective 1 is gone
      And objective 2 stands on its own
      And objective 3 is gone
      And task 1 serves nothing

    @tui-objective-a-delete-asked-about-is-kept-on-escape
    Scenario: escape keeps an objective whose deletion was asked about
      Given I added an objective titled "Home automation"
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "tab", "shift+d", "l", "esc"
      Then objective 1 is still there
      And task 1 serves objective 1

  Rule: Work started early is warned about

    @tui-objective-work-started-early-is-warned-about
    Scenario: putting a task in focus says when its objective waits on one not reached
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      And I added a task titled "OTA updates" serving objective 2
      When I open the board
      And I press "i", "f"
      Then the board says "#2 Features waits on #1 Firmware, not reached"

  Rule: Nothing is written until it is saved

    @tui-escape-closes-the-form
    Scenario: escape closes the form and writes nothing
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I press "esc"
      Then the mode is "INSERT"
      And task 1 is titled "Write the spec"

    @tui-a-form-cancelled-changes-nothing
    Scenario: leaving the form throws the change away
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I set the field "title" to "Something else"
      And I press "esc"
      Then task 1 is titled "Write the spec"

    @tui-a-form-saved-changes-the-task
    Scenario: saving is what writes it
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I set the field "title" to "Something else"
      And I save the form
      Then task 1 is titled "Something else"

    @tui-adding-a-task-lands-in-the-column-under-the-cursor
    Scenario: a task added takes the column the cursor was on
      When I open the board
      And I press "l"
      And I press "a"
      And I set the field "title" to "Write the spec"
      And I save the form
      Then task 1 is in the column "todo"

  Rule: Two tasks are linked from the form, and a card opens a task linked to it

    @tui-the-form-links-a-task
    Scenario: a link typed in the form links the two tasks
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the board
      And I press "i", "j", "e"
      And I set the field "links" to "1"
      And I save the form
      Then task 2 is linked to task 1
      And task 1 is linked to task 2

    @tui-o-opens-a-linked-task
    Scenario: o on a card opens a new task linked to it, in todo
      Given I added a task titled "Ship it" with the state "done"
      When I open the board
      And I press "l", "l", "l", "i", "o"
      And I set the field "title" to "It does not hold"
      And I save the form
      Then task 2 is in the column "todo"
      And task 2 is linked to task 1
      And task 1 is linked to task 2

  Rule: A form is filled in the same way as the window's

    @tui-enter-writes-a-line-in-the-description
    Scenario: enter writes a line in the description rather than saving
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e", "tab"
      And I type "Deux colonnes"
      And I press "enter"
      And I type "et une bande"
      Then the mode is "EDIT"
      And task 1 is titled "Write the spec"

    @tui-ctrl-s-saves-every-field
    Scenario: ^S saves the form, and every field typed into is kept
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I set the field "labels" to "backend"
      And I press "ctrl+s"
      And the board looks again
      Then the card for task 1 says "backend"

  Rule: Losing work is asked about first

    @tui-deleting-asks-first
    Scenario: deleting asks before it takes the task away
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "D"
      Then the board says "delete task 1?"
      And the board carries 1 task
      When I press "esc"
      Then the board carries 1 task
      When I press "D", "enter"
      And the board looks again
      Then the board carries 0 tasks

    @tui-d-deletes-nothing
    Scenario: one letter destroys nothing: d no longer deletes, D asks first
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "d"
      Then the board carries 1 task
      And the board says nothing of "delete task 1?"

    @tui-m-moves-by-a-columns-letter
    Scenario: m then a column's first letter moves the chosen task there
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the board
      And I press "l", "i", "m", "d"
      And the board looks again
      Then "DONE" carries "Write the spec"

    @tui-a-filled-form-asks-before-it-is-thrown-away
    Scenario: a form somebody typed into is asked about before it goes
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I set the field "title" to "Something else"
      And I press "esc"
      Then the board says "throw away what you typed?"
      When I press "esc"
      Then the mode is "EDIT"

    @tui-an-untouched-form-asks-nothing
    Scenario: a form nobody typed into closes without a question
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "e"
      And I press "esc"
      Then the board does not say "throw away what you typed?"
      And the mode is "INSERT"

    @tui-a-whole-column-goes-off-the-board-at-once
    Scenario: the column the cursor is on is archived in one gesture
      Given I added a task titled "Write the spec" with the state "done"
      And I added a task titled "Ship it" with the state "done"
      When I open the board
      And I press "l", "l", "l"
      And I press "shift+x"
      Then the board says "archive every task in DONE?"
      When I press "enter"
      And the board looks again
      Then the board carries 0 tasks

  Rule: The detail shows what a card cannot

    @tui-the-detail-shows-the-description-and-the-trace
    Scenario: opening a task shows what the card only marked
      Given I added a task titled "Write the spec" with the description "The client wants two columns"
      When I open the board
      And I press "i", "enter"
      Then the board says "The client wants two columns"
      And the board says "created"

    @tui-the-steps-left-come-first
    Scenario: what is left comes first, ticked or not, and counted
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1, done
      And I added the step "Get the client's nod" to task 1
      When I open the board
      And I press "i", "enter"
      Then the board says "1 left · 1 done"
      And the board says "[ ] 2  Get the client's nod"
      And the board says "[x] 1  Draft the columns"

    @tui-a-step-is-deleted-once-agreed
    Scenario: the arrows choose a step of the open task, and a delete key takes it away once agreed
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      When I open the board
      And I press "i", "enter", "down"
      Then the board says "› [ ] 1  Draft"
      When I press "backspace"
      Then the board says "delete step 1 of task 1, “Draft”?"
      When I press "enter"
      And the board looks again
      Then the steps of task 1 read "Keep"
      And the board says "[ ] 1  Keep"

    @tui-a-step-is-kept-when-its-delete-is-cancelled
    Scenario: the forward delete asks too, and esc keeps the step
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      When I open the board
      And I press "i", "enter", "down", "down"
      Then the board says "› [ ] 2  Keep"
      When I press "delete"
      Then the board says "delete step 2 of task 1, “Keep”?"
      When I press "esc"
      And the board looks again
      Then the steps of task 1 read "Draft", "Keep"
      And the board says "› [ ] 2  Keep"

    @tui-the-detail-scrolls-its-description
    Scenario: a long description stays in its box, and ^D walks it to the end
      Given I added a task titled "Write the spec" with a description of 80 lines ending "LA FIN"
      When I open the board
      And I press "i", "enter"
      Then the board says nothing of "LA FIN"
      When I press "ctrl+d" 40 times
      Then the board says "LA FIN"

    @tui-the-detail-edits-what-it-is-showing
    Scenario: a task being read is edited without closing it first
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "enter"
      And I press "e"
      Then the mode is "EDIT"
      And the board says "Write the spec"

    @tui-the-detail-focuses-what-it-is-showing
    Scenario: the key the detail offers for focus focuses what it is showing
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "enter"
      And I press "f"
      And I press "esc"
      Then the card for task 1 is drawn as the one in focus

    @tui-an-open-task-says-where-its-context-lives
    Scenario: an open task says where what was written about it elsewhere lives
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "enter"
      Then the open task says where the context of task 1 lives

  Rule: The board follows the store while it stays open

    @tui-a-change-made-elsewhere-appears
    Scenario: a task added at the command line turns up
      Given I added a task titled "Write the spec"
      When I open the board
      And I added a task titled "Buy a lamp"
      And the board looks again
      Then the board carries 2 tasks

    @tui-the-cursor-holds-a-number-not-a-place
    Scenario: the cursor keeps its task when the board is redrawn
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I open the board
      And I press "i", "j"
      And I edited task 1 with the state "done"
      And the board looks again
      Then the cursor is on the task "Second"

    @tui-a-task-deleted-elsewhere-moves-the-cursor-on
    Scenario: the cursor falls to the next task in the same column
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I open the board
      And I press "i"
      And I deleted task 1
      And the board looks again
      Then the cursor is on the task "Second"

  Rule: The time spent and the backups are reached from the terminal board too

    @tui-the-time-tab
    Scenario: the third tab lays out the week's time, and the arrows walk the weeks
      Given I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-09-01"
      When I open the board
      And I press "tab"
      And I press "tab"
      Then the tab is "TIME"
      And the board says "Tue 01/09"
      And the board says "1h30m  #1 Write the spec"
      When I press "left"
      Then the board says "Mon 24/08"
      And the board does not say "Write the spec"
      When I press "t"
      Then the board says "Tue 01/09"

    @tui-a-week-without-its-weekend
    Scenario: with the weekends hidden, the time tab lays out a week as five days, Monday to Friday
      Given the weekends are hidden
      And I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-08-29"
      When I open the board
      And I press "tab"
      And I press "tab"
      And I press "left"
      Then the board says "Mon 24/08"
      And the board says "Fri 28/08"
      And the board does not say "Sat 29/08"
      And the board does not say "Sun 30/08"
      And the board does not say "Write the spec"

    @tui-a-command-says-its-answer
    Scenario: a command typed at the prompt says what it answered, until esc
      Given I added a task titled "Write the spec"
      When I open the board
      And I press ":"
      And I type "backup list"
      And I press "enter"
      Then the board says "there is no backup"
      When I press "esc"
      Then the board does not say "there is no backup"

    @tui-backups-from-the-board
    Scenario: b backs the store up and B lists the backups
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "b"
      Then the board says "#1"
      And the board says "holds 1 task(s)"
      When I press "esc"
      And I press "B"
      Then the board says "#1  made"

  Rule: An open task shows what is done to it elsewhere

    @tui-an-open-task-shows-a-step-added-elsewhere
    Scenario: a step added elsewhere appears on the task open on the board
      Given I added a task titled "Ship it"
      When I open the board
      And I press "i", "enter"
      And somebody else adds the step "Draft" to task 1
      And the board looks again
      Then the board says "[ ] 1  Draft"

    @tui-a-form-keeps-what-was-typed-when-its-task-changes-elsewhere
    Scenario: the form keeps what was typed, and takes the rest from what was done elsewhere
      Given I added a task titled "Ship it"
      When I open the board
      And I press "i", "e"
      And I set the field "title" to "Ship it now"
      And somebody else renames task 1 "Ship it today"
      And somebody else adds the step "Draft" to task 1
      And the board looks again
      Then the board says "Ship it now"
      And the board says "[ ] Draft"
      And the board does not say "Ship it today"

  Rule: The board is filtered by column, priority, label, objective and due day, and every filter is cleared at once

    Background:
      Given I added a task titled "Write the spec" in "todo" at the priority "high" carrying the label "backend"
      And I added a task titled "Review the plan" in "todo" at the priority "normal" carrying the label "backend"
      And I added a task titled "Ship it" in "current" at the priority "high" carrying the label "sprint"

    @tui-a-filter-narrows-the-board
    Scenario: a filter keeps only the cards it reaches, as it is changed
      When I open the board
      And I press "shift+f", "l", "j", "j", "j"
      Then the mode is "FILTER"
      And the board says "priority high"
      And the board says "Write the spec"
      And the board says "Ship it"
      And the board says nothing of "Review the plan"

    @tui-filters-narrow-together
    Scenario: two filters keep only what both reach
      When I open the board
      And I press "shift+f", "l", "j", "j", "j", "l", "j"
      Then the board says "Write the spec"
      And the board says nothing of "Ship it"

    @tui-every-filter-is-cleared-at-once
    Scenario: one key takes every filter away
      When I open the board
      And I press "shift+f", "l", "j", "j", "j", "l", "j", "backspace"
      Then the board says "Review the plan"
      And the board says "Ship it"

    @tui-the-board-says-what-filters-it
    Scenario: the board says what it is filtered by, once the filters are closed
      When I open the board
      And I press "shift+f", "l", "j", "j", "j", "esc"
      Then the mode is "NORMAL"
      And the board says "filtered by priority high"
      And the board says nothing of "Review the plan"

    @tui-a-filter-walks-back-to-its-last-value
    Scenario: a filter walked back from any comes to its last value
      When I open the board
      And I press "shift+f", "l", "k", "esc"
      Then the board says "filtered by priority critical"

  Rule: A store the board does not match is asked about, or said

    @tui-a-store-behind-is-offered-its-changes
    Scenario: a board opened on a store behind offers to apply what is waiting
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I open the board
      Then the board says "apply what this workspace is waiting for?"
      When I press "enter"
      And the board looks again
      Then task 1 is in the column "backlog"
      And a backup was set aside
      And no change is waiting

    @tui-a-store-behind-is-left-as-it-was-when-quit
    Scenario: leaving instead leaves the store as it was
      Given a store from an older cairn, holding a task titled "Write the spec"
      When I open the board
      And I press "esc"
      Then the board is leaving
      And no backup was set aside
      And 1 change is still waiting

    @tui-a-store-put-behind-while-open-is-noticed
    Scenario: a store put behind while the board is open is noticed
      Given I added a task titled "Write the spec"
      When I open the board
      And another cairn rolls the last change back
      And the board looks again
      Then the board says "apply what this workspace is waiting for?"

    @tui-a-store-from-a-newer-cairn-stops-the-board
    Scenario: a store a newer cairn changed stops the board
      Given I added a task titled "Write the spec"
      When I open the board
      And another cairn applies a change this cairn does not know
      And the board looks again
      Then the board says "this store comes from a newer cairn"
      When I press "enter"
      Then the board says "this store comes from a newer cairn"

  Rule: The board names the workspace it shows, and opens another in its place

    @tui-workspace-the-header-names-it
    Scenario: the header opens with the workspace the board shows
      Given the workspace "work", the default
      And the workspace "home"
      When I open the board
      Then the header opens with "● work"

    @tui-workspace-a-loose-store-shows-its-folder
    Scenario: a store no workspace names shows its folder instead
      Given the workspace "work", the default
      And a store no workspace names, in the folder "scratch"
      When I open the board
      Then the header opens with "◌ " and the folder "scratch"

    @tui-workspace-ctrl-o-lists-them
    Scenario: ctrl+o lists the workspaces, the one shown marked
      Given the workspace "work", the default
      And the workspace "home"
      When I open the board
      And I press "ctrl+o"
      Then the board says "OPEN A WORKSPACE"
      And the workspaces listed read "home", "work"
      And the workspace "work" is marked "here · default"
      And the workspace "home" is marked "1"

    @tui-workspace-choosing-replaces-the-board
    Scenario: choosing a workspace replaces the board in place, and says so
      Given the workspace "work", the default
      And the workspace "home"
      And I added a task titled "Rotate the certificates"
      And the workspace "home" holds a task titled "Fix the shed door"
      When I open the board
      And I press "ctrl+o"
      And I press "down", "up"
      And I press "enter"
      Then the header opens with "● home"
      And "BACKLOG" carries "Fix the shed door"
      And the board says nothing of "Rotate the certificates"
      And the board says "opened home in place of work"
      And the default workspace is still "work"

    @tui-workspace-chosen-by-its-place
    Scenario: a digit opens the workspace at that place
      Given the workspace "work", the default
      And the workspace "home"
      When I open the board
      And I press "ctrl+o", "1"
      Then the header opens with "● home"

    @tui-workspace-the-focus-left-is-still-held
    Scenario: a focus held in the workspace left is still held after switching
      Given the workspace "work", the default
      And the workspace "home"
      And I added a task titled "Choose the month" with the state "current"
      And I put task 1 in focus
      When I open the board
      And I press "ctrl+o", "1"
      Then the board says "work's focus on #1 is still held"
      And task 1 of the workspace "work" is still in focus

    @tui-workspace-esc-leaves-the-board-as-it-was
    Scenario: esc leaves the board as it was
      Given the workspace "work", the default
      And the workspace "home"
      And I added a task titled "Rotate the certificates"
      When I open the board
      And I press "ctrl+o", "down", "esc"
      Then the board does not say "OPEN A WORKSPACE"
      And the header opens with "● work"
      And "BACKLOG" carries "Rotate the certificates"

    @tui-workspace-the-list-narrows-as-one-types
    Scenario: the list narrows as one types
      Given the workspace "work", the default
      And the workspace "home"
      And the workspace "client-a"
      When I open the board
      And I press "ctrl+o"
      And I type "ho"
      Then the workspaces listed read "home"
      When I press "enter"
      Then the header opens with "● home"

    @tui-workspace-a-folder-gone-is-refused
    Scenario: a workspace whose folder is gone is said not found, and refused when chosen
      Given the workspace "work", the default
      And the workspace "home"
      And the folder of the workspace "home" is gone
      When I open the board
      And I press "ctrl+o"
      Then the workspace "home" is marked "not found"
      When I press "1"
      Then the board says "the folder of this workspace is not there"
      And the board says "OPEN A WORKSPACE"
      When I press "esc"
      Then the header opens with "● work"

    @tui-workspace-time-off-there-takes-the-time-tab
    Scenario: the time off in the settings of the workspace opened takes the time tab away
      Given the workspace "work", the default
      And the workspace "home"
      And the settings of the workspace "home" turn the time off
      When I open the board
      Then the board says "time"
      When I press "ctrl+o", "1"
      Then the header opens with "● home"
      And the board does not say "time"

    @tui-workspace-the-header-wears-its-colour
    Scenario: the dot of the header wears the colour the workspace's settings say
      Given the workspace "work", the default
      And the settings of the workspace "work" say its colour is "green"
      When I open the board
      Then the dot of the header wears "green"

    @tui-workspace-recent-folders-listed
    Scenario: ctrl+o lists the recent folders under the workspaces, one gone said not found
      Given the workspace "work", the default
      And the recent folder "scratch", opened yesterday
      And the recent folder "trip", whose store is gone
      When I open the board
      And I press "ctrl+o"
      Then the board says "RECENT FOLDERS"
      And the recent folder "scratch" is marked "opened yesterday"
      And the recent folder "trip" is marked "not found now"

    @tui-workspace-a-recent-folder-opens-in-place
    Scenario: a recent folder chosen opens in place of the board
      Given the workspace "work", the default
      And the folder "scratch" holds a task titled "Plant the bulbs"
      And the recent folder "scratch", opened yesterday
      When I open the board
      And I press "ctrl+o", "down", "enter"
      Then the header opens with "◌ " and the folder "scratch"
      And "BACKLOG" carries "Plant the bulbs"

    @tui-workspace-a-folder-opened-in-place
    Scenario: ctrl+o in the list opens a folder typed, in place, and remembers it
      Given the workspace "work", the default
      And the folder "scratch" holds a task titled "Plant the bulbs"
      When I open the board
      And I press "ctrl+o", "ctrl+o"
      And I type the folder "scratch"
      And I press "enter"
      Then the header opens with "◌ " and the folder "scratch"
      And "BACKLOG" carries "Plant the bulbs"
      When I press "ctrl+o"
      Then the recent folder "scratch" is marked "opened today"

    @tui-workspace-a-relative-folder-is-refused
    Scenario: a folder typed that is not absolute is refused, and the board stays
      Given the workspace "work", the default
      When I open the board
      And I press "ctrl+o", "ctrl+o"
      And I type "scratch"
      And I press "enter"
      Then the board says "must be absolute"
      When I press "esc", "esc"
      Then the header opens with "● work"

    @tui-workspace-a-new-one-made-and-opened
    Scenario: ctrl+n in the list makes a workspace and opens it in place
      Given the workspace "work", the default
      When I open the board
      And I press "ctrl+o", "ctrl+n"
      And I type "garden"
      And I press "enter"
      And I type the folder "cairn/garden"
      And I press "enter"
      Then the header opens with "● garden"
      And the configuration names the workspace "garden"
      And the default workspace is still "work"

    @tui-workspace-a-new-one-refuses-a-name-known
    Scenario: a new workspace under a name already there is refused, as workspace add refuses it
      Given the workspace "work", the default
      When I open the board
      And I press "ctrl+o", "ctrl+n"
      And I type "work"
      And I press "enter"
      And I type the folder "cairn/other"
      And I press "enter"
      Then the board says "already there"
      And the header opens with "● work"

    @tui-workspace-a-recent-folder-kept
    Scenario: ctrl+k on a recent folder keeps it as a workspace
      Given the workspace "work", the default
      And the folder "scratch" holds a task titled "Plant the bulbs"
      And the recent folder "scratch", opened yesterday
      When I open the board
      And I press "ctrl+o", "down", "ctrl+k"
      And I type "scratch"
      And I press "enter"
      Then the workspaces listed read "scratch", "work"
      And the board does not say "RECENT FOLDERS"
      And the configuration names the workspace "scratch"

  Rule: The board wears the terminal's own colours, whatever theme the terminal has

    @tui-colours-follow-the-terminal
    Scenario: on the fixtures, every screen is drawn in the terminal's own colours alone
      Given the fixtures
      When I open the board
      Then the board says "Done this week"
      And no screen wears a colour the terminal's theme does not set

    @tui-colours-faint-is-the-terminals-own-foreground
    Scenario: the rule between two columns is faint, in the terminal's own foreground
      When I open the board
      Then the rule between two columns is faint, in no colour of its own

    @tui-colours-a-priority-is-one-of-the-terminals
    Scenario Outline: the left edge of a card wears its priority in one of the terminal's colours
      Given I added a task titled "Ask legal" in "todo" at the priority "<priority>" carrying the label "spec"
      When I open the board
      Then the left edge of the card for task 1 wears the terminal's <colour>

      Examples:
        | priority | colour         |
        | low      | cyan           |
        | normal   | own foreground |
        | high     | yellow         |
        | critical | red            |

    @tui-colours-the-tab-open-is-reversed
    Scenario: the tab open is the terminal's blue, reversed onto its ground
      When I open the board
      Then the tab "TASKS" is drawn reversed in the terminal's blue

    @tui-colours-a-workspace-wears-the-nearest-of-the-terminals
    Scenario Outline: the colour of a workspace is the nearest the terminal has
      Given the workspace "work", the default
      And the settings of the workspace "work" say its colour is "<colour>"
      When I open the board
      Then the dot of the header wears the terminal's <terminal>

      Examples:
        | colour   | terminal       |
        | green    | green          |
        | lavender | bright blue    |
        | peach    | bright yellow  |
        | maroon   | bright red     |

  Rule: Every screen fits the terminal it is drawn in, whatever the store holds

    @tui-the-fixtures-fit
    Scenario Outline: on the fixtures, no screen runs past the terminal
      Given the fixtures
      And a window of <width> by <height>
      When I open the board
      Then the board says "Done this week"
      And no screen runs past the terminal

      Examples:
        | width | height |
        | 80    | 26     |
        | 120   | 40     |

  Rule: A view keeps the board's filters under a name, and opens them again in one act

    Background:
      Given I added a task titled "Write the spec" in "todo" at the priority "high" carrying the label "backend"
      And I added a task titled "Review the plan" in "todo" at the priority "normal" carrying the label "backend"
      And I added a task titled "Ship it" in "current" at the priority "high" carrying the label "sprint"

    @tui-view-v-lists-the-views
    Scenario: v turns the foot into the views, each after the figure that opens it
      Given I kept the view "Sprint" holding the label "sprint"
      And I kept the view "High" holding the priority "high"
      When I open the board
      And I press "v"
      Then the mode is "VIEWS"
      And the board says "1 Sprint 1"
      And the board says "2 High 2"
      And the board says "0 whole board"

    @tui-view-a-figure-opens-it
    Scenario: a figure after v opens the view in that place, and the line above the foot names it
      Given I kept the view "Sprint" holding the label "sprint"
      When I open the board
      And I press "v", "1"
      Then the mode is "NORMAL"
      And the board carries 1 task
      And the board says "view Sprint · 1 of 3 tasks"

    @tui-view-0-goes-back-to-the-whole-board
    Scenario: 0 after v leaves the view for the whole board, and esc leaves what is in force
      Given I kept the view "Sprint" holding the label "sprint"
      When I open the board
      And I press "v", "1", "v", "esc"
      Then the board carries 1 task
      When I press "v", "0"
      Then the board carries 3 tasks
      And the board does not say "view Sprint"

    @tui-view-new-keeps-the-filters
    Scenario: :view new keeps the filters in force under the name given
      When I open the board
      And I press "shift+f", "l", "j", "j", "j", "esc"
      And I press ":"
      And I type "view new Urgent"
      And I press "enter"
      Then the view "Urgent" holds the priority "high"
      And the board says "view Urgent"

    @tui-view-save-puts-the-change-in
    Scenario: a filter on top of a view is marked, and :view save puts it into the view
      Given I kept the view "High" holding the priority "high"
      When I open the board
      And I press "v", "1"
      And I press "shift+f", "l", "l", "j", "esc"
      Then the board says "view High *"
      When I press ":"
      And I type "view save"
      And I press "enter"
      Then the view "High" holds the label "backend"
      And the board does not say "view High *"

    @tui-view-opened-by-its-name
    Scenario: :view and a name opens that view
      Given I kept the view "Sprint" holding the label "sprint"
      When I open the board
      And I press ":"
      And I type "view sprint"
      And I press "enter"
      Then the board carries 1 task
      And the board says "view Sprint"

    @tui-view-outside-is-a-reference
    Scenario: a task outside the view touching one inside is shown with why, and not counted
      Given I kept the view "Sprint" holding the label "sprint"
      And I edited task 3 with the blocker "1"
      And I linked task 2 to task 3
      When I open the board
      And I press "v", "1"
      Then the board carries 1 task
      And the board says "blocks 3"
      And the board says "linked to 3"

    @tui-view-new-proposes-the-objective
    Scenario: an objective's tasks shown are kept under the objective's name when none is given
      Given I added an objective titled "Home automation"
      And I added a task titled "Buy the relay" serving objective 1
      When I open the board
      And I press "tab", "t"
      And I press ":"
      And I type "view new"
      And I press "enter"
      Then the view "Home automation" holds objective 1

Rule: A task that comes back is told from another at a glance, and its series has a tab of its own

  @tui-series-a-card-carries-the-loop-and-the-rule
  Scenario: a recurring card carries ↻ and its rule on the line its title leaves free
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every weekday 10:30"
    When I open the board
    Then the card for task 1 says "↻"
    And the card for task 1 says "every weekday 10:30"

  @tui-series-the-detail-says-the-rule
  Scenario: an open occurrence says its series, its rule and what it did so far
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "enter"
    Then the board says "↻1 every day, on the calendar"
    And the board says "0 done · 0 skipped · 0 missed"
    And the bar offers "skip"

  @tui-series-the-trace-says-made-from-series
  Scenario: an occurrence its series made says so in its trace
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And task 1 was skipped
    And a window of 140 by 44
    When I open the board
    And I press "l", "i", "enter"
    Then the board says "made from series  ↻1"

  @tui-series-the-trace-goes-on-with-the-series-before-it
  Scenario: the trace of an occurrence goes on with what its series did before it
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And task 1 was skipped
    And a window of 140 by 44
    When I open the board
    And I press "l", "i", "enter"
    Then the board says "in the series, before this one"
    And the board says "skipped #1"

  @tui-series-the-tab-says-it-as-drawn
  Scenario: the series tab writes a rule the way the mockup wrote it
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every weekday 10:30"
    When I open the board
    And I press "tab", "tab", "tab"
    Then the board says "every weekday · 10:30"

  @tui-series-the-board-narrows-to-what-comes-back
  Scenario: the series filter keeps the tasks that come back
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And I added a task titled "Write the spec"
    When I open the board
    And I press "shift+f", "l", "l", "l", "l", "l", "j"
    Then the board says "series all"
    And the board says "Stand-up"
    And the board says nothing of "Write the spec"

  @tui-the-board-narrows-to-what-is-in-focus
  Scenario: the focus filter keeps the tasks in focus
    Given I added a task titled "Write the spec"
    And I added a task titled "Fix the gate"
    And I put task 2 in focus
    When I open the board
    And I press "shift+f", "l", "l", "l", "l", "l", "j"
    Then the board says "focus in focus"
    And the board says "Fix the gate"
    And the board says nothing of "Write the spec"

  @tui-series-r-s-skips
  Scenario: r s skips the occurrence under the cursor, and its series makes the next
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "r", "s"
    Then task 1 is archived
    And task 2 is titled "Stand-up"
    And task 2 is in the column "todo"

  @tui-series-r-x-stops-and-r-r-resumes
  Scenario: r x stops the series, leaving the task on the board, and r r resumes it
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "r", "x"
    Then series 1 is stopped
    And task 1 is in the column "todo"
    When I press "r", "r"
    Then series 1 is active

  @tui-series-r-e-opens-its-series
  Scenario: r e opens the series of the occurrence in its tab
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "r", "e"
    Then the tab is "SERIES"
    And the board says "TEMPLATE"

  @tui-series-r-is-offered-only-on-a-recurring-card
  Scenario: the r keys are on the bar only while the card under the cursor comes back
    Given recurrence is on on the board
    And I added a task titled "Write the spec"
    When I open the board
    And I press "i"
    Then the bar does not offer "skip"

  @tui-series-the-tabs-count-what-they-hold
  Scenario: each tab counts what it holds, as the window's tabs do
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And I added a task titled "Write the spec"
    And I added an objective titled "Ship the board"
    When I open the board
    Then the board says "TASKS 2"
    And the board says "objectives 1"
    And the board says "time 0"
    And the board says "series 1"

  @tui-the-time-tab-counts-before-it-is-opened
  Scenario: the time tab counts the tasks worked on this week before it is ever opened
    Given I added a task titled "Write the spec"
    And I spent "1h" on task 1 on "2026-09-01"
    When I open the board
    Then the board says "time 1"

  @tui-series-the-tab-lists-them
  Scenario: the series tab lists the active series apart from the stopped ones
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And I added a task titled "Expense report" coming back "every 4 weeks on monday"
    And I stopped series 2
    When I open the board
    And I press "tab", "tab", "tab"
    Then the tab is "SERIES"
    And the board says "Active"
    And the board says "Stopped"
    And the board says "↻1"
    And the board says "↻2"
    And the board says "stopped since"

  @tui-series-the-tab-reaches-an-archived-occurrence
  Scenario: an occurrence archived long ago is reached from its series
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And task 1 was skipped
    When I open the board
    And I press "tab", "tab", "tab", "enter", "j", "enter"
    Then the mode is "DETAILS"
    And the open task is task 1

  @tui-series-the-tab-skips-stops-and-resumes
  Scenario: in the tab, s skips the open occurrence, x stops the series and r resumes it
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "tab", "tab", "tab", "s"
    Then task 1 is archived
    When I press "x"
    Then series 1 is stopped
    And task 2 is in the column "todo"
    When I press "r"
    Then series 1 is active

  @tui-series-deleting-keeps-the-tasks-unless-told
  Scenario: deleting a series asks what becomes of its tasks, keeping them unless told
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "tab", "tab", "tab", "D"
    Then the board says "delete the series ↻1 Stand-up?"
    And the board says "keep its tasks, as ordinary tasks"
    When I press "enter"
    Then there is no series 1
    And task 1 is in the column "todo"

  @tui-series-deleting-the-open-one
  Scenario: deleting a series may delete the one open and keep the rest
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "tab", "tab", "tab", "D", "j", "enter"
    Then there is no series 1
    And there is no task 1

  @tui-series-deleting-every-task-asks-once-more
  Scenario: deleting a series with every task it made asks once more, naming the count
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And task 1 was skipped
    When I open the board
    And I press "tab", "tab", "tab", "D", "j", "j", "enter"
    Then the board says "delete all 2 tasks too?"
    And series 1 is active
    When I press "enter"
    Then there is no series 1
    And there is no task 1
    And there is no task 2

  @tui-series-the-form-changes-the-series
  Scenario: e in the tab opens the series' form, and saving changes what comes next
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "tab", "tab", "tab", "e"
    And I set the field "title" to "Daily stand-up"
    And I save the form
    Then series 1 is titled "Daily stand-up"
    And task 1 is titled "Stand-up"

  @tui-series-n-makes-a-new-one
  Scenario: n in the tab opens a task's form with the rule to fill in
    Given recurrence is on on the board
    When I open the board
    And I press "tab", "tab", "tab", "n"
    And I set the field "title" to "Water the plants"
    And I set the field "rule" to "every 3 days after done"
    And I save the form
    Then series 1 is titled "Water the plants"
    And task 1 is an occurrence of series 1

  @tui-series-n-lands-on-the-rule
  Scenario: n in the tab opens a task's form with the cursor on its rule
    Given recurrence is on on the board
    When I open the board
    And I press "tab", "tab", "tab", "n"
    Then the cursor is on the field "rule"

  @tui-series-a-task-given-a-rule-becomes-the-first
  Scenario: a task given a rule in its form becomes the first occurrence of a new series
    Given recurrence is on on the board
    And I added a task titled "Water the plants"
    When I open the board
    And I press "i", "e"
    And I set the field "rule" to "every 3 days"
    And I save the form
    Then task 1 is an occurrence of series 1

  @tui-series-saving-an-occurrence-asks-this-one-or-all
  Scenario: changing the title of an occurrence asks whether the series changes too
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "e"
    And I set the field "title" to "Daily stand-up"
    And I press "ctrl+s"
    Then the board says "this one only"
    When I press "s"
    Then task 1 is titled "Daily stand-up"
    And series 1 is titled "Daily stand-up"

  @tui-series-saving-this-one-only-leaves-the-series
  Scenario: saving an occurrence for this one only leaves its series as it was
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "l", "i", "e"
    And I set the field "title" to "Daily stand-up"
    And I press "ctrl+s"
    And I press "enter"
    Then task 1 is titled "Daily stand-up"
    And series 1 is titled "Stand-up"

  @tui-series-the-fixtures-fit
  Scenario: on the fixtures, the series tab and its form run past no terminal
    Given recurrence is on on the board
    And the fixtures
    And a window of 80 by 26
    When I open the board
    Then no screen runs past the terminal

  @tui-series-off-shows-nothing
  Scenario: with recurrence off, there is no ↻, no rule, no r key and no series tab
    Given recurrence is off on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    Then the board does not say "↻"
    And the board does not say "every day"
    When I press "tab", "tab"
    Then the tab is "TIME"
    When I press "tab", "l", "i"
    Then the bar does not offer "skip"

  Rule: The configuration is read again on the board's beat, and a key is set from the : line

    @tui-config-read-again-says-what-changed
    Scenario: settings changed by hand are read again at the next beat, applied, and what changed is said
      Given the workspace "home", the default
      When I open the board
      Then the board says "time 0"
      When the settings of the workspace "home" are written by hand:
        """
        [pages]
        time = "off"
        """
      And the beat comes
      Then the board says "settings.toml read again: pages.time on → off"
      And the board does not say "time 0"

    @tui-config-a-file-that-cannot-be-read-keeps-what-it-had
    Scenario: settings that cannot be read leave the board as it was, and say where and why until they read
      Given the workspace "home", the default
      When I open the board
      And the settings of the workspace "home" are written by hand:
        """
        [pages]
        time = "hiden"
        """
      And the beat comes
      Then the board says "settings.toml:2"
      And the board says "hiden"
      And the board says "the board keeps what it read at 10:00"
      And the board says "time 0"
      When I press "l"
      Then the board says "the board keeps what it read at 10:00"
      When the settings of the workspace "home" are written by hand:
        """
        [pages]
        time = "off"
        """
      And the beat comes
      Then the board does not say "the board keeps what it read"
      And the board does not say "time 0"

    @tui-set-a-key-from-the-command-line
    Scenario: :set writes a key in the settings of the workspace, and the board wears it at once, saying only what it wrote
      Given the workspace "home", the default
      When I open the board
      And I press ":"
      And I type "set pages.time off"
      And I press "enter"
      Then the board says "written in"
      And the board does not say "time 0"
      And the settings of the workspace "home" say `time = "off"`
      When the beat comes
      Then the board does not say "read again"

    @tui-set-a-key-in-the-global-configuration
    Scenario: :set --global writes the global configuration
      Given the workspace "home", the default
      When I open the board
      And I press ":"
      And I type "set --global focus.session 45m"
      And I press "enter"
      Then the global configuration says `session = "45m"`

    @tui-set-refuses-what-the-key-does-not-take
    Scenario: a value the key does not take is said, and nothing is written
      Given the workspace "home", the default
      When I open the board
      And I press ":"
      And I type "set pages.time hiden"
      And I press "enter"
      Then the board says "is not a word pages.time takes (on, hidden, off)"
      And the board says "time 0"

    @tui-unset-a-key
    Scenario: :unset takes the key's line out, and the value under it answers again
      Given the workspace "home", the default
      And the settings of the workspace "home" turn the time off
      When I open the board
      And I press ":"
      And I type "unset pages.time"
      And I press "enter"
      Then the board says "pages.time is no longer said"
      And the board says "time 0"

    @tui-set-completes-the-keys
    Scenario: the keys are completed from the catalogue, each with its value and where it comes from
      Given the workspace "home", the default
      And the settings of the workspace "home" turn the time off
      When I open the board
      And I press ":"
      And I type "set pages.t"
      Then the board says "pages.time"
      And the board says "pages.tasks"
      And the board says "set here ← global: on"
      And the board says "default"
      And the board does not say "window.theme"
      When I type "i"
      And I press "tab"
      Then the command line reads "set pages.time "
