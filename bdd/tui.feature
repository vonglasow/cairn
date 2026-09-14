Feature: The board
  The same cairn, laid out on a screen that stays open. Four columns for where
  the work stands, a band underneath for what is not moving, and a second tab
  for what the work is for.

  A board decides nothing. It draws what the engine holds, and every key that
  changes something goes back through the engine — so what is done here is done
  the same way it is done at the command line.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a window of 100 by 44

  Rule: The work is laid out in four columns and a band

    Left to right, the order the work goes through: backlog, todo, current,
    done. The three states that mean "not moving" share a band underneath,
    because what tells them apart is what unblocks them, not where they sit.

    @tui-four-columns-left-to-right
    Scenario: the columns are named in the order the work goes through
      When I open the board
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

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

    Four columns need sixty columns of terminal. Below that a title says
    nothing anybody can read, so the board says what it needs instead of
    drawing something useless.

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

    Two modes, and which one you are in says what the cursor is holding. The
    bar at the foot says which, so it is never guessed.

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

    Reading across a board is reading across it. A cursor that jumped back to
    the first card every time it changed column would make the third task in
    one column and the third in the next two separate journeys.

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

    Three cards to a row, so along a row is left and right and down is a row
    lower. A key that pointed down and moved the cursor sideways would be a key
    nobody could trust.

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

  Rule: Moving a task changes where it stands

    Sideways it changes column, and the state follows. Up and down it changes
    place among the ones it sits with, and never the column — until there is
    nowhere lower in the column, and what is below the column is the band.

    @tui-moving-a-task-across-changes-its-column
    Scenario: pushing a task right puts it in the next column
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "i", "shift+l"
      Then task 1 is in the column "todo"

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

    A key on the bar does something in this view. A key that does nothing here
    is not on it, so the bar is never read against what the board will refuse.

    @tui-the-bar-follows-the-mode
    Scenario: the bar changes with the mode
      Given I added a task titled "Write the spec"
      When I open the board
      Then the bar does not offer "move"
      When I press "i"
      Then the bar offers "move"

  Rule: Help answers in every mode

    Help that cannot be reached from the mode you are stuck in is help you have
    to already understand to find.

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

    The columns stay, so where a match sits in the flow is part of the answer.

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

    @tui-leaving-a-search-gives-the-board-back
    Scenario: escape ends the search
      Given I added a task titled "Write the spec"
      And I added a task titled "Buy a lamp"
      When I open the board
      And I press "/"
      And I type "spec"
      And I press "esc"
      Then the board carries 2 tasks

  Rule: A card says what its task carries, and no more

    Twenty characters of width decide what fits. What is on a card was chosen
    once and holds everywhere.

    @tui-a-card-carries-the-number-and-the-title
    Scenario: a card is its number and its title
      Given I added a task titled "Write the spec"
      When I open the board
      Then the card for task 1 says "1"
      And the card for task 1 says "Write the spec"

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

    @tui-the-task-in-focus-is-told-apart
    Scenario: the task in hand is drawn unlike any other
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      When I open the board
      Then the card for task 1 is drawn as the one in focus

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

  Rule: Nothing is written until it is saved

    The same form adds and edits, and it holds everything back until it is
    told to keep it — otherwise cancelling would mean nothing.

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

  Rule: The detail shows what a card cannot

    @tui-the-detail-shows-the-description-and-the-history
    Scenario: opening a task shows what the card only marked
      Given I added a task titled "Write the spec" with the description "The client wants two columns"
      When I open the board
      And I press "i", "enter"
      Then the board says "The client wants two columns"
      And the board says "created"

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
