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

  Rule: A click outside a panel closes it, where closing loses nothing

    A hand on the mouse should not have to reach for the keyboard to put
    something away. Clicking off a panel does what esc does — the detail, the
    help, the question about deleting, the list of objectives — because in each
    of those the way out costs nothing.

    A form is the exception, and the only one. It holds what somebody has typed
    and not saved, and a click landing anywhere outside it would throw that
    away without asking. Losing work without being asked is how work is lost;
    esc is still there, and it is offered on the bar as Cancel.

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
      Then the detail is open on "Write the spec"

    @gui-clicking-off-the-detail-closes-it
    Scenario: clicking off an open task puts it away
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      And I click outside the panel
      Then the detail is closed

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
      Then the detail is closed
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
    Scenario: double-clicking a card opens its detail
      Given I added a task titled "Write the spec" with the description "Two columns"
      When I open the window
      And I double-click the card "Write the spec"
      Then the detail is open on "Write the spec"
      And the detail shows "Two columns"

    @gui-the-detail-shows-the-whole-task
    Scenario: the detail shows everything a task carries, not only what was done to it
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I double-click the card "Write the spec"
      Then the detail shows the field "column" reading "todo"
      And the detail shows the field "priority"
      And the detail shows the field "labels"

    @gui-the-detail-is-closed-by-clicking
    Scenario: the detail is closed without knowing a key
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I click "Close"
      Then the detail is closed

    @gui-the-detail-closes
    Scenario: the detail closes and leaves the board as it was
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I press "esc"
      Then the detail is closed
      And the columns read "BACKLOG", "TODO", "CURRENT", "DONE"

    @gui-the-wheel-scrolls-one-column
    Scenario: the wheel scrolls the column under the pointer and no other
      Given I added 12 tasks to "backlog"
      And I added 12 tasks to "todo"
      When I open the window
      And I scroll down over "BACKLOG"
      Then "BACKLOG" is scrolled
      And "TODO" is not scrolled

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
      Then the detail is open on "Write the spec"

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

    @gui-clicking-a-field
    Scenario: clicking a field is how the keyboard gets there
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "priority"
      Then the keyboard is on "priority"

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
      And I click "Delete"
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
      Then the detail is open on "Write the spec"

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
      Then the detail is open on "Write the spec"

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

    @gui-the-arrows-scroll-the-history
    Scenario: the arrows scroll the history while a task is open
      Given I added a task titled "Write the spec"
      And it was worked on 20 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      And I press "down"
      Then the history has moved
      When I press "up"
      Then the history is at its beginning

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

    They sit under the fields, in what the left column has left, and they are
    cards — the same cards the board draws, at the same height, scrolled the
    same way. A run of text lines would make the one place that lists tasks the
    one place a task does not look like a task.

    The right column stays the history: it is what a task opens onto, and
    moving it aside for an objective would cost every task the thing it was
    built for.

    @gui-an-objective-opens-on-what-serves-it
    Scenario: opening an objective shows the work that serves it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Buy a lamp"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then the detail is open on "Ship the desktop board"
      And the detail draws a card for "Write the spec" under its fields
      And the detail draws no card for "Buy a lamp"
      And what hangs under it is drawn one line to a tile
      And the right column is still headed "HISTORY"

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

    @gui-the-arrows-scroll-the-work-under-an-objective
    Scenario: the arrows walk the work, and the box follows them out of sight
      Given I added an objective titled "Ship the desktop board"
      And I added 30 tasks serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then the work under it is at its beginning
      When I press "down" 20 times
      Then the work under it has moved
      When I press "up" 20 times
      Then the work under it is at its beginning

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

    @gui-an-objective-detail-is-its-own
    Scenario: an objective's detail is its own, not that of the task sharing its number
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then the detail is open on "Ship the desktop board"

    @gui-an-objective-with-nothing-under-it-says-so
    Scenario: an objective nothing serves yet says so rather than showing an empty space
      Given I added an objective titled "Ship the desktop board"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the desktop board"
      And I press "enter"
      Then the detail is open on "Ship the desktop board"
      And the detail says nothing serves it yet

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

    @gui-an-open-task-is-headed-by-its-whole-title
    Scenario: an open task says its whole title on the line beside its number
      Given I added a task titled "install-argonos-root-ca : la sonde teste /readyz au lieu du quorum etcd"
      When I open the window
      And I click the card "install-argonos-root-ca : la sonde teste /readyz au lieu du quorum etcd"
      And I press "enter"
      Then the panel is headed "task 1 · install-argonos-root-ca : la sonde teste /readyz au lieu du quorum etcd"

    @gui-an-open-objective-is-headed-by-its-title-too
    Scenario: an objective says its title there as well
      Given I added an objective titled "Ship cairn"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      And I press "enter"
      Then the panel is headed "objective 1 · Ship cairn"

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
      Then the detail is open on "Write the spec"
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
    and the answer is the tasks under it. They sit in the box under the fields,
    because it is the same question asked of a different thing.

    They are drawn as tiles of one line and not as the cards a board draws. A
    card is 96 pixels high and uses 50 of them: the rest was room for a
    deadline and a progress bar a sub-task no longer carries. One line holds
    everything one has left — whether it is done, its number, its title — and
    fits five in the box where a card fit one.

    The strip down the left says done or not, where a card's says its priority.
    It is the only fact a part carries that a glance should not have to read
    the words for.

    A title too long for its tile is cut with an ellipsis, measured rather than
    reckoned: the font is proportional, so counting an average glyph width left
    a title that ran to the frame's edge and was clipped there, against the bar
    beside it.

    The arrows walk the tiles, and space marks the one they are on done or not
    done. They used to scroll the history: the box takes them because it is
    what somebody acts on, and a line of history is read rather than changed.
    The history keeps the wheel over it and the bar beside it.

    The bar says so. `Done` appears on it while a task with parts is open, the
    way `Edit` and `Close` do, because the bar is where a key is looked for —
    and the arrows are in the help, where the board's own arrows are, since the
    bar carries what changes a task and not how to get about.

    Nothing reaches the store until space is pressed, and then at once — this
    is not a form, and a tile is one fact somebody is either changing or not.

    A tile is dragged to reorder it. Solidary with its parent does not mean
    fixed among its sisters: a part can be wanted before another, and the order
    is what says which. It moves within its box and nowhere else — there is no
    column to land in, and its parent is where it belongs.

    Dragged the way everything on this window is dragged: the same hold, the
    same ghost following the pointer, the same line where it would land. What
    changes is where it may land, not how a hand does it. A second mechanism
    written beside the first drew no ghost and no line, so a drop was aimed by
    hope — and it left the board's own landing line drawn underneath the panel,
    because the board was still being asked where a point over it would land.

    A task nothing hangs under says nothing there: an empty frame reads as a
    box that failed to load, and most tasks have no sub-tasks at all.

    @gui-a-task-opens-on-what-hangs-under-it
    Scenario: opening a parent shows the work hanging under it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" hanging under task 1
      And I added a task titled "Something else"
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the detail is open on "Ship it"
      And the detail draws a card for "Write the spec" under its fields
      And the detail draws no card for "Something else"
      And the right column is still headed "HISTORY"

    @gui-a-part-in-hand-shows-where-it-would-land
    Scenario: a tile in hand follows the pointer and says where it would go
      Given I added a task titled "Ship it"
      And I added a task titled "First" hanging under task 1
      And I added a task titled "Second" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      And I take hold of the tile "Second" and move it over the tile "First"
      Then the tile being dragged follows the pointer
      And what is in hand says what it carries
      And it shows where the tile would land
      And the board behind shows no landing

    @gui-a-part-is-dragged-to-reorder-it
    Scenario: dragging a tile onto another puts it there
      Given I added a task titled "Ship it"
      And I added a task titled "First" hanging under task 1
      And I added a task titled "Second" hanging under task 1
      And I added a task titled "Third" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      And I drag the tile "Third" onto the tile "First"
      Then the engine was asked "task move" for task 4
      And what hangs under it reads "Third", "First", "Second"

    @gui-a-part-dragged-nowhere-moves-nothing
    Scenario: a tile let go outside its box asks nothing
      Given I added a task titled "Ship it"
      And I added a task titled "First" hanging under task 1
      And I added a task titled "Second" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      And I drag the tile "Second" out of its box
      Then the engine was not asked "task move"
      And what hangs under it reads "First", "Second"

    @gui-a-long-part-is-cut-before-the-edge
    Scenario: a title too long for its tile stops before the frame
      Given I added a task titled "Ship it"
      And I added a task titled "Court" hanging under task 1
      And I added a task titled "T090f · Ajouter une étape pour réouvrir (puis refermer) les droits avant intervention" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the second tile says what it holds
      And the second tile leaves room before its frame

    @gui-the-arrows-walk-what-hangs-under-a-task
    Scenario: the arrows walk the tiles, and stop at both ends
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" hanging under task 1
      And I added a task titled "Draw the band" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the tile the keyboard is on is "Write the spec"
      And the tile the keyboard is on is lit
      When I press "down"
      Then the tile the keyboard is on is "Draw the band"
      And the tile the keyboard is on is lit
      When I press "down"
      Then the tile the keyboard is on is "Draw the band"
      When I press "up"
      Then the tile the keyboard is on is "Write the spec"
      When I press "up"
      Then the tile the keyboard is on is "Write the spec"

    @gui-the-bar-offers-marking-a-part-done
    Scenario: the bar says what space does, where space does it
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the bar offers "Done" under "space"

    @gui-the-bar-offers-nothing-to-mark-where-nothing-hangs
    Scenario: a task nothing hangs under is offered nothing to mark
      Given I added a task titled "Ship it"
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the bar does not offer "Done"

    @gui-the-help-says-what-the-arrows-do-in-the-box
    Scenario: the help says what the arrows walk
      When I open the window
      And I press "?"
      Then the help says what "↑↓ under a task" does

    @gui-space-marks-a-part-done
    Scenario: space marks the part the keyboard is on done, and undone again
      Given I added a task titled "Ship it"
      And I added a task titled "Write the spec" hanging under task 1
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      And I press "space"
      Then the engine was asked "task edit" for task 2
      And task 2 is in the column "done"
      When I press "space"
      Then task 2 is in the column "todo"

    @gui-space-does-nothing-where-nothing-hangs
    Scenario: space asks nothing of a task nothing hangs under
      Given I added a task titled "Ship it"
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      And I press "space"
      Then nothing was asked of the engine

    @gui-a-task-with-nothing-under-it-says-nothing
    Scenario: a task nothing hangs under says nothing where the list would be
      Given I added a task titled "Ship it"
      When I open the window
      And I click the card "Ship it"
      And I press "enter"
      Then the detail is open on "Ship it"
      And the box under the fields is empty

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
    Scenario: space takes an objective in, and takes it out again
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "space"
      Then objective 1 is ticked
      When I press "space"
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

    @gui-what-was-ticked-is-what-the-field-carries
    Scenario: closing the list leaves the field carrying what was ticked
      Given I added an objective titled "Ship the desktop board"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "space"
      And I press "down"
      And I press "space"
      And I press "esc"
      Then the form is open on task 1
      And the field "objectives" names "Ship the desktop board"
      And the field "objectives" names "Keep it honest"
      And the field "objectives" is drawn by name and not by number

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

    @gui-the-bar-teaches-the-key
    Scenario: the bar says what to press, not only what can be done
      When I open the window
      Then the bar offers "New" under "⌘N"
      And the bar offers "Search" under "/"
      And the bar offers "Help" under "?"

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
