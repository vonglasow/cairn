Feature: The desktop board
  The same cairn as the terminal board, in a window a mouse can reach. Four
  columns for where the work stands, a band underneath for what is not moving,
  and a second tab for what the work is for.

  A window decides nothing. It draws what the engine holds, and every gesture
  that changes something goes back through the engine — so what is done here is
  done the same way it is done at the command line.

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a window of 1200 by 800

  Rule: The work is laid out the way the terminal lays it out

    The same shape, with more room. A person who has learnt one board has
    learnt the other — sub-tasks left off both, since a part of the work is
    read where the work is read.

    @gui-four-columns-left-to-right
    Scenario: the columns are named in the order the work goes through
      When I open the window
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @gui-archived-are-off-the-board-until-asked-for
    Scenario: what was archived comes back when it is asked for
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      And I archived task 1
      When I open the window
      Then the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      Then "BACKLOG" carries "Migration Jenkins"
      And the card "Migration Jenkins" is dimmed
      When I press "shift+a"
      Then the board says nothing of "Migration Jenkins"

    @gui-a-card-is-taken-off-the-board
    Scenario: a chosen card is taken off the board, and brought back
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      When I open the window
      And I click the card "Migration Jenkins"
      Then the bar offers "Archive" under "x"
      When I choose "Archive"
      Then the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      And I click the card "Migration Jenkins"
      Then the bar offers "Restore" under "x"
      When I choose "Restore"
      Then "BACKLOG" carries "Migration Jenkins"
      And the card "Migration Jenkins" is not dimmed

    @gui-a-sub-task-is-not-on-the-board
    Scenario: a sub-task does not stand in a column of its own
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" hanging under task 1
      And I added a task titled "Something else"
      When I open the window
      Then "BACKLOG" carries "Ship it"
      And "BACKLOG" carries "Something else"
      And the board says nothing of "Write the spec"

    @gui-an-empty-column-stays
    Scenario: a column with nothing in it is still drawn
      Given I added a task titled "Write the spec"
      When I open the window
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @gui-a-task-is-drawn-in-its-column
    Scenario: a task is drawn under the column it stands in
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      Then "TODO" carries "Write the spec"

    @gui-the-band-has-three-lanes
    Scenario: the band names the three kinds of stopped
      When I open the window
      Then the band is named "STALLED"
      And the band's lanes read "blocked", "watching", "paused"

    @gui-a-stopped-task-sits-in-its-lane
    Scenario: a stopped task sits in the lane of its kind
      Given I added a task titled "Client's answer" with the state "blocked"
      And I added a task titled "Handed over" with the state "watching"
      When I open the window
      Then the lane "blocked" carries "Client's answer"
      And the lane "watching" carries "Handed over"

    @gui-two-tabs
    Scenario: there are two tabs and the work is the first
      When I open the window
      Then the tabs read "TASKS", "OBJECTIVES"
      And the tab is "TASKS"

    @gui-clicking-a-tab-changes-it
    Scenario: clicking the other tab changes to it
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the tab is "OBJECTIVES"
      And I click the tab "TASKS"
      Then the tab is "TASKS"

    @gui-the-objectives-tab-lists-them
    Scenario: the objectives tab lists what the work is for
      Given I added an objective titled "Ship cairn"
      And I added an objective titled "Keep it honest"
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the objectives read "Ship cairn", "Keep it honest"

    @gui-choosing-an-objective
    Scenario: clicking an objective chooses it
      Given I added an objective titled "Ship cairn"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      Then the objective "Ship cairn" is chosen

    @gui-editing-an-objective
    Scenario: an objective is edited through the same form
      Given I added an objective titled "Ship cairn"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      And I choose "Edit"
      And I set the field "title" to "Ship cairn properly"
      And I choose "Save"
      Then objective 1 is called "Ship cairn properly"

  Rule: Cards are the same height, whatever they carry

    A fixed height is what lets a column say how many fit and which one is
    under the pointer without drawing anything.

    @gui-cards-share-one-height
    Scenario: a card with a great deal on it is as tall as an empty one
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it" with the description "Two columns, the client says"
      When I open the window
      Then every card in "BACKLOG" is the same height

    @gui-a-long-title-is-cut
    Scenario: a title too long for its card is cut rather than wrapped
      Given I added a task titled "A title far longer than any column could ever hope to show in full"
      When I open the window
      Then "BACKLOG" carries a card whose title ends in "…"

  Rule: A window too small says what it needs

    Below what a card needs to stay readable, the window says so rather than
    drawing something nobody can use.

    @gui-too-small-says-so
    Scenario: a window under the smallest size says what it needs
      Given a window of 320 by 240
      When I open the window
      Then it says the window is too small
      And it says what size it needs

    @gui-growing-back-draws-again
    Scenario: making the window big enough again draws the board
      Given a window of 320 by 240
      When I open the window
      And I resize the window to 1200 by 800
      Then the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @gui-growing-back-gives-the-form-back
    Scenario: a window dragged narrow and back gives back what was open
      Given a window of 1200 by 800
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I resize the window to 320 by 240
      Then it says the window is too small
      When I resize the window to 1200 by 800
      Then the window is on the form

  Rule: A click outside a panel closes it, where closing loses nothing

    A hand on the mouse should not have to reach for the keyboard to put
    something away. Clicking off a panel does what esc does — an open task
    nobody typed into, the help, the question about deleting, the list of
    objectives — because in each of those the way out costs nothing.

    A form somebody typed into is the exception, and the only one. It holds
    what was typed and not saved, and a click landing anywhere outside it would
    throw that away without asking. Losing work without being asked is how
    work is lost; esc is still there, and it is offered on the panel as Cancel.

    Closing a panel closes the panel and nothing else. A search that was
    narrowing the board stays in force, because it is not what was opened: a
    task found by searching, opened, changed and put away, gave back the whole
    board and left somebody to type their search again. The second esc is what
    clears it, which is what esc on the board has always done.

    @gui-a-double-click-is-not-a-click-off-what-it-opened
    Scenario: opening a card by double-clicking it does not close it again
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      Then "Write the spec" is open to be edited

    @gui-clicking-off-an-open-task-closes-it
    Scenario: clicking off an open task nobody typed into puts it away
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      And I click outside the panel
      Then the window is on the board

    @gui-clicking-off-the-help-closes-it
    Scenario: clicking off the help puts it away
      When I open the window
      And I press "?"
      And I click outside the panel
      Then the help is closed

    @gui-clicking-off-a-form-keeps-it
    Scenario: clicking off a form keeps what was typed
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type "X"
      And I click outside the panel
      Then the form is open on task 1
      And the field "title" reads "Write the specX"

    @gui-closing-a-panel-keeps-the-search
    Scenario: putting a task away gives back what the search left, not the board
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "enter"
      And I click the card "Write the spec"
      And I press "enter"
      And I press "esc"
      Then the window is on the board
      And the search is still in force
      And "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"
      When I press "esc"
      Then the search is not in force
      And "BACKLOG" carries "Ship the board"

  Rule: The mouse chooses, opens and scrolls

    No modes and no cursor to move. What a person points at is what happens.

    @gui-a-click-chooses
    Scenario: clicking a card chooses it
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      Then the card "Write the spec" is chosen

    @gui-a-click-elsewhere-chooses-elsewhere
    Scenario: clicking another card chooses that one instead
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the window
      And I click the card "Write the spec"
      And I click the card "Ship it"
      Then the card "Ship it" is chosen
      And the card "Write the spec" is not chosen

    @gui-a-double-click-opens
    Scenario: double-clicking a card opens it to be edited
      Given I added a task titled "Write the spec" with the description "Two columns"
      When I open the window
      And I double-click the card "Write the spec"
      Then "Write the spec" is open to be edited
      And the form shows "Two columns"

    @gui-an-open-task-shows-the-whole-task
    Scenario: an open task shows everything it carries, not only what was done to it
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I double-click the card "Write the spec"
      Then the form shows the field "column" reading "todo"
      And the form shows the field "priority"
      And the form shows the field "labels"

    @gui-an-open-task-says-where-its-context-lives
    Scenario: an open task says where what was written about it elsewhere lives
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      Then the open task says where the context of task 1 lives
      And the form asks nothing about "context"

    @gui-the-context-is-held-past-a-long-description
    Scenario: the place an open task names is held however long its description runs
      Given I added a task titled "Write the spec" with a description of 80 lines ending "the end"
      When I open the window
      And I double-click the card "Write the spec"
      And I triple-click the first word of "field:context" on the open task
      And I press "cmd+c"
      Then the clipboard reads "/vault/1"

    @gui-the-offers-answer-past-a-long-description
    Scenario: what an open task offers answers a click however long its description runs
      Given I added a task titled "Write the spec" with a description of 80 lines ending "the end"
      When I open the window
      And I double-click the card "Write the spec"
      And I click "Cancel"
      Then the window is on the board

    @gui-an-open-task-is-put-away-by-clicking
    Scenario: an open task is put away without knowing a key
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I click "Cancel"
      Then the window is on the board

    @gui-an-open-task-closes
    Scenario: an open task closes and leaves the board as it was
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I press "esc"
      Then the window is on the board
      And the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @gui-the-wheel-scrolls-one-column
    Scenario: the wheel scrolls the column under the pointer and no other
      Given I added 12 tasks to "backlog"
      And I added 12 tasks to "todo"
      When I open the window
      And I scroll down over "BACKLOG"
      Then "BACKLOG" is scrolled
      And "TODO" is not scrolled

    @gui-the-wheel-scrolls-one-lane
    Scenario: the wheel scrolls the lane of the band under the pointer and no other
      Given I added 6 tasks to "watching"
      And I added 6 tasks to "blocked"
      When I open the window
      And I scroll down over the lane "watching"
      Then the lane "watching" is scrolled
      And the lane "blocked" is not scrolled

    @gui-a-click-in-a-scrolled-lane
    Scenario: a click in a scrolled lane chooses the card drawn under the pointer
      Given I added 6 tasks to "watching"
      When I open the window
      And I scroll down over the lane "watching"
      And I click the top of the lane "watching"
      Then the card "Task 2" is chosen

  Rule: The arrows do what the pointer does

    A hand on the keyboard should not have to reach for the mouse to look at
    the next task. The terminal board walks with the arrows; so does this one.

    @gui-the-arrows-walk-the-cards
    Scenario: the arrows walk from one card to the next
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      When I open the window
      And I click the card "First"
      And I press "down"
      Then the card "Second" is chosen
      And I press "up"
      Then the card "First" is chosen

    @gui-the-arrows-walk-the-columns
    Scenario: the arrows walk from one column to the next
      Given I added a task titled "In the heap" with the state "backlog"
      And I added a task titled "Up next" with the state "todo"
      When I open the window
      And I click the card "In the heap"
      And I press "right"
      Then the card "Up next" is chosen
      And I press "left"
      Then the card "In the heap" is chosen

    @gui-enter-opens-what-is-chosen
    Scenario: enter opens what the arrows walked to
      Given I added a task titled "Write the spec" with the description "two columns"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then "Write the spec" is open to be edited

  Rule: Dragging a card moves the work

    A drop says two things at once — which column, and which place among the
    cards there — and the engine is told them in that order.

    @gui-a-small-move-is-still-a-click
    Scenario: a click that barely moves is a click
      Given I added a task titled "Write the spec"
      When I open the window
      And I press the card "Write the spec" and move it 2 pixels and let go
      Then the card "Write the spec" is chosen
      And nothing was asked of the engine

    @gui-a-drag-shows-where-it-would-land
    Scenario: a card being dragged says where it would go
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "backlog"
      When I open the window
      And I take hold of the card "Second" and move it over "TODO"
      Then the card being dragged follows the pointer
      And it shows where the card would land in "TODO"

    @gui-a-drag-changes-the-column
    Scenario: dropping a card in another column changes its state
      Given I added a task titled "Write the spec" with the state "backlog"
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      Then the engine was asked "task edit" for task 1
      And task 1 is in the column "todo"

    @gui-a-drag-carries-the-place
    Scenario: dropping a card above another puts it there
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Third" with the state "backlog"
      When I open the window
      And I drag the card "Third" onto "TODO" above "Second"
      Then "TODO" reads "First", "Third", "Second"

    @gui-a-drag-within-a-column-only-moves
    Scenario: dropping a card in its own column changes the place and not the state
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      When I open the window
      And I drag the card "Second" onto "TODO" above "First"
      Then the engine was asked "task move" for task 2
      And the engine was not asked "task edit"
      And "TODO" reads "Second", "First"

    @gui-a-drop-in-a-lane-says-which-kind
    Scenario: dropping a card in a lane of the band says which kind of stopped
      Given I added a task titled "Client's answer" with the state "todo"
      When I open the window
      And I drag the card "Client's answer" onto the lane "watching"
      Then task 1 is in the column "watching"
      And the lane "watching" carries "Client's answer"

    @gui-a-drop-that-is-refused-shows-the-truth
    Scenario: a drop the engine refuses leaves the board saying what is really there
      Given I added a task titled "Write the spec" with the state "backlog"
      And the store refuses what it is asked to change
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      Then it says what went wrong
      And "BACKLOG" carries "Write the spec"

  Rule: A card says when it is wanted and what it has cost

    The window knew nothing about time: no deadline, no notion of late, where
    the terminal has written it on the card all along. Deciding what to do
    without knowing what is due is not deciding.

    Beside it, what the work has cost against what it was reckoned to cost, and
    what it serves — three facts an open task carried and no overview summarised.

    @gui-a-card-says-when-it-is-due
    Scenario: a day close enough to matter is written on the card
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-10-01"
      When I open the window
      Then the card "Write the spec" says "2d"

    @gui-a-card-says-nothing-of-a-day-months-away
    Scenario: a day too far off to act on is not written on the card
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2027-06-01"
      When I open the window
      Then the card "Write the spec" says nothing of "2027"

    @gui-a-card-says-what-it-has-cost
    Scenario: what a task has cost is written beside what it was reckoned to cost
      Given I added a task titled "Write the spec" with the estimate "2h"
      When I open the window
      Then the card "Write the spec" says "2h"

  Rule: A form is the whole of a task

    Adding and editing are one form. Every field is asked about, including the
    empty ones, so a field left blank is a field somebody emptied.

    @gui-editing-is-reachable-from-the-board
    Scenario: a chosen task is edited without knowing a shortcut
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I click "Edit"
      Then the form asks about "title", "description"

    @gui-the-form-asks-about-every-field
    Scenario: the form asks about everything a task carries
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      Then the form asks about "title", "description", "column", "priority"
      And the form asks about "labels", "blockers", "because", "deadline"
      And the form asks about "estimate", "time spent", "parent", "objectives"

    @gui-the-form-starts-on-what-is-there
    Scenario: the form starts on what the task already carries
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      Then the field "title" reads "Write the spec"
      And the field "column" reads "todo"

    @gui-typing-into-a-field
    Scenario: what is typed into a field is what is saved
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " properly"
      And I choose "Save"
      Then task 1 is called "Write the spec properly"

    @gui-taking-a-character-back
    Scenario: a character taken back is not saved
      Given I added a task titled "Write the specx"
      When I open the window
      And I click the card "Write the specx"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "backspace"
      And I choose "Save"
      Then task 1 is called "Write the spec"

    @gui-walking-the-fields
    Scenario: the keyboard walks from one field to the next
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "tab"
      Then the keyboard is on "description"
      And I press "shift+tab"
      Then the keyboard is on "title"

    @gui-a-field-with-a-closed-set
    Scenario: a field with a closed set is changed without typing
      Given I added a task titled "Write the spec" with the state "backlog"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "column"
      And I press "right"
      And I choose "Save"
      Then task 1 is in the column "todo"

    @gui-a-closed-set-is-chosen-from-a-list
    Scenario: a column is chosen from a list, the way a parent is
      Given I added a task titled "Write the spec" with the state "backlog"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "column"
      Then the window is choosing
      And the list reads "backlog", "todo", "current", "blocked", "watching", "paused", "done"
      And the row the keyboard is on is "backlog"
      And the list has taken 1
      When I press "down"
      And I press "enter"
      Then the window is on the form
      And the field "column" reads "todo"

    @gui-a-priority-is-found-by-typing
    Scenario: a priority is found by typing, the way a parent is
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "priority"
      And I type "hi"
      Then the list reads "high"
      When I press "enter"
      Then the field "priority" reads "high"

    @gui-a-closed-set-is-never-emptied
    Scenario: taking the column a task is already in leaves it there
      Given I added a task titled "Write the spec" with the state "backlog"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "column"
      And I press "enter"
      Then the window is on the form
      And the field "column" reads "backlog"

    @gui-a-standing-is-chosen-from-a-list
    Scenario: an objective's standing is chosen from the same list
      Given I added an objective titled "Ship cairn"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      And I choose "Edit"
      And I click the field "standing"
      Then the window is choosing
      And the list reads "active", "on-hold", "reached", "dropped"

    @gui-clicking-a-field
    Scenario: clicking a field is how the keyboard gets there
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      Then the keyboard is on "deadline"

    @gui-the-caret-says-where-typing-goes
    Scenario: the form says where what is typed will go
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      Then the field "title" shows a caret
      And no other field shows a caret

    @gui-the-caret-moves-with-the-arrows
    Scenario: the caret moves through the text with the arrows
      Given I added a task titled "spec"
      When I open the window
      And I click the card "spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      Then the caret is at 4
      And I press "left"
      And I press "left"
      Then the caret is at 2
      And I type "XY"
      Then the field "title" reads "spXYec"

    @gui-the-caret-stops-at-both-ends
    Scenario: the caret stops at both ends of the text
      Given I added a task titled "ab"
      When I open the window
      And I click the card "ab"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "right"
      Then the caret is at 2
      And I press "left"
      And I press "left"
      And I press "left"
      Then the caret is at 0

    @gui-clicking-into-the-text
    Scenario: clicking in the text puts the caret there
      Given I added a task titled "abcdefgh"
      When I open the window
      And I click the card "abcdefgh"
      And I choose "Edit"
      And I click 3 characters into the field "title"
      Then the caret is at 3

    @gui-taking-back-at-the-caret
    Scenario: what is taken back is what is before the caret
      Given I added a task titled "abcd"
      When I open the window
      And I click the card "abcd"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "left"
      And I press "backspace"
      Then the field "title" reads "abd"

    @gui-the-search-has-a-caret-too
    Scenario: the search line is text like any other
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "left"
      And I type "X"
      Then the search reads "speXc"

    @gui-a-long-field-has-room
    Scenario: a description has room for more than one line
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      Then the field "description" has room for more than one line
      And the field "title" has room for one

    @gui-a-value-longer-than-its-room
    Scenario: a value longer than its room shows its end, where typing happens
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "estimate"
      And I type "a value far longer than the room this field could ever give it, long enough that its beginning has to go somewhere out of sight to leave the end showing"
      Then the field "estimate" shows the end of what was typed

    @gui-saving-writes-once
    Scenario: saving writes what the form holds
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Write the whole spec"
      And I choose "Save"
      Then task 1 is called "Write the whole spec"
      And the engine was asked "task edit" once

    @gui-cancelling-writes-nothing
    Scenario: cancelling leaves the task as it was
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Something else"
      And I choose "Cancel"
      Then task 1 is called "Write the spec"
      And nothing was asked of the engine

    @gui-escape-cancels
    Scenario: escape is the other way out, and writes nothing
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Something else"
      And I press "esc"
      Then task 1 is called "Write the spec"

    @gui-adding-starts-empty
    Scenario: adding starts on an empty form and writes a new task
      When I open the window
      And I press "cmd+n"
      And I set the field "title" to "Write the spec"
      And I choose "Save"
      Then "BACKLOG" carries "Write the spec"

  Rule: What the board can do, the window can do

    A person who has learnt the terminal board has learnt this one. Anything
    the terminal offers and the window does not is a reason to go back to the
    terminal.

    @gui-deleting-a-task
    Scenario: a task is deleted once, having been asked about
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I click "Delete"
      And I click "Agree"
      Then "BACKLOG" does not carry "Write the spec"

    @gui-deleting-asks-first
    Scenario: deleting asks before it loses anything
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I click "Delete"
      And I click "Cancel"
      Then "BACKLOG" carries "Write the spec"

    @gui-searching-narrows-the-board
    Scenario: searching narrows the board to what it reaches
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      Then "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"

    @gui-searching-reaches-the-description
    Scenario: a search reaches what a card does not show
      Given I added a task titled "Write the spec" with the description "two columns"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "columns"
      Then "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"

    @gui-searching-raises-the-work-a-part-belongs-to
    Scenario: a word found only in a part answers the work it is part of
      Given I added a task titled "Ship it"
      And I added a task titled "Write the pipeline" hanging under task 1
      And I added a task titled "Buy a lamp"
      When I open the window
      And I press "/"
      And I type "pipeline"
      Then "BACKLOG" carries "Ship it"
      And "BACKLOG" does not carry "Buy a lamp"

    @gui-searching-by-number
    Scenario: a number is the way back when it is what somebody has
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "2"
      Then "BACKLOG" carries "Ship the board"
      And "BACKLOG" does not carry "Write the spec"

    @gui-searching-reaches-what-the-command-line-reaches
    Scenario: a search reaches every field the command line looks in
      Given I added a task titled "Write the spec" with the estimate "90m"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "1h30m"
      Then "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"

    @gui-a-card-can-be-clicked-while-still-searching
    Scenario: a card is clicked while the search box is still open
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      And I click the card "Write the spec"
      Then the card "Write the spec" is chosen
      And the window is asking what to look for
      And "BACKLOG" does not carry "Ship the board"

    @gui-a-card-opens-while-still-searching
    Scenario: a card opens on a double click while the search box is still open
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "/"
      And I type "spec"
      And I double-click the card "Write the spec"
      Then "Write the spec" is open to be edited

    @gui-clicking-the-search-box-lets-go-of-nothing
    Scenario: clicking the box being typed into does not let go of the card
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "/"
      And I type "spec"
      And I click the card "Write the spec"
      And I click the search box
      Then the card "Write the spec" is chosen
      And the window is asking what to look for

    @gui-a-search-stays-in-force
    Scenario: a search stays in force once it has been asked for
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "enter"
      Then the search box is closed
      And the search is still in force
      And "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"

    @gui-the-board-still-works-under-a-search
    Scenario: what a search leaves can be walked, opened and edited
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "enter"
      And I press "down"
      Then the card "Write the spec" is chosen
      And I press "enter"
      Then "Write the spec" is open to be edited

    @gui-two-ways-into-the-search
    Scenario: the search opens the way a person expects
      When I open the window
      And I press "cmd+f"
      Then the search box is open
      And I press "esc"
      And I press "/"
      Then the search box is open

    @gui-leaving-a-search
    Scenario: leaving a search puts everything back
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "esc"
      Then "BACKLOG" carries "Write the spec"
      And "BACKLOG" carries "Ship the board"

    @gui-the-help
    Scenario: the help says what everything does
      When I open the window
      And I click "Help"
      Then the help is open
      And the help says what "drag" does
      And the help says what "double-click" does
      And I press "esc"
      Then the help is closed

  Rule: The window follows the store

    Somebody else at a command line, or an agent over MCP, is working in the
    same store. What they change shows up without anybody asking.

    @gui-somebody-else-added-something
    Scenario: a task added elsewhere appears
      When I open the window
      And somebody else adds a task titled "Client's answer"
      And the window looks again
      Then "BACKLOG" carries "Client's answer"

    @gui-nothing-changed-nothing-redrawn
    Scenario: a store nobody touched is not read again
      Given I added a task titled "Write the spec"
      When I open the window
      And the window looks again
      Then the store was read once

  Rule: A task's history stays inside the panel that shows it

    A task worked on for a while carries more history than any panel can hold.
    It was drawn down the panel and out of the bottom of the window — a history
    nobody can read, and nothing on screen saying it is there at all.

    @gui-opening-a-task-opens-it-to-be-edited
    Scenario: opening a task opens it to be edited, its history beside what it carries
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I double-click the card "Write the spec"
      Then the window is on the form
      And every field of the form stands left of the history
      And the history begins with the last thing done to task 1

    @gui-opening-an-objective-opens-it-to-be-edited
    Scenario: opening an objective opens it to be edited, the work serving it beside it
      Given I added an objective titled "Ship the board"
      And I added a task titled "Write the spec" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the board"
      And I press "enter"
      Then the window is on the form
      And the work beside the form reads "Write the spec"

    @gui-the-history-has-its-own-place
    Scenario: the history is beside what the task carries, not under it
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the history is beside the fields
      And the history stays inside the panel

    @gui-a-long-description-scrolls
    Scenario: a description longer than its box scrolls, and says so with a bar
      Given I added a task titled "Write the spec" with a description of 20 lines
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the description has more to show than it has room for
      And the description is at its beginning
      And a bar sits beside the description
      When I scroll down over the description
      Then the description has moved
      When I scroll up over the description
      Then the description is at its beginning
      When I scroll down over the description 99 times
      Then the description is at its end
      When I scroll up over the description 99 times
      And I drag the description's bar to the bottom
      Then the description is at its end

    @gui-a-short-description-keeps-its-bar
    Scenario: a description that fits keeps its bar, so it is where it is looked for
      Given I added a task titled "Write the spec" with the description "Two columns"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then a bar sits beside the description
      And the description is at its beginning

    @gui-the-history-reads-newest-first
    Scenario: the history reads from the last thing done, with no blank line between two acts
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the history begins with the last thing done to task 1
      And no line of the history is empty

    @gui-the-history-scrolls
    Scenario: a history longer than its box scrolls, and stops at both ends
      Given I added a task titled "Write the spec"
      And it was worked on 20 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the history has more to show than it has room for
      And the history is at its beginning
      When I scroll down over the history
      Then the history has moved
      And the history stays inside the panel
      When I scroll up over the history
      Then the history is at its beginning
      When I scroll down over the history 99 times
      Then the history is at its end
      And the history stays inside the panel

    @gui-the-bar-beside-the-history-is-dragged
    Scenario: the bar beside the history is taken hold of and moved
      Given I added a task titled "Write the spec"
      And it was worked on 20 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      And I drag the history's bar to the bottom
      Then the history is at its end
      And the history stays inside the panel

  Rule: What the arrows reach, they bring into sight

    Walking past the bottom of a column used to choose a card nobody could
    see: the choice moved and the column stayed where it was. A card chosen
    out of sight is a card whose state, whose title and whose place are all
    unknown to the person who chose it.

    @gui-the-arrows-bring-a-card-into-sight
    Scenario: walking down past what a column shows brings the next card into sight
      Given I added 12 tasks to "backlog"
      When I open the window
      And I press "down" 11 times
      Then what is chosen is in sight
      When I press "up" 11 times
      Then what is chosen is in sight

    @gui-walking-across-brings-the-card-into-sight
    Scenario: walking into a column somebody had scrolled brings its card into sight
      Given I added 12 tasks to "backlog"
      And I added 12 tasks to "todo"
      When I open the window
      And I scroll down over "TODO"
      And I press "right"
      Then what is chosen is in sight

    @gui-the-arrows-bring-a-lane-card-into-sight
    Scenario: walking along a lane past what it shows brings the next card into sight
      Given I added 6 tasks to "watching"
      When I open the window
      And I click the card "Task 1"
      And I press "down" 5 times
      Then what is chosen is in sight
      When I press "up" 5 times
      Then what is chosen is in sight

    @gui-the-arrows-bring-an-objective-into-sight
    Scenario: walking down the objectives brings the next one into sight
      Given I added 12 objectives
      When I open the window
      And I click the tab "OBJECTIVES"
      And I press "down" 11 times
      Then what is chosen is in sight

    @gui-the-objectives-are-a-grid
    Scenario: the objectives are laid out across the board rather than one to a row
      Given I added 12 objectives
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the objectives are laid out more than one to a row
      And every objective is drawn inside the board
      And the objectives all fit without scrolling

    @gui-the-arrows-walk-the-grid
    Scenario: the arrows walk the grid across and down
      Given I added 12 objectives
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Objective 1"
      And I press "right"
      Then what is chosen has moved on one
      When I press "down"
      Then what is chosen has moved down one row
      And what is chosen is in sight

    @gui-the-wheel-scrolls-the-objectives
    Scenario: the wheel moves the objectives, which used to stand still
      Given I added 40 objectives
      When I open the window
      And I click the tab "OBJECTIVES"
      And I scroll down over the objectives
      Then the objectives have moved

  Rule: The two pages are walked with tab

    @gui-tab-walks-the-pages
    Scenario: tab moves to the next page and shift+tab to the one before
      When I open the window
      Then the tab is "TASKS"
      When I press "tab"
      Then the tab is "OBJECTIVES"
      When I press "shift+tab"
      Then the tab is "TASKS"
      When I press "shift+tab"
      Then the tab is "OBJECTIVES"

    @gui-tab-in-a-form-walks-the-fields
    Scenario: tab in a form walks its fields rather than changing the page
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "tab"
      Then the tab is "TASKS"
      And the caret is on the second field

    @gui-a-tab-says-how-much-it-holds
    Scenario: a tab says how much its page holds
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I added an objective titled "Ship by June"
      When I open the window
      Then the tab "TASKS" says it holds 2
      And the tab "OBJECTIVES" says it holds 1

  Rule: An objective opens on the work that serves it

    An objective is a question — is this finished — and the answer is the tasks
    under it. Opening one shows them by name and by where each stands, because
    a count alone says how far along it is and never what is left.

    They sit beside the form, in the column a task's history has: an objective
    has no history of its own — the trace is kept of tasks — and the question
    it is opened to answer is what serves it. Each is a tile of one line,
    scrolled by the wheel over the column and by the bar beside it.

    @gui-an-objective-opens-on-what-serves-it
    Scenario: opening an objective shows the work that serves it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Buy a lamp"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then "Ship the desktop board" is open to be edited
      And the work beside the form carries "Write the spec"
      And the work beside the form does not carry "Buy a lamp"
      And what hangs under it is drawn one line to a tile
      And the column beside the form is headed "WHAT SERVES IT"

    @gui-the-work-under-an-objective-scrolls
    Scenario: more work than the box holds scrolls, and stops at both ends
      Given I added an objective titled "Ship the desktop board"
      And I added 30 tasks serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then the work under it has more to show than it has room for
      And the work under it is at its beginning
      When I scroll down over the work under it
      Then the work under it has moved
      When I scroll up over the work under it
      Then the work under it is at its beginning
      When I scroll down over the work under it 99 times
      Then the work under it is at its end

    @gui-an-objective-card-says-how-much-is-done
    Scenario: an objective says how much work it carries and how much is done
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Ship it" with the state "done" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the card for objective 1 says "1/2"
      And the card for objective 1 says "50%"

    @gui-an-objective-nothing-serves-carries-no-count
    Scenario: an objective nothing serves carries no count at all
      Given I added an objective titled "Ship the desktop board"
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the card for objective 1 says nothing of how far along it is

    @gui-an-open-objective-is-its-own
    Scenario: an open objective is its own, not the task sharing its number
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then "Ship the desktop board" is open to be edited

    @gui-an-objective-with-nothing-under-it-says-so
    Scenario: an objective nothing serves yet says so rather than showing an empty space
      Given I added an objective titled "Ship the desktop board"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then "Ship the desktop board" is open to be edited
      And the work beside the form says nothing serves it yet

  Rule: The window takes work up

    @gui-focus-takes-a-task-up
    Scenario: a task is taken up and let go from the window
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "f"
      Then the engine was asked "task focus" for task 1
      When I press "f"
      Then the engine was asked "task unfocus" for task 1

    @gui-a-parent-is-chosen-from-a-list
    Scenario: the task something hangs under is chosen from a list, not typed as a number
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the window is choosing
      And the panel is headed "a parent for task 2"
      And the list offers "Ship it"

    @gui-a-list-is-corrected-like-any-other-field
    Scenario: what was typed into a list can be taken back out
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I type "shipx"
      Then the list offers nothing
      When I press "backspace"
      Then the list offers "Ship it"

    @gui-enter-takes-and-cmd-enter-keeps
    Scenario: enter takes a row and the list is kept when they are all in
      Given I added an objective titled "Ship cairn"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      Then the bar offers "Take" under "⏎"
      And the bar offers "Keep" under "⌘⏎"
      When I press "enter"
      And I press "down"
      And I press "enter"
      Then the list has taken 2
      When I press "cmd+enter"
      Then the window is on the form
      And the field "objectives" names "Ship cairn"
      And the field "objectives" names "Keep it honest"

    @gui-text-is-held-and-carried
    Scenario: a run of text is held, copied, and put down somewhere else
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "description"
      And I type "Deux colonnes"
      And I press "shift+left" 8 times
      Then the field "description" holds "colonnes"
      When I press "cmd+c"
      Then the clipboard reads "colonnes"
      When I click the field "title"
      And I press "cmd+a"
      And I press "cmd+v"
      Then the field "title" reads "colonnes"

    @gui-a-run-is-held-with-the-mouse
    Scenario: a run of text is held by dragging across it
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I drag across the field "title" from 0 to 5
      Then the field "title" holds "Write"

    @gui-a-run-follows-the-pointer-out-of-its-field
    Scenario: a run keeps following the pointer once it leaves the field
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I drag across the field "title" from 6 to past its end
      Then the field "title" holds "the spec"

    @gui-a-double-click-dragged-holds-whole-words
    Scenario: a double click dragged on holds whole words
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I double-click the field "title" at 2 and drag to 7
      Then the field "title" holds "Write the"

    @gui-a-triple-click-holds-the-paragraph
    Scenario: a triple click holds the whole paragraph it lands in
      Given I added a task titled "Write the spec" with the paragraphs "Deux colonnes et une bande" and "Puis le reste"
      When I open the window
      And I double-click the card "Write the spec"
      And I triple-click the first word of "field:description" on the open task
      And I press "cmd+c"
      Then the clipboard reads "Deux colonnes et une bande"

    @gui-a-held-run-is-typed-over
    Scenario: typing over a held run replaces it, and cutting takes it away
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "title"
      And I press "cmd+a"
      And I press "cmd+x"
      Then the field "title" reads ""
      And the clipboard reads "Write the spec"

    @gui-a-form-can-always-be-saved
    Scenario: a form is saved whatever field the keyboard is on
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I press "enter"
      Then the window is on the form
      And the bar offers "Save" under "⌘⏎"
      And the bar offers "Choose" under "⏎"
      When I press "cmd+enter"
      Then the window is on the board
      And the engine was asked "task edit" for task 2

    @gui-a-parent-is-taken-off-the-way-it-was-taken-on
    Scenario: the task something hangs under can be taken off again
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I press "enter"
      Then the field "parent" reads "1"
      When I click the field "parent"
      Then the window is choosing
      And the list has taken 1
      And the row the keyboard is on is "Ship it"
      When I press "enter"
      Then the window is on the form
      And the field "parent" reads ""

    @gui-a-list-of-tasks-offers-the-worked-on-first
    Scenario: the list offers what is being worked on before what is waiting
      Given I added a task titled "Un jour peut-être"
      And I added a task titled "En cours maintenant" with the state "current"
      And I added a task titled "À faire ensuite" with the state "todo"
      And I added a task titled "Déjà fini" with the state "done"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the list reads "En cours maintenant", "À faire ensuite", "Un jour peut-être", "Déjà fini"

    @gui-two-in-one-column-keep-the-order-the-board-gives-them
    Scenario: two tasks in the same column keep the order the board gives them
      Given I added a task titled "La première" with the state "current"
      And I added a task titled "La seconde" with the state "current"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the list reads "La première", "La seconde"

    @gui-a-list-of-tasks-says-where-each-one-stands
    Scenario: a row says which column its task sits in
      Given I added a task titled "En cours maintenant" with the state "current"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the row "En cours maintenant" says "current"

    @gui-a-space-in-a-list-is-a-space
    Scenario: a space typed into a list is a space, because a title has them
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I type "ship it"
      Then the list offers "Ship it"

    @gui-a-double-click-is-done-with-the-list
    Scenario: a double-click takes a row and is done, which is what a hand expects
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I double-click the list row "Ship it"
      Then the window is on the form
      And the field "parent" names "Ship it"

    @gui-a-long-field-takes-a-line-break
    Scenario: a description is written in paragraphs, not on one line
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "description"
      And I type "Deux colonnes"
      And I press "shift+enter"
      And I type "et une bande"
      Then the window is on the form
      And the field "description" is written on more than one line
      And the field "description" names "Deux colonnes"
      And the field "description" names "et une bande"

    @gui-enter-does-not-save
    Scenario: enter writes a line in the description, and saves nothing
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "description"
      And I type "Deux colonnes"
      And I press "enter"
      And I type "et une bande"
      Then the window is on the form
      And the field "description" is written on more than one line
      And the field "description" names "Deux colonnes"
      And the field "description" names "et une bande"
      And task 1 is called "Write the spec"

    @gui-enter-moves-on-from-a-one-line-field
    Scenario: enter on a field of one line moves on to the next
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "title"
      And I press "enter"
      Then the keyboard is on "description"
      And task 1 is called "Write the spec"

    @gui-a-line-break-takes-the-caret-with-it
    Scenario: the caret moves to the new line, before anything is typed on it
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "description"
      And I type "Deux colonnes"
      And I press "shift+enter"
      Then the caret is on line 2 of the field "description"

    @gui-a-long-list-stays-inside-its-panel
    Scenario: a list of more tasks than fit stays inside the panel and scrolls
      Given I added a task titled "Write the spec"
      And I added 40 tasks
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the list has more to show than it has room for
      And every row of the list is inside the panel
      And a bar sits beside the list
      And the list is at its beginning
      When I scroll down over the list
      Then the list has moved
      And every row of the list is inside the panel
      When I scroll down over the list 99 times
      Then the list is at its end
      And every row of the list is inside the panel

    @gui-walking-past-the-last-row-in-sight-scrolls-the-list
    Scenario: the keyboard brings a row into sight rather than leaving it below
      Given I added a task titled "Write the spec"
      And I added 40 tasks
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I press "down" 30 times
      Then the row the keyboard is on is in sight

    @gui-clicking-a-field-that-a-list-answers-opens-it
    Scenario: clicking the field opens its list, which is what a pointer is for
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      Then the window is choosing
      And the panel is headed "a parent for task 2"

    @gui-a-list-answers-the-mouse
    Scenario: a row is chosen by clicking it, the way everything else here is
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I click the list row "Ship it"
      Then the window is choosing
      When I press "enter"
      Then the window is on the form
      And the field "parent" names "Ship it"

    @gui-a-click-in-a-list-leaves-the-form-under-it-alone
    Scenario: a click inside the list does not reach the form behind it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I click the line that narrows the list
      Then the window is choosing
      And the form is still on the field "parent"

    @gui-the-list-narrows-as-it-is-typed-into
    Scenario: typing in the list keeps only what matches
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      And I added a task titled "Rewrite the tests"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I type "tests"
      Then the list offers "Rewrite the tests"
      And the list does not offer "Ship it"

    @gui-choosing-a-parent-says-which-task-it-is
    Scenario: the field says which task it was attached to, not only its number
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "parent"
      And I press "enter"
      Then the window is on the form
      And the field "parent" names "Ship it"
      And the field "parent" reads "1"

    @gui-a-card-says-it-carries-more-than-it-shows
    Scenario: a card says a description exists and how far its parts have got
      Given I added a task titled "Write the spec" with the description "Two columns"
      And I added a task titled "Ship it"
      And I added a task titled "Write the tests" hanging under task 2
      And I edited task 3 with the state "done"
      When I open the window
      Then the card for task 1 is marked as described
      And the card for task 2 says "1/1"
      And the card for task 2 is not marked as described

    @gui-labels-show-on-the-card
    Scenario: a card says what it is labelled without being opened
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "backend" and "sp26-20"
      When I open the window
      Then the card "Write the spec" says "backend"
      And the card "Write the spec" says "sp26-20"

    @gui-a-card-says-how-many-labels-it-could-not-show
    Scenario: a card too narrow for every label says how many it left out
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "an-altogether-overlong-label" and "another-just-as-long-one"
      When I open the window
      Then the card "Write the spec" says "+1"

    @gui-focus-shows-on-the-card
    Scenario: a card says it is being worked on
      Given I added a task titled "Write the spec"
      And I put task 1 in focus
      When I open the window
      Then the card "Write the spec" says it is being worked on

  Rule: A key does what the bar offers, where the bar offers it

    The bar along the bottom is what somebody reads to find out what a key
    does. A key that stays silent where the bar offers it makes the bar a
    thing to distrust — and one that works where the bar says nothing is a
    thing nobody can find.

    @gui-e-edits-an-open-task
    Scenario: a task being read is edited without closing it first
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then "Write the spec" is open to be edited
      When I press "e"
      Then the form is open on task 1

    @gui-every-key-the-bar-offers-works
    Scenario: every key the bar offers does what the bar says, on every screen
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      Then every key the bar offers does what it offers

    @gui-adding-does-not-throw-away-an-open-form
    Scenario: the key that adds does nothing where adding is not offered
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Write it properly"
      And I press "cmd+n"
      Then the form is open on task 1
      And the field "title" reads "Write it properly"

  Rule: A task opens on the work hanging under it

    A parent is a question the same way an objective is — is this finished —
    and the answer is the tasks under it. They are the last thing its form
    holds: a list to tick while the keyboard is elsewhere, and the lines being
    typed while it is on them.

    They are drawn as tiles of one line and not as the cards a board draws. A
    card is 96 pixels high and uses 50 of them: the rest was room for a
    deadline and a progress bar a sub-task no longer carries. One line holds
    everything one has left — whether it is done, its number, its title.
    What is left comes first, and what is done after it.

    A part is changed through its task's form, with the keyboard on its list:
    the arrows walk the tiles, space ticks the one the keyboard is on, enter
    renames it, backspace strikes it out — and again brings it back — and a
    tile is dragged to another place. The last line adds one: enter, a title,
    enter, and the next is already waiting.

    A tick is done at once: a box ticked on a part says it is done, and a part
    done only once its task's form was saved is a box that lies until then.
    Everything else — a name, a part added, one struck out, the order — waits
    for the form to be saved, and goes with a form thrown away.

    @gui-the-arrows-walk-the-parts
    Scenario: the arrows walk the parts, and space ticks the one they are on
      Given I added a task titled "Ship it"
      And I added a task titled "Draft the columns" hanging under task 1
      And I added a task titled "Get the nod" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      Then the part the keyboard is on is "Draft the columns"
      When I press "down"
      Then the part the keyboard is on is "Get the nod"
      When I press "space"
      Then the part "Get the nod" is ticked
      And task 3 is in the column "done"

    @gui-a-part-is-renamed-in-its-tile
    Scenario: enter renames the part the keyboard is on, in its tile
      Given I added a task titled "Ship it"
      And I added a task titled "Draft" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      And I press "enter"
      And I type " the columns"
      And I press "enter"
      And I choose "Save"
      Then task 2 hangs under task 1, titled "Draft the columns"

    @gui-parts-are-added-one-after-another
    Scenario: the last line adds a part, and the next is waiting behind it
      Given I added a task titled "Ship it"
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      And I press "enter"
      And I type "Draft the columns"
      And I press "enter"
      And I type "Get the nod"
      And I press "enter"
      And I choose "Save"
      Then task 2 hangs under task 1, titled "Draft the columns"
      And task 3 hangs under task 1, titled "Get the nod"

    @gui-a-part-is-struck-out-and-back
    Scenario: backspace strikes a part out until the form is saved, and again brings it back
      Given I added a task titled "Ship it"
      And I added a task titled "Draft" hanging under task 1
      And I added a task titled "Keep" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      And I press "backspace"
      Then the part "Draft" is struck out
      When I press "backspace"
      Then the part "Draft" is not struck out
      When I press "down"
      And I press "backspace"
      And I choose "Save"
      Then the engine was asked "task delete" for task 3
      And task 2 hangs under task 1, titled "Draft"

    @gui-the-wheel-scrolls-the-parts
    Scenario: the wheel scrolls the parts, wherever the keyboard is
      Given a window of 1024 by 640
      And I added a task titled "Ship it"
      And I added 12 tasks hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      Then the parts are at their beginning
      When I scroll down over the parts
      Then the parts have moved
      When I scroll down over the parts 99 times
      Then the last part is in sight

    @gui-a-part-in-hand-shows-where-it-lands
    Scenario: a part in hand follows the pointer, and says where it would land
      Given I added a task titled "Ship it"
      And I added a task titled "First" hanging under task 1
      And I added a task titled "Second" hanging under task 1
      And I added a task titled "Third" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I take hold of the part "First" and hold it over the part "Third"
      Then a part is in hand, under the pointer
      And it would land below "Third"

    @gui-a-part-is-dragged-to-its-place
    Scenario: a part dragged onto another takes its place
      Given I added a task titled "Ship it"
      And I added a task titled "First" hanging under task 1
      And I added a task titled "Second" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I drag the part "Second" onto the part "First"
      Then the parts are drawn as a list to tick, "Second" first
      When I choose "Save"
      Then the engine was asked "task move" for task 3

    @gui-parts-done-and-left-in-turn-are-saved
    Scenario: parts done and left in turn are saved, each kept among its own
      Given I added a task titled "Ship it"
      And I added a task titled "One" hanging under task 1
      And I added a task titled "Two" hanging under task 1
      And I added a task titled "Three" hanging under task 1
      And I added a task titled "Four" hanging under task 1
      And I added a task titled "Five" hanging under task 1
      And I edited task 2 with the state "done"
      And I edited task 3 with the state "done"
      And I edited task 5 with the state "done"
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      And I press "down"
      And I press "space"
      And I choose "Save"
      Then the window is on the board
      And task 6 is in the column "done"

    @gui-a-tick-is-kept-whatever-becomes-of-the-form
    Scenario: a part ticked is done at once, and nothing is left to save for it
      Given I added a task titled "Ship it"
      And I added a task titled "Draft" hanging under task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "parts"
      And I press "space"
      Then task 2 is in the column "done"
      When I press "esc"
      Then the window is on the board
      And task 2 is in the column "done"

    @gui-an-open-task-says-its-title-in-its-heading
    Scenario: an open task says its title on the line beside its number
      Given I added a task titled "Ship it"
      When I open the window
      And I double-click the card "Ship it"
      Then the panel is headed "task 1 · Ship it"

    @gui-every-text-on-an-open-task-is-copied
    Scenario: a word of anything an open task writes is held and copied
      Given I added a task titled "Write the spec" with the description "Deux colonnes et une bande"
      When I open the window
      And I double-click the card "Write the spec"
      And I double-click the first word of "field:description" on the open task
      And I press "cmd+c"
      Then the clipboard reads "Deux"

    @gui-the-parts-left-come-first
    Scenario: what is left comes first
      Given I added a task titled "Ship it"
      And I added a task titled "Draft the columns" hanging under task 1
      And I added a task titled "Get the client's nod" hanging under task 1
      And I edited task 2 with the state "done"
      When I open the window
      And I double-click the card "Ship it"
      Then the parts are drawn as a list to tick, "Get the client's nod" first

  Rule: The objectives a task serves are picked from a list, not typed

    A task serves as many objectives as it needs, and the form asked for them
    as numbers typed by hand. Nobody knows an objective by its number: the one
    place that says what a number means is the other tab, so filling this field
    meant leaving the form to go and look.

    Enter opens the list over the form. Everywhere else in a form enter saves,
    and here it does not: this is the one field that cannot be typed into, so
    the key that would have finished a line opens the thing the line is picked
    from instead. Saving is still on the bar, which is where a key is looked
    for.

    The list is drawn over the form and not instead of it: what is being
    filled in stays on screen behind it, because a choice made with the rest of
    the form out of sight is a choice made from memory.

    The field itself reads the objectives by name. It holds the numbers the
    engine takes — a surface inventing its own spelling of an argument would be
    a surface deciding something — but nobody knows an objective by its number,
    so what is drawn is what a person would say.

    The arrows walk the list, and so do j and k, which is what the board is
    walked with. Space takes an objective in or out — ticking is not the same
    act as finishing, and a list that closed on the first choice could never be
    used to pick two. esc closes it, and what was ticked stays ticked; nothing
    reaches the store until the form is saved, the way nothing else in a form
    does.

    @gui-the-objectives-field-opens-a-list
    Scenario: the objectives field opens a list of what there is to serve
      Given I added an objective titled "Ship the desktop board"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      Then the list of objectives is open
      And the list offers "Ship the desktop board"
      And the list offers "Keep it honest"
      And the form is still drawn behind it

    @gui-space-ticks-an-objective
    Scenario: enter takes an objective in, and takes it out again
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "enter"
      Then objective 1 is ticked
      When I press "enter"
      Then objective 1 is not ticked

    @gui-the-list-is-walked-like-the-board
    Scenario: the arrows walk the list, and so do j and k
      Given I added an objective titled "Ship the desktop board"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      Then the list is on "Ship the desktop board"
      When I press "down"
      Then the list is on "Keep it honest"
      When I press "k"
      Then the list is on "Ship the desktop board"
      When I press "j"
      Then the list is on "Keep it honest"
      When I press "j"
      Then the list is on "Keep it honest"
      When I press "k"
      And I press "k"
      Then the list is on "Ship the desktop board"

    @gui-what-was-taken-is-what-the-field-carries
    Scenario: keeping the list leaves the field carrying what was taken
      Given I added an objective titled "Ship the desktop board"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "enter"
      And I press "down"
      And I press "enter"
      And I press "cmd+enter"
      Then the form is open on task 1
      And the field "objectives" names "Ship the desktop board"
      And the field "objectives" names "Keep it honest"
      And the field "objectives" is drawn by name and not by number

    @gui-backing-out-of-a-list-keeps-nothing
    Scenario: backing out of a list puts nothing in the field, and keeps the form
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "enter"
      And I press "esc"
      Then the form is open on task 1
      And the field "objectives" reads ""

    @gui-a-store-with-no-objectives-says-so
    Scenario: a list with nothing to offer says so rather than opening empty
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      Then the list of objectives is open
      And the list says there is nothing to serve yet
      When I press "space"
      Then the list says there is nothing to serve yet

  Rule: The key that opens a field does not write in it

    A keyboard can say the same keystroke twice: once as a key with a name, and
    once as the text an input method produced from it. Some systems send both,
    some send one — so a window that listens to both hears `e` as "open the
    form" and then as "write an e in it", and the title gains a letter nobody
    typed on purpose. The same keystroke opening a search left every search
    beginning with a slash, which finds nothing.

    What opens a place to type does not also type in it. Text is taken only
    where there was already somewhere for it to go.

    @gui-the-key-that-opens-a-form-writes-nothing-in-it
    Scenario: the key that opens the form does not land in the field
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And the keyboard says "e" as both a key and as text
      Then the form is open on task 1
      And the field "title" reads "Write the spec"

    @gui-the-key-that-opens-a-search-writes-nothing-in-it
    Scenario: the key that opens the search does not land in the box
      Given I added a task titled "Write the spec"
      When I open the window
      And the keyboard says "/" as both a key and as text
      Then the window is asking what to look for
      And nothing has been looked for yet

    @gui-typing-into-an-open-field-still-writes
    Scenario: typing into a field that is already open still writes
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And the keyboard says "s" as both a key and as text
      Then the field "title" reads "Write the specs"

  Rule: Losing work is asked about first

    Two gestures lose work: deleting a task, and throwing away a form somebody
    has typed into. Both are asked about, in a box no bigger than the question.
    A whole column going off the board is asked about as well — nothing is
    lost, but a key that moves twelve cards is not a key to press and find out.

    A form is left with esc, and esc again throws it away: that is what a hand
    leaving a form does, and a question that sent the second esc back to the
    form kept somebody in a form they were leaving. What was typed is not lost
    for all that — it is set aside, and the form of the same task offers to
    bring it back.

    @gui-a-filled-form-asks-before-it-is-thrown-away
    Scenario: a form somebody typed into is asked about before it goes
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Something else"
      And I choose "Cancel"
      Then the panel is headed "Throw away what you typed?"
      And the question is no bigger than what it asks
      When I choose "Back"
      Then the window is on the form

    @gui-esc-twice-sets-what-was-typed-aside
    Scenario: esc twice throws a form away, and its form brings what was typed back
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I set the field "title" to "Write the whole spec"
      And I press "esc"
      And I press "esc"
      Then the window is on the board
      And task 1 is titled "Write the spec"
      When I choose "Edit"
      Then the bar offers "Recover" under "⌘Z"
      When I press "cmd+z"
      Then the field "title" reads "Write the whole spec"
      When I choose "Save"
      Then task 1 is titled "Write the whole spec"

    @gui-an-untouched-form-asks-nothing
    Scenario: a form nobody typed into closes without a question
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I choose "Cancel"
      Then the window is on the board

    @gui-a-whole-column-goes-off-the-board-at-once
    Scenario: the column a card stands in is archived in one gesture
      Given I added a task titled "Write the spec" with the state "done"
      And I added a task titled "Ship it" with the state "done"
      When I open the window
      And I click the card "Write the spec"
      And I press "shift+x"
      Then the panel is headed "Take every task in DONE off the board?"
      When I choose "Agree"
      Then "DONE" does not carry "Write the spec"
      And "DONE" does not carry "Ship it"

  Rule: What is on screen says what it is

    Somebody arriving at this window should not have to guess. Which page they
    are on, whether they are searching, and what a key does are all things the
    window says out loud, because a board that has to be learned once is a
    board that is learned wrong.

    @gui-nothing-runs-out-of-a-field
    Scenario: what is typed into a field stays inside it, on as many lines as it takes
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "description"
      And I type "This description is far longer than one line of this form can hold, so it has to go on somewhere rather than run out of the box it was typed into"
      Then the field "description" is written on more than one line
      And no line of the field "description" runs out of it

    @gui-the-form-keeps-its-offers-on-a-line-of-their-own
    Scenario: the form of a task with parts fits above its offers, even in the smallest window
      Given a window of 1024 by 640
      And I added a task titled "Ship it"
      And I added a task titled "Draft the columns" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I choose "Edit"
      Then no field of the form reaches what it offers
      And no two fields of the form overlap
      And the title runs the width of the form, above everything else
      And the description runs the width of the form, under the title
      And the form groups its fields under "STATE", "TIME", "BLOCKED", "LINKS"

    @gui-a-new-task-goes-in-with-its-parts
    Scenario: a new task is added with the parts already written down for it
      When I open the window
      And I choose "New"
      And I set the field "title" to "Ship it"
      And I set the field "parts" to "Draft the columns"
      And I choose "Save"
      Then task 2 hangs under task 1, titled "Draft the columns"

    @gui-the-parts-read-as-a-list-to-tick
    Scenario: the parts of a task being edited read as a list to tick, written as lines underneath
      Given I added a task titled "Ship it"
      And I added a task titled "Draft the columns" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I choose "Edit"
      Then the parts are drawn as a list to tick, "Draft the columns" among them
      When I click the field "parts"
      Then the field "parts" reads "#2 [ ] Draft the columns"

    @gui-the-bar-teaches-the-key
    Scenario: the bar says what to press, not only what can be done
      When I open the window
      Then the bar offers "New" under "⌘N"
      And the bar offers "Search" under "/"
      And the bar offers "Help" under "?"

    @gui-the-board-keeps-its-bar-under-a-panel
    Scenario: opening a task leaves the bar under the board as it was, and the panel offers its own
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I double-click the card "Write the spec"
      Then the bar under the board reads "New", "Search", "Archived", "Help", "Edit", "Archive", "Delete"
      And the bar under the board answers nothing
      And the bar offers "Save" under "⌘⏎"

    @gui-a-search-says-it-is-a-search
    Scenario: a search says it is one, and the key that opened it is not typed
      When I open the window
      And I press "/"
      Then the window is asking what to look for
      And the search reads ""
      When I type "spec"
      Then the search reads "spec"

    @gui-the-help-is-headed
    Scenario: the help says what it is before it says anything else
      When I open the window
      And I click "Help"
      Then the help is headed "What everything does"

  Rule: Several cards are taken at once

    A board that deletes or archives one card at a time makes a clear-out a
    chore. ⌘-click gathers cards, a card at a time; ⇧-click takes every card
    between the one chosen and the one clicked, in their column; a plain click
    starts again. ⇧ and the arrows already move a card, so the keyboard gathers
    nothing — the pointer does.

    What is gathered is deleted or archived in one act, all of it or none, and
    asked about first. The question names what would take more with it — the
    parts of a task, a task waiting on it, the task in focus — and what cannot
    go at all — an objective work still serves, a task already off the board.
    The answer is one of three: nothing; all but what was named; or all that
    can go.

    @gui-several-are-gathered-with-cmd
    Scenario: ⌘-clicking gathers cards, and a plain click starts again
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I open the window
      And I click the card "First"
      And I ⌘-click the card "Third"
      Then the cards taken together are "First", "Third"
      When I click the card "Second"
      Then no card is taken with another

    @gui-a-run-is-gathered-with-shift
    Scenario: ⇧-clicking takes every card between, in their column
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I open the window
      And I click the card "First"
      And I ⇧-click the card "Third"
      Then the cards taken together are "First", "Second", "Third"

    @gui-several-are-deleted-at-once
    Scenario: several tasks are deleted in one act, once asked
      Given I added a task titled "First"
      And I added a task titled "Second"
      And I added a task titled "Third"
      When I open the window
      And I click the card "First"
      And I ⌘-click the card "Third"
      And I press "d"
      Then the panel is headed "Delete 2 tasks?"
      When I choose "All possible"
      Then "BACKLOG" does not carry "First"
      And "BACKLOG" does not carry "Third"
      And "BACKLOG" carries "Second"

    @gui-what-takes-more-can-be-left-out
    Scenario: what would take more with it is named, and can be left out
      Given I added a task titled "Ship it"
      And I added a task titled "Draft the columns" hanging under task 1
      And I added a task titled "Other"
      When I open the window
      And I click the card "Ship it"
      And I ⌘-click the card "Other"
      And I press "d"
      Then the question says of task 1 "its 1 part goes with it"
      When I choose "Skip flagged"
      Then "BACKLOG" carries "Ship it"
      And "BACKLOG" does not carry "Other"

    @gui-nothing-goes-when-nothing-is-chosen
    Scenario: nothing goes when the answer is nothing
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I open the window
      And I click the card "First"
      And I ⌘-click the card "Second"
      And I press "d"
      And I choose "Cancel"
      Then "BACKLOG" carries "First"
      And "BACKLOG" carries "Second"

    @gui-several-are-archived-at-once
    Scenario: several tasks are archived in one act, once asked
      Given I added a task titled "First"
      And I added a task titled "Second"
      When I open the window
      And I click the card "First"
      And I ⌘-click the card "Second"
      And I press "x"
      Then the panel is headed "Archive 2 tasks?"
      When I choose "All possible"
      Then "BACKLOG" does not carry "First"
      And "BACKLOG" does not carry "Second"

    @gui-what-cannot-go-is-left-out
    Scenario: an objective work still serves is left out, and the rest goes
      Given I added an objective titled "Ship cairn"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      And I ⌘-click the objective "Keep it honest"
      And I press "d"
      Then the question says of objective 1 "1 task serves it: it cannot go"
      When I choose "All possible"
      Then the objectives read "Ship cairn"
