Feature: The desktop board

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a window of 1200 by 800

  Rule: The work is laid out the way the terminal lays it out

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

    @gui-a-card-is-archived
    Scenario: a chosen card is archived, and unarchived
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      When I open the window
      And I click the card "Migration Jenkins"
      Then the bar offers "Archive" under "x"
      When I choose "Archive"
      Then the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      And I click the card "Migration Jenkins"
      Then the bar offers "Unarchive" under "x"
      When I choose "Unarchive"
      Then "BACKLOG" carries "Migration Jenkins"
      And the card "Migration Jenkins" is not dimmed

    @gui-a-step-is-not-on-the-board
    Scenario: a step does not stand in a column of its own
      Given I added a task titled "Ship it"
      And I added the step "Write the spec" to task 1
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
    Scenario: there are three tabs and the work is the first
      When I open the window
      Then the tabs read "TASKS", "OBJECTIVES", "TIME"
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

    @gui-a-cut-title-is-read-whole-under-the-pointer
    Scenario: a card whose title is cut shows it whole while the pointer rests on it
      Given I added a task titled "A title far longer than any column could ever hope to show in full, even over two lines of a card"
      When I open the window
      And I rest the pointer on the card "A title far longer than any column could ever hope to show in full, even over two lines of a card"
      Then a bubble shows "A title far longer than any column could ever hope to show in full, even over two lines of a card"

    @gui-a-title-that-fits-has-no-bubble
    Scenario: a card whose title fits shows no bubble
      Given I added a task titled "Write the spec"
      When I open the window
      And I rest the pointer on the card "Write the spec"
      Then no bubble shows

  Rule: Each column and the band stand on a ground of their own

    @gui-columns-stand-apart
    Scenario: each column is drawn on a ground of its own, edged, apart from the window's
      When I open the window
      Then every column stands on a ground of its own
      And the band stands on a ground of its own

  Rule: A window too small says what it needs

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
    Scenario: clicking off a form keeps what was typed, saved, and puts it away
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type "X"
      And I click outside the panel
      Then the window is on the board
      And task 1 is called "Write the specX"

    @gui-clicking-off-a-new-task-saves-nothing
    Scenario: clicking off a task being added is leaving it unsaved: it is not made, and is kept as a draft
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And I click outside the panel
      Then the window is on the board
      And there are no tasks
      When I press "cmd+n"
      Then the window asks to pick up the task left unsaved

    @gui-a-dead-key-composes-one-letter
    Scenario: a dead key and the letter after it are written as the one letter they make
      When I open the window
      And I press "cmd+n"
      And I type "Fen"
      And I compose "ê" from the dead key "^" and the letter "e"
      And I type "tre"
      Then the field "title" reads "Fenêtre"

    @gui-shift-held-still-deletes
    Scenario: holding shift down while deleting still deletes
      When I open the window
      And I press "cmd+n"
      And I type "Plans"
      And I press "backspace" holding shift
      Then the field "title" reads "Plan"

    @gui-a-click-off-the-calendar-closes-it
    Scenario: a click on the form outside the calendar closes the calendar alone
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-12-25"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      And I click the field "title"
      Then the window is on the form
      And the field "deadline" reads "2026-12-25"

    @gui-a-click-off-the-form-closes-the-calendar-too
    Scenario: a click outside the form closes the calendar and the form
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      And I click outside the panel
      Then the window is on the board

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

    @gui-the-wheel-stays-inside-a-panel
    Scenario: the wheel over a panel never reaches the board behind it
      Given I added 12 tasks to "backlog"
      When I open the window
      And I press "cmd+b"
      And I press "shift+b"
      And I scroll down over "BACKLOG"
      Then "BACKLOG" is not scrolled

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
      Then the engine was asked "task move" for task 1
      And the engine was not asked "task edit"
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

    @gui-cards-taken-together-are-dragged-together
    Scenario: dragging one of the cards taken together brings the others along
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "backlog"
      And I added a task titled "Third" with the state "backlog"
      When I open the window
      And I click the card "Second"
      And I ⌘-click the card "Third"
      And I drag the card "Second" onto "TODO"
      Then "TODO" reads "Second", "Third", "First"
      And "BACKLOG" does not carry "Third"

    Scenario: a card dropped under the last of its column goes last
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Third" with the state "todo"
      When I open the window
      And I drag the card "First" to the foot of "TODO"
      Then "TODO" reads "Second", "Third", "First"

    @gui-a-drag-reaches-the-foot-of-a-long-column
    Scenario: a card dragged over a column longer than the window scrolls it, and goes last
      Given I added 20 tasks to "todo"
      And I added a task titled "Moved" with the state "backlog"
      When I open the window
      And I take hold of the card "Moved" and move it over "TODO"
      And I scroll down over "TODO" 99 times
      And I let go at the foot of "TODO"
      Then the last card of "TODO" is "Moved"

    @gui-a-drop-in-a-lane-says-which-kind
    Scenario: dropping a card in a lane of the band says which kind of stopped
      Given I added a task titled "Client's answer" with the state "todo"
      When I open the window
      And I drag the card "Client's answer" onto the lane "watching"
      Then task 1 is in the column "watching"
      And the lane "watching" carries "Client's answer"

    @gui-a-drop-in-done-asks-about-open-steps
    Scenario: a card dropped in DONE with steps still open asks to tick them too
      Given I added a task titled "Ship it" with the state "todo"
      And I added the step "Write the spec" to task 1
      When I open the window
      And I drag the card "Ship it" onto "DONE"
      Then the window asks to put task 1 in done, ticking its steps
      When I click "Agree"
      Then task 1 is in the column "done"
      And every step of task 1 is done

    @gui-a-drop-in-done-can-be-called-off
    Scenario: calling off the question leaves the card and its steps as they were
      Given I added a task titled "Ship it" with the state "todo"
      And I added the step "Write the spec" to task 1
      When I open the window
      And I drag the card "Ship it" onto "DONE"
      And I click "Cancel"
      Then task 1 is in the column "todo"
      And no step of task 1 is done

    Scenario: a drop the engine refuses leaves the board saying what is really there
      Given I added a task titled "Write the spec" with the state "backlog"
      And the store refuses what it is asked to change
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      Then it says what went wrong
      And "BACKLOG" carries "Write the spec"

    @gui-what-went-wrong-leaves-the-bar
    Scenario: what went wrong is said above the bar, and the bar still offers
      Given I added a task titled "Write the spec" with the state "backlog"
      And the store refuses what it is asked to change
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      Then it says what went wrong
      And what went wrong is drawn clear of the bar

    @gui-what-went-wrong-stays-on-its-page
    Scenario: what went wrong on one page is not carried to the other
      Given I added a task titled "Write the spec" with the state "backlog"
      And the store refuses what it is asked to change
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      And I press "tab"
      Then the tab is "OBJECTIVES"
      And it says nothing went wrong

    @gui-what-went-wrong-fades
    Scenario: what went wrong goes on its own once it has been read
      Given I added a task titled "Write the spec" with the state "backlog"
      And the store refuses what it is asked to change
      When I open the window
      And I drag the card "Write the spec" onto "TODO"
      And the window has looked at the store 3 times
      Then it says what went wrong
      When the window has looked at the store 1 time
      Then it says nothing went wrong

  Rule: A card says when it is wanted and what it has cost

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

    @gui-a-finished-card-says-nothing-of-its-day
    Scenario: a task that is done says nothing of the day it was wanted
      Given I added a task titled "Shipped" with the state "done"
      And I edited task 1 with the deadline "2026-09-20"
      When I open the window
      Then "DONE" carries "Shipped"
      And the card "Shipped" says nothing of "9d"

    @gui-a-card-says-what-it-has-cost
    Scenario: what a task has cost is written beside what it was reckoned to cost
      Given I added a task titled "Write the spec" with the estimate "2h"
      When I open the window
      Then the card "Write the spec" says "2h"

  Rule: A form is the whole of a task

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
      And the form asks about "estimate", "time spent", "links", "objectives"

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
    Scenario: a column is chosen from a list, the way an objective is
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

    @gui-a-list-is-the-size-of-what-it-holds
    Scenario: a field's list hangs under it, as long as what it offers
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "priority"
      Then the window is choosing
      And the list hangs under the field "priority"
      And the list has room for 4 rows and no more

    @gui-a-priority-is-found-by-typing
    Scenario: a priority is found by typing, the way an objective is
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "priority"
      And I type "hi"
      Then the list reads "high"
      When I press "enter"
      Then the field "priority" reads "high"

    @gui-a-closed-set-takes-no-free-text
    Scenario: typing on a field with a closed set opens its list, narrowed
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I press "tab"
      And I press "tab"
      And I press "tab"
      Then the form is still on the field "priority"
      When I type "h"
      Then the window is choosing
      And the list reads "high"
      When I press "esc"
      Then the field "priority" reads "normal"

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
      And I click the field "estimate"
      Then the keyboard is on "estimate"

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

    @gui-the-caret-goes-to-the-ends-of-its-line
    Scenario: the caret goes to the start and the end of the line it is on
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "home"
      Then the caret is at 0
      When I press "cmd+right"
      Then the caret is at 14
      When I press "cmd+left"
      Then the caret is at 0
      When I press "end"
      Then the caret is at 14

    @gui-the-caret-walks-a-word-at-a-time
    Scenario: the caret walks a word at a time, and holds as it goes with shift
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "alt+left"
      Then the caret is at 10
      When I press "alt+shift+left"
      And I press "backspace"
      Then the field "title" reads "Write spec"

    @gui-holding-to-the-ends-of-a-line
    Scenario: shift holds from the caret to either end of the line
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "alt+left"
      And I press "cmd+shift+left"
      And I press "backspace"
      Then the field "title" reads "spec"
      When I press "shift+end"
      And I type "plan"
      Then the field "title" reads "plan"

    @gui-taking-out-forward-and-in-strokes
    Scenario: what is after the caret is taken out, and a word or a line in one stroke
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I press "home"
      And I press "delete-forward"
      Then the field "title" reads "rite the spec"
      When I press "end"
      And I press "alt+backspace"
      Then the field "title" reads "rite the "
      When I press "cmd+backspace"
      Then the field "title" reads ""

    @gui-holding-lines-of-a-description
    Scenario: shift with the arrows holds whole lines of a description, which go in one stroke
      Given I added a task titled "Write the spec" with the paragraphs "one" and "two three"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "description"
      And I press "left"
      And I press "left"
      And I press "shift+up"
      And I press "backspace"
      Then the field "description" reads "oneee"

    @gui-a-run-held-across-lines-goes-in-one-stroke
    Scenario: a run held across the lines of a description is taken out whole
      Given I added a task titled "Write the spec" with the paragraphs "one" and "two"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I drag across the field "description" from line 0 character 1 to line 1 character 2
      And I press "backspace"
      Then the field "description" reads "oo"

    @gui-the-search-walks-like-a-field
    Scenario: the search line walks and takes out like any other text
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "/"
      And I type "the spec"
      And I press "home"
      And I press "delete-forward"
      And I press "end"
      And I press "alt+backspace"
      Then the search reads "he "

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

    @gui-a-field-says-what-it-is
    Scenario: the pointer resting on a field's name says what the field is
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I rest the pointer on the name of the field "estimate"
      Then the window explains "how long it should take"

    @gui-a-field-shows-what-to-type
    Scenario: the explanation of a field gives an example of what to type in it
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I rest the pointer on the name of the field "deadline"
      Then the window explains "the day it is due"
      And the explanation gives the example "2026-12-25"
      When I rest the pointer on the name of the field "estimate"
      Then the explanation gives the example "1h30m"

    @gui-an-explanation-is-the-size-of-what-it-says
    Scenario: an explanation is no bigger than what it says
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I rest the pointer on the name of the field "deadline"
      Then the explanation is no bigger than what it says
      When I rest the pointer on the name of the field "time spent"
      Then the explanation is no bigger than what it says
      And the explanation stays inside the window

    @gui-a-field-stops-explaining
    Scenario: the pointer leaving a field's name takes its explanation away
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I rest the pointer on the name of the field "time spent"
      And I rest the pointer on the name of the field "deadline"
      Then the window explains "the day it is due"
      When the pointer leaves the window
      Then the window explains nothing

    @gui-adding-starts-empty
    Scenario: adding starts on an empty form and writes a new task
      When I open the window
      And I press "cmd+n"
      And I set the field "title" to "Write the spec"
      And I choose "Save"
      Then "BACKLOG" carries "Write the spec"

  Rule: The window opens on its board being built of stones

    @gui-the-window-opens-on-a-cairn
    Scenario: the first frame starts a cairn, which a key ends without being swallowed
      Given I added a task titled "Write the spec"
      When I open the window
      And the window draws its first frame
      Then a cairn is being built
      When I press "?"
      Then no cairn is being built
      And the help is open

    @gui-no-task-no-cairn
    Scenario: a board with no task opens on nothing being built
      When I open the window
      And the window draws its first frame
      Then no cairn is being built

    @gui-the-cairn-is-built-in-a-breath
    Scenario: the cairn is done on its own a breath after the window opens
      Given I added a task titled "Write the spec"
      When I open the window
      And the window draws its first frame
      And the window draws a frame 2 seconds later
      Then no cairn is being built

  Rule: The window carries cairn's logo

    @gui-the-window-carries-the-logo
    Scenario: the window hands the system cairn's logo, drawn large, its corners clear
      When I open the window
      Then the window is known to the system as "cairn"
      And the icon it hands the system is cairn's logo at 1024 pixels, its corners clear

    @gui-x11-is-handed-the-logo
    Scenario: on X11 the window hands its logo over itself, at the sizes a bar asks for
      When I open the window
      Then the icon it hands X11 holds cairn's logo at 16, 32, 48, 64 and 128 pixels, its corners clear

    @gui-a-linux-desktop-is-given-an-entry
    Scenario: on a Linux desktop the window writes the entry its logo is found through
      When the window settles in on a Linux desktop
      Then the desktop entry "cairn" names the icon "cairn"
      And the icon "cairn" is there at 16, 32, 48, 64, 128, 256 and 512 pixels

    @gui-an-entry-starts-cairn-by-the-path-it-is-found-at
    Scenario: the entry starts the cairn on the PATH by its full path, which a desktop's own PATH need not reach
      Given the cairn running is the one the PATH leads to as "bin/cairn"
      When the window settles in on a Linux desktop
      Then the desktop entry "cairn" starts the full path of "bin/cairn" with "gui"

    @gui-an-entry-starts-a-cairn-off-the-path-where-it-is
    Scenario: a cairn the PATH does not lead to is started where it is
      Given the cairn running is not the one the PATH leads to
      When the window settles in on a Linux desktop
      Then the desktop entry "cairn" starts the cairn running with "gui"

    @gui-an-entry-hiding-the-logo-is-replaced-once-agreed
    Scenario: an entry somebody else wrote that hides the logo is replaced, once agreed
      Given a desktop entry "cairn" written by somebody else, showing the icon "view-list-details"
      When the window opens on a Linux desktop
      Then the window asks "Show cairn's logo on this desktop?"
      When I press "enter"
      Then the desktop entry "cairn" names the icon "cairn"
      And the window is on the board

    @gui-an-entry-kept-is-asked-about-again-only-once-it-changes
    Scenario: an entry kept as it is is asked about again only once it has changed
      Given a desktop entry "cairn" written by somebody else, showing the icon "view-list-details"
      When the window opens on a Linux desktop
      And I press "esc"
      Then the desktop entry "cairn" still shows the icon "view-list-details"
      When the window opens on a Linux desktop
      Then the window is on the board
      When somebody changes the desktop entry "cairn"
      And the window opens on a Linux desktop
      Then the window asks "Show cairn's logo on this desktop?"

    @gui-an-entry-showing-the-logo-asks-nothing
    Scenario: an entry somebody else wrote that shows cairn's logo is left without a word
      Given a desktop entry "cairn" written by somebody else, showing the icon "cairn"
      When the window opens on a Linux desktop
      Then the window is on the board

    @gui-an-entry-of-somebody-elses-is-kept
    Scenario: an entry cairn did not write is left as it was
      Given a desktop entry "cairn" written by somebody else
      When the window settles in on a Linux desktop
      Then the desktop entry "cairn" is the one somebody else wrote

  Rule: The time spent is seen a week or a month at a time, every task of it

    @gui-a-day-scrolls-to-its-last-task
    Scenario: a day worked on more tasks than its column holds scrolls to the last
      Given I added 20 tasks to "backlog"
      And I spent "15m" on each of tasks 1 to 20 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      Then the day "2026-09-29" has more to show than it has room for
      When I scroll down over the day "2026-09-29" 99 times
      Then the day "2026-09-29" shows "Task 20"
      And the day "2026-09-29" does not show "Task 1"

    @gui-the-time-page-shows-cards
    Scenario: a task on the time page is the card the board shows, with its time that day
      Given I added a task titled "Write the spec" with the estimate "3h"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      Then the day "2026-09-29" shows the card "Write the spec" for "1h30m"

    @gui-a-card-of-the-week-is-chosen
    Scenario: a click on a card of the time page chooses it, like on the board
      Given I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I click the card "Write the spec" on the time page
      Then the card "Write the spec" is chosen

    @gui-a-card-of-the-week-opens
    Scenario: a double-click on a card of the time page opens it, like on the board
      Given I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I double-click the card "Write the spec" on the time page
      Then "Write the spec" is open to be edited

    @gui-a-card-of-the-month-opens
    Scenario: a double-click on a task in a day of the month opens it
      Given I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I click "Month" on the time page
      And I double-click "Write the spec" in the month's day "2026-09-29"
      Then "Write the spec" is open to be edited

    @gui-the-time-page-shows-a-month
    Scenario: the time page shows a month, every day with what was worked on
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I spent "1h" on task 1 on "2026-09-03"
      And I spent "2h" on task 2 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I click "Month" on the time page
      Then the month shown is "September 2026"
      And the month's day "2026-09-03" shows "Write the spec"
      And the month's day "2026-09-29" shows "Ship the board"
      When I click the month's day "2026-09-03"
      Then the week shown starts on "2026-08-31"

    @gui-the-time-page-is-walked-by-the-pointer
    Scenario: the arrows at the top of the time page, and Today, walk what it shows
      Given today is "2026-09-16"
      When I open the window
      And I click the tab "TIME"
      And I click "‹" on the time page
      Then the week shown starts on "2026-09-07"
      When I click "Today" on the time page
      Then the week shown starts on "2026-09-14"
      When I click "Month" on the time page
      And I click "›" on the time page
      Then the month shown is "October 2026"
      When I press "left"
      Then the month shown is "September 2026"
      When I click "Week" on the time page
      Then the week shown starts on "2026-09-14"

    @gui-the-week-of-time
    Scenario: the time page shows the week's days, what was worked on and for how long
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And I spent "45m" on task 2 on "2026-09-29"
      And I spent "2h" on task 2 on "2026-10-02"
      And today is "2026-10-03"
      When I open the window
      And I press "tab"
      And I press "tab"
      Then the tab is "TIME"
      And the week shown starts on "2026-09-28"
      And the day "2026-09-29" says "2h15m"
      And the day "2026-09-29" carries "Write the spec" for "1h30m"
      And the day "2026-10-02" carries "Ship the board" for "2h"
      And the day "2026-09-28" carries nothing

    @gui-the-weeks-are-walked
    Scenario: the arrows walk the weeks, and t comes back to this one
      Given today is "2026-10-14"
      When I open the window
      And I press "tab"
      And I press "tab"
      And I press "left"
      Then the week shown starts on "2026-10-05"
      When I press "right"
      And I press "right"
      Then the week shown starts on "2026-10-19"
      When I press "t"
      Then the week shown starts on "2026-10-12"

    @gui-the-time-tab-is-clicked
    Scenario: the time page is reached by clicking its tab, like the other two
      Given today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      Then the tab is "TIME"
      And the week shown starts on "2026-09-28"

    @gui-a-page-lets-go-of-what-was-taken
    Scenario: changing page lets go of the cards taken together, which are not on the new one
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      When I open the window
      And I click the card "First"
      And I ⌘-click the card "Second"
      And I press "tab"
      And I press "tab"
      Then no card is taken with another
      And the bar does not offer "Delete"

    @gui-the-time-page-chooses-no-card
    Scenario: a click on the time page chooses no card behind it
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I press "tab"
      And I press "tab"
      And I click the column "TODO"
      Then no card is chosen

  Rule: The time spent is seen a day at a time too, and the search narrows it

    @gui-the-time-page-shows-a-day
    Scenario: the time page shows one day, and Today comes back to it
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I spent "1h" on task 1 on "2026-09-29"
      And I spent "2h" on task 2 on "2026-09-30"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I click "Day" on the time page
      Then the day shown is "2026-09-30"
      And the day shown shows "Ship the board"
      And the day shown does not show "Write the spec"
      When I click "‹" on the time page
      Then the day shown is "2026-09-29"
      And the day shown shows "Write the spec"
      When I click "Today" on the time page
      Then the day shown is "2026-09-30"
      When I click "Week" on the time page
      Then the week shown starts on "2026-09-28"

    @gui-a-day-of-the-week-is-shown-alone
    Scenario: a click on a day's head in a week shows that day alone
      Given today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I click the head of the day "2026-09-29"
      Then the day shown is "2026-09-29"

    @gui-the-time-page-is-searched
    Scenario: a search narrows the time page to the tasks it reaches
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I spent "1h" on task 1 on "2026-09-29"
      And I spent "2h" on task 2 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I click the tab "TIME"
      And I press "/"
      And I type "spec"
      Then the day "2026-09-29" shows "Write the spec"
      And the day "2026-09-29" does not show "Ship the board"
      And the day "2026-09-29" shows a total of "1h"

  Rule: Tasks are exported from the window, and imported into it, as a file

    @gui-the-tasks-are-exported
    Scenario: ⌘E writes the tasks to a file chosen in the system's dialog
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+e"
      And I save it in the dialog as a file of my own
      Then that file holds the task "Write the spec"
      And the window tells "the tasks were written to"

    @gui-tasks-are-imported
    Scenario: ⌘O adds the tasks of a file chosen in the system's dialog
      Given I added a task titled "Write the spec"
      And a file of my own exported from a store with a task titled "Ship it"
      When I open the window
      And I press "cmd+o"
      And I pick that file in the dialog
      Then the window is on the board
      And "BACKLOG" carries "Ship it"
      And "BACKLOG" carries "Write the spec"
      And the window tells "1 task(s) and 0 objective(s) were added"

    @gui-a-dialog-closed-changes-nothing
    Scenario: closing the dialog without choosing a file changes nothing
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+o"
      And I close the dialog
      Then the window is on the board
      And "BACKLOG" carries "Write the spec"
      And there are 0 backups

    @gui-carrying-the-store-is-offered
    Scenario: the bar offers to export the tasks and to import some
      When I open the window
      Then the bar offers "Export" under "⌘E"
      And the bar offers "Import" under "⌘O"

  Rule: Every act is in the window's menu bar

    @gui-the-menu-bar-acts-from-the-board
    Scenario: a line of the menu bar does what its key does
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I pick "New" in the menu bar
      Then the window is on the form
      When I press "esc"
      And I pick "Objectives" in the menu bar
      Then the window is on the page "OBJECTIVES"

    @gui-the-menu-button-stands-in-for-a-menu-bar
    Scenario: with no menu bar of the system's, a button opens the lines that can be done now
      Given the window runs on "linux"
      When I open the window
      And I click the menu button
      Then the menu offers "New", "New backup", "Backups", "Labels", "Export", "Import", "Search", "Objectives", "Time", "Archived", "Help", "About cairn"

    @gui-the-version-is-shown-from-the-menu
    Scenario: About cairn, in the menu bar, says which version this is
      When I open the window
      And I pick "About cairn" in the menu bar
      Then the window says the version this build reports
      When I click "Close"
      Then the window is on the board

  Rule: A page turned off in the configuration is not drawn

    @gui-a-page-off-has-no-tab
    Scenario: the time page turned off has no tab, and tab walks the others
      Given the pages turned off are "time"
      When I open the window
      Then the tabs read "TASKS", "OBJECTIVES"
      When I press "tab"
      And I press "tab"
      Then the window is on the page "TASKS"

    @gui-the-menu-offers-what-can-be-done-now
    Scenario: the menu bar offers what can be done now, and what is chosen opens the rest
      Given I added a task titled "Write the spec"
      When I open the window
      Then the menu bar can do "New", "Search", "Archived", "Export", "Import", "Labels", "Objectives", "Time", "Help"
      And the menu bar cannot do "Edit", "Focus", "Archive", "Delete", "Tasks"
      When I click the card "Write the spec"
      Then the menu bar can do "Edit", "Focus", "Archive", "Delete"

    @gui-the-menu-follows-the-page-in-sight
    Scenario: on the objectives, the lines that act on a task cannot be done
      Given I added an objective titled "Ship it"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship it"
      Then the menu bar can do "New", "Edit", "Delete", "Tasks"
      And the menu bar cannot do "Focus", "Archive", "Archived", "Objectives"

    @gui-the-menu-waits-while-a-panel-is-open
    Scenario: with a form open, nothing of the menu bar can be done until it closes
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      Then the menu bar cannot do "New", "Edit", "Search", "Help", "Tasks"

    @gui-a-line-that-cannot-be-done-does-nothing
    Scenario: picking a line that cannot be done now does nothing, and the button's menu does not offer it
      Given I added a task titled "Write the spec"
      When I open the window
      And I pick "Edit" from the menu bar
      Then the window is on the board
      When I open the menu from its button
      Then the menu offers no "Edit", "Focus", "Archive", "Delete"
      And the menu also offers "New", "Search"

    @gui-a-page-off-takes-its-menus
    Scenario: a page turned off takes its lines out of the menu bar, and only those
      Given the pages turned off are "time"
      When I open the window
      Then the menu bar offers no "Time"
      And the menu bar offers "Tasks", "Objectives", "Focus", "Archive"

    @gui-the-tasks-off-take-their-menus
    Scenario: the tasks turned off take every line that acts on a task
      Given the pages turned off are "tasks"
      When I open the window
      Then the menu bar offers no "Tasks", "Focus", "Archive", "Archived", "Export", "Import"
      And the menu bar offers "Objectives", "New", "Edit", "Delete", "Search", "Labels"

    @gui-every-page-off-leaves-what-belongs-to-none
    Scenario: every page turned off leaves the menu bar as empty as the window
      Given the pages turned off are "tasks, objectives, time"
      When I open the window
      Then the menu bar offers nothing
      When I pick "Help" from the menu bar
      Then the window is on the board

    @gui-the-first-page-shown-opens
    Scenario: with the tasks turned off the window opens on the objectives
      Given the pages turned off are "tasks"
      When I open the window
      Then the window is on the page "OBJECTIVES"

    @gui-every-page-off-is-an-empty-window
    Scenario: every page turned off opens a window with nothing in it
      Given the pages turned off are "tasks, objectives, time"
      When I open the window
      Then the window draws nothing but its ground
      When I press "cmd+n"
      Then the window is on the board

  Rule: Labels are seen and managed apart from any task

    @gui-labels-are-listed
    Scenario: the labels are listed with what carries each
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I pick "Labels" in the menu bar
      Then the labels list "doc" carried by 1 task and 0 objectives
      And the labels list "sp26-20" carried by 1 task and 0 objectives

    @gui-a-label-is-deleted-everywhere
    Scenario: a label is taken off everything carrying it once that is agreed to
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 1 with the labels "doc" and "sp26-20"
      And I edited task 2 with the labels "release" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the label "sp26-20"
      And I press "d"
      Then the window asks "Delete the label sp26-20 everywhere?"
      When I press "enter"
      Then task 1 carries exactly the labels "doc"
      And task 2 carries exactly the labels "release"
      And the labels list no "sp26-20"

    @gui-a-label-is-renamed-everywhere
    Scenario: a label is renamed on everything carrying it
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the label "sp26-20"
      And I press "r"
      And I press "backspace" 7 times
      And I type "Sprint 20"
      And I press "enter"
      Then task 1 carries exactly the labels "doc, sprint-20"
      And the labels list "sprint-20" carried by 1 task and 0 objectives

    @gui-a-rename-is-given-up
    Scenario: escape gives a rename up and leaves the label as it was
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the label "sp26-20"
      And I press "r"
      And I type "x"
      And I press "esc"
      Then task 1 carries exactly the labels "doc, sp26-20"
      And the labels are listed

  Rule: A new backup is made, and a backup restored, from the window

    @gui-backing-up-from-the-window
    Scenario: ⌘B backs the store up and says where the backup is
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+b"
      Then the window tells "backup 1 is"
      And there is 1 backup

    @gui-putting-a-backup-back-from-the-window
    Scenario: a backup chosen from the list is restored once agreed to
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+b"
      And I added a task titled "Ship the board"
      And the window has looked at the store 1 time
      And I press "shift+b"
      Then the backups listed are 1
      When I press "enter"
      Then the window is asking
      When I press "enter"
      Then the window is on the board
      And "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"
      And the window tells "backup 1 was restored"
      And there are 2 backups

    @gui-the-backups-are-left-without-putting-one-back
    Scenario: esc leaves the backups without putting one back
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+b"
      And I press "shift+b"
      And I press "esc"
      Then the window is on the board
      And there is 1 backup

  Rule: A right click says what can be done to what is under it

    @gui-a-right-click-on-a-card
    Scenario: a right click on a card chooses it and offers what can be done to it
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      Then the card "Write the spec" is chosen
      And the menu offers "Edit", "New linked task", "Focus", "Move to…", "Priority…", "Labels…", "Archive", "Delete"

    @gui-a-menu-acts
    Scenario: what is clicked in the menu is done to the card
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I click "Focus" in the menu
      Then task 1 is in focus
      And the window is on the board
      When I right-click the card "Write the spec"
      Then the menu offers "Edit", "New linked task", "Unfocus", "Move to…", "Priority…", "Labels…", "Archive", "Delete"

    @gui-a-card-is-moved-from-its-menu
    Scenario: a card's menu moves it to another column, and sets its priority
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I click "Move to…" in the menu
      Then the menu offers "backlog", "current", "blocked", "watching", "paused", "done"
      When I click "done" in the menu
      Then task 1 is in the column "done"
      When I right-click the card "Write the spec"
      And I click "Priority…" in the menu
      And I click "high" in the menu
      Then task 1 carries the priority "high"

    @gui-a-cards-menu-opens-a-linked-task
    Scenario: a card's menu opens a new task linked to it
      Given I added a task titled "Ship it" with the state "done"
      When I open the window
      And I right-click the card "Ship it"
      And I click "New linked task" in the menu
      And I set the field "title" to "It does not hold"
      And I choose "Save"
      Then task 2 is linked to task 1

    @gui-o-opens-a-linked-task
    Scenario: o on a chosen card opens a new task linked to it, in todo
      Given I added a task titled "Ship it" with the state "done"
      When I open the window
      And I click the card "Ship it"
      And I press "o"
      And I set the field "title" to "It does not hold"
      And I choose "Save"
      Then task 2 is in the column "todo"
      And task 2 is linked to task 1
      And task 1 is linked to task 2

    @gui-the-form-links-a-task
    Scenario: a link typed in the window's form links the two tasks
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the window
      And I click the card "Ship it"
      And I choose "Edit"
      And I set the field "links" to "1"
      And I choose "Save"
      Then task 2 is linked to task 1

    @gui-labels-from-a-cards-menu
    Scenario: a card's menu opens the list of its labels
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I click "Labels…" in the menu
      Then the window is choosing
      And the panel is headed "labels for task 1"

    @gui-a-menu-is-walked-with-the-keys
    Scenario: the menu is walked with the arrows, enter takes a line and esc closes it
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I press "esc"
      Then the window is on the board
      When I right-click the card "Write the spec"
      And I press "enter"
      Then the window is on the form

    @gui-a-right-click-off-the-cards
    Scenario: a right click where there is no card offers what the board does
      When I open the window
      And I right-click the column "TODO"
      Then the menu offers "New", "Search", "Archived", "Help", "New backup", "Backups", "Labels", "Export", "Import"

    @gui-a-click-off-the-menu-closes-it
    Scenario: a click off the menu closes it and does nothing else
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I click the column "BACKLOG"
      Then the window is on the board

  Rule: A deadline is chosen on a calendar

    @gui-a-click-on-a-deadline-opens-the-calendar
    Scenario: a click on the deadline opens the calendar under it
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      Then the calendar is open on "2026-10-12"
      And the calendar hangs under the field "deadline"

    @gui-the-pointer-walks-the-months
    Scenario: the arrows of its heading and the wheel walk the calendar's months
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      And I click the next month on the calendar
      Then the calendar is open on "2026-11-12"
      When I click the month before on the calendar
      And I click the month before on the calendar
      Then the calendar is open on "2026-09-12"
      When I scroll down over the calendar
      Then the calendar is open on "2026-10-12"

    @gui-a-deadline-is-chosen-on-a-calendar
    Scenario: the calendar walks the days and the months, and enter takes one
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      Then the calendar is open on "2026-10-12"
      When I press "right" 3 times
      And I press "down"
      And I press "shift+right"
      Then the calendar is open on "2026-11-22"
      When I press "enter"
      Then the window is on the form
      And the field "deadline" reads "2026-11-22"

    @gui-a-calendar-opens-on-the-day-held
    Scenario: the calendar opens on the day the field holds, and t goes back to today
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-12-25"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      Then the calendar is open on "2026-12-25"
      When I press "t"
      Then the calendar is open on "2026-10-12"

    @gui-a-calendar-takes-the-deadline-away
    Scenario: backspace on the calendar leaves the task with no deadline
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-12-25"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      And I press "backspace"
      Then the window is on the form
      And the field "deadline" reads ""

    @gui-a-calendar-offer-is-clicked-where-it-is-drawn
    Scenario: what the calendar offers is drawn once, and answers a click
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-12-25"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      And I click "No day"
      Then the window is on the form
      And the field "deadline" reads ""

    @gui-a-calendar-left-changes-nothing
    Scenario: esc leaves the calendar and the deadline as it was
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-12-25"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      And I press "right"
      And I press "esc"
      Then the window is on the form
      And the field "deadline" reads "2026-12-25"

    @gui-a-day-is-clicked-on-the-calendar
    Scenario: a day clicked on the calendar is taken
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      And I click the day "2026-10-30" on the calendar
      Then the field "deadline" reads "2026-10-30"

    @gui-an-objective-deadline-is-chosen-too
    Scenario: an objective's deadline opens the same calendar
      Given I added an objective titled "Ship the board"
      And today is "2026-10-12"
      When I open the window
      And I press "tab"
      And I click the objective "Ship the board"
      And I choose "Edit"
      And I put the keyboard on "deadline"
      And I press "enter"
      And I press "enter"
      Then the field "deadline" reads "2026-10-12"

  Rule: A form saves itself

    @gui-a-form-saves-itself
    Scenario: what was typed is saved once the typing pauses, and the form stays open
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " now"
      And the window has looked at the store 2 times
      Then task 1 is called "Write the spec now"
      And the window is on the form

    @gui-a-form-that-saved-itself-shows-it
    Scenario: the trace beside a form that saved itself says what it saved
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " now"
      And the window has looked at the store 2 times
      Then the trace begins with the last thing done to task 1

    @gui-typing-holds-the-save-off
    Scenario: nothing is saved while the typing goes on
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " no"
      And the window has looked at the store 1 time
      And I type "w"
      And the window has looked at the store 1 time
      Then task 1 is called "Write the spec"

    @gui-a-saved-form-closes-without-a-question
    Scenario: a form that saved itself closes on esc without asking
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " now"
      And the window has looked at the store 2 times
      And I press "esc"
      Then the window is on the board
      And task 1 is called "Write the spec now"

    @gui-a-new-task-is-made-when-it-is-saved
    Scenario: a task being added is made when it is saved, and not before
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      Then there are no tasks
      And the window is on the form
      When I choose "Save"
      Then task 1 is called "Plan the week"
      And there is 1 task

    @gui-a-new-task-has-an-empty-trace
    Scenario: the trace of a task being added stays empty while it is typed
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      Then the form writes the trace, empty

    @gui-a-new-task-shows-no-other-tasks-trace
    Scenario: a task added after another was opened shows an empty trace, not the other's
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I double-click the card "Write the spec"
      And I press "esc"
      And I press "cmd+n"
      Then the form writes the trace, empty

    @gui-cancelling-a-new-task-makes-nothing
    Scenario: cancelling a task being added makes nothing, asks nothing, and leaves nothing to pick up
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      And I choose "Cancel"
      Then the window is on the board
      And there are no tasks
      When the window is closed and opened again
      And I press "cmd+n"
      Then the window is on the form
      And the field "title" reads ""

    @gui-escape-on-a-new-task-makes-nothing
    Scenario: escape on a task being added is cancelling it, without a question
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And I press "esc"
      Then the window is on the board
      And there are no tasks
      When I press "cmd+n"
      Then the window is on the form

    @gui-a-new-task-is-kept-as-a-draft
    Scenario: a task being added and left unsaved is kept as a draft, and the next one offers to pick it up
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      And the window is closed and opened again
      Then there are no tasks
      When I press "cmd+n"
      Then the window asks to pick up the task left unsaved
      When I choose "Pick it up"
      Then the window is on the form
      And the field "title" reads "Plan the week"
      When I choose "Save"
      Then task 1 is called "Plan the week"

    @gui-a-draft-can-be-left
    Scenario: a draft not wanted is left, and a new task starts from nothing
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      And the window is closed and opened again
      And I press "cmd+n"
      And I choose "Start afresh"
      Then the window is on the form
      And the field "title" reads ""
      When I press "esc"
      And the window is closed and opened again
      And I press "cmd+n"
      Then the window is on the form

    @gui-a-saved-task-leaves-no-draft
    Scenario: a task saved leaves no draft behind it
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And the window has looked at the store 3 times
      And I choose "Save"
      And the window is closed and opened again
      And I press "cmd+n"
      Then the window is on the form
      And the field "title" reads ""

    @gui-a-task-is-added-to-the-column-clicked
    Scenario: a task added after a click on a column goes in that column
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I click the card "Write the spec"
      And I click the column "CURRENT"
      And I click "New"
      And I type "Plan the week"
      And I choose "Save"
      Then "CURRENT" carries "Plan the week"

    @gui-a-double-click-on-a-column-adds-there
    Scenario: a double click where a column holds no card adds a task to it
      When I open the window
      And I double-click the column "TODO"
      Then the window is on the form
      When I type "Plan the week"
      And I choose "Save"
      Then "TODO" carries "Plan the week"

    @gui-a-click-between-two-cards-chooses-neither
    Scenario: a click in the space between two cards chooses neither of them
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship the board" with the state "todo"
      When I open the window
      And I click between the cards "Write the spec" and "Ship the board"
      Then no card is chosen

    @gui-a-double-click-between-two-cards-adds-there
    Scenario: a double click between two cards adds a task between them
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship the board" with the state "todo"
      When I open the window
      And I double-click between the cards "Write the spec" and "Ship the board"
      Then the window is on the form
      When I type "Plan the week"
      And I choose "Save"
      Then "TODO" reads "Write the spec", "Plan the week", "Ship the board"

    @gui-a-right-click-on-a-column-adds-there
    Scenario: New, from a right click on a column, adds a task to that column
      When I open the window
      And I right-click the column "TODO"
      And I click "New" in the menu
      And I type "Plan the week"
      And I choose "Save"
      Then "TODO" carries "Plan the week"

    @gui-a-saved-form-keeps-what-is-typed
    Scenario: saving on its own keeps the space typed at the end of a title
      Given I added a task titled "Fix the"
      When I open the window
      And I click the card "Fix the"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " "
      And the window has looked at the store 2 times
      And I type "bug"
      And the window has looked at the store 2 times
      Then task 1 is called "Fix the bug"

    @gui-a-refused-save-is-not-tried-again
    Scenario: what the store refuses is said once and not tried again until it changes
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "title"
      And I type " now"
      And the store refuses what it is asked to change
      And the window has looked at the store 2 times
      Then it says what went wrong
      When the window has looked at the store 4 times
      Then it says nothing went wrong
      And the engine was asked "task edit" once

    @gui-a-task-added-once-whatever-its-steps
    Scenario: a task whose steps were refused is not made, and not tried again every beat
      Given the store refuses to add a step
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And I set the field "steps" to "Book the room"
      And I press "cmd+enter"
      And the window has looked at the store 6 times
      Then there are no tasks
      And the engine was asked "task new" once

    @gui-a-new-task-with-no-title-is-not-made
    Scenario: a form adding a task with no title yet makes nothing
      When I open the window
      And I press "cmd+n"
      And the window has looked at the store 3 times
      Then there are no tasks

  Rule: What the board can do, the window can do

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

    @gui-searching-raises-the-task-a-step-belongs-to
    Scenario: a word found only in a step answers the task it belongs to
      Given I added a task titled "Ship it"
      And I added the step "Write the pipeline" to task 1
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

    @gui-searching-by-label-and-by-column
    Scenario: a search reaches the labels and the column, as the engine's does
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship the board"
      And I edited task 1 with the labels "backend" and "sprint-172"
      When I open the window
      And I press "/"
      And I type "backend"
      Then "TODO" carries "Write the spec"
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

    @gui-the-help-names-the-keys-of-the-system
    Scenario: away from a Mac the help and the bar say Ctrl where a Mac says ⌘
      Given the window runs on "windows"
      When I open the window
      And I click "Help"
      Then the help says what "Ctrl+N" does
      And the help says nothing of "⌘"

    @gui-a-mac-keeps-its-own-keys
    Scenario: on a Mac the help says ⌘
      Given the window runs on "darwin"
      When I open the window
      And I click "Help"
      Then the help says what "⌘N" does

    @gui-the-help-scrolls
    Scenario: a help longer than the window scrolls, and says so
      Given a window of 1024 by 640
      When I open the window
      And I click "Help"
      Then the help shows some of its lines and not all
      And the help draws a bar to scroll it by
      When I press "down" 3 times
      Then the help starts 3 lines down
      When I press "down" 200 times
      Then the help shows its last line

    @gui-the-help-fits-what-it-says
    Scenario: a help on a tall window is as tall as what it says
      Given a window of 1600 by 2000
      When I open the window
      And I click "Help"
      Then the help shows every line
      And the help is no taller than what it says

    @gui-the-help-keeps-its-words-inside
    Scenario: no line of the help runs past its panel
      Given a window of 1024 by 640
      When I open the window
      And I click "Help"
      Then every line of the help fits inside its panel

  Rule: The window follows the store

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

  Rule: A task's trace stays inside the panel that shows it

    @gui-opening-a-task-opens-it-to-be-edited
    Scenario: opening a task opens it to be edited, its trace beside what it carries
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I double-click the card "Write the spec"
      Then the window is on the form
      And every field of the form stands left of the trace
      And the trace begins with the last thing done to task 1

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

    @gui-the-trace-has-its-own-place
    Scenario: the trace is beside what the task carries, not under it
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the trace is beside the fields
      And the trace stays inside the panel

    @gui-a-new-task-shows-its-trace-from-the-start
    Scenario: a task being added shows its trace from the start, empty, so nothing appears beside the form once it saves itself
      When I open the window
      And I press "cmd+n"
      Then the form writes the trace, empty
      And the trace is beside the fields

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

    @gui-the-trace-reads-newest-first
    Scenario: the trace reads from the last thing done, with no blank line between two acts
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the trace begins with the last thing done to task 1
      And no line of the trace is empty

    @gui-the-trace-scrolls
    Scenario: a trace longer than its box scrolls, and stops at both ends
      Given I added a task titled "Write the spec"
      And it was worked on 20 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      Then the trace has more to show than it has room for
      And the trace is at its beginning
      When I scroll down over the trace
      Then the trace has moved
      And the trace stays inside the panel
      When I scroll up over the trace
      Then the trace is at its beginning
      When I scroll down over the trace 99 times
      Then the trace is at its end
      And the trace stays inside the panel

    @gui-the-bar-beside-the-trace-is-dragged
    Scenario: the bar beside the trace is taken hold of and moved
      Given I added a task titled "Write the spec"
      And it was worked on 20 times
      When I open the window
      And I click the card "Write the spec"
      And I press "enter"
      And I drag the trace's bar to the bottom
      Then the trace is at its end
      And the trace stays inside the panel

  Rule: What the arrows reach, they bring into sight

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

  Rule: The pages are walked round with tab

    @gui-tab-walks-the-pages
    Scenario: tab moves to the next page and shift+tab to the one before
      When I open the window
      Then the tab is "TASKS"
      When I press "tab"
      Then the tab is "OBJECTIVES"
      When I press "shift+tab"
      Then the tab is "TASKS"
      When I press "shift+tab"
      Then the tab is "TIME"

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

    @gui-the-time-tab-counts-before-it-is-opened
    Scenario: the time tab says how many tasks were worked on this week before it is opened
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the board"
      And I spent "1h" on each of tasks 1 to 2 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      Then the tab "TIME" says it holds 2

  Rule: An objective opens on the work that serves it

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
      And what serves it is drawn one line to a tile
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

    @gui-an-objective-card-says-the-time-spent
    Scenario: an objective's card says the time spent on the tasks serving it
      Given I added an objective titled "Ship the desktop board"
      And I added a task titled "Write the spec" serving objective 1
      And I added a task titled "Ship it" serving objective 1
      And I spent "1h" on task 1 on "2026-09-29"
      And I spent "30m" on task 2 on "2026-09-30"
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the card for objective 1 says "1h30m"

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
    Scenario: a task is focused and unfocused from the window
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "f"
      Then the engine was asked "task focus" for task 1
      When I press "f"
      Then the engine was asked "task unfocus" for task 1

    @gui-a-list-names-the-task-it-is-for
    Scenario: the list a field opens says which task it is for
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      Then the window is choosing
      And the panel is headed "objectives for task 1"
      And the list offers "Ship it"

    @gui-a-list-is-corrected-like-any-other-field
    Scenario: what was typed into a list can be taken back out
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
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
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I press "enter"
      And I press "cmd+enter"
      Then the window is on the form
      And the bar offers "Save" under "⌘⏎"
      And the bar offers "Choose" under "⏎"
      When I press "cmd+enter"
      Then the window is on the board
      And the engine was asked "task edit" for task 1

    @gui-a-space-in-a-list-is-a-space
    Scenario: a space typed into a list is a space, because a title has them
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I type "ship it"
      Then the list offers "Ship it"

    @gui-a-double-click-is-done-with-the-list
    Scenario: a double-click takes a row and is done, which is what a hand expects
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I double-click the list row "Ship it"
      Then the window is on the form
      And the field "objectives" names "Ship it"

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
    Scenario: a list of more than fit stays inside the panel and scrolls
      Given I added a task titled "Write the spec"
      And I added 40 objectives
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
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
      And I added 40 objectives
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I press "down" 30 times
      Then the row the keyboard is on is in sight

    @gui-clicking-a-field-that-a-list-answers-opens-it
    Scenario: clicking the field opens its list, which is what a pointer is for
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      Then the window is choosing
      And the panel is headed "objectives for task 1"

    @gui-a-list-answers-the-mouse
    Scenario: a row is chosen by clicking it, the way everything else here is
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I click the list row "Ship it"
      Then the window is choosing
      When I press "enter"
      And I press "cmd+enter"
      Then the window is on the form
      And the field "objectives" names "Ship it"

    @gui-a-click-in-a-list-leaves-the-form-under-it-alone
    Scenario: a click inside the list does not reach the form behind it
      Given I added an objective titled "Ship it"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I click the line that narrows the list
      Then the window is choosing
      And the form is still on the field "objectives"

    @gui-the-list-narrows-as-it-is-typed-into
    Scenario: typing in the list keeps only what matches
      Given I added an objective titled "Ship it"
      And I added an objective titled "Rewrite the tests"
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "objectives"
      And I type "tests"
      Then the list offers "Rewrite the tests"
      And the list does not offer "Ship it"

    @gui-a-card-says-it-carries-more-than-it-shows
    Scenario: a card says a description exists and how far its steps have got
      Given I added a task titled "Write the spec" with the description "Two columns"
      And I added a task titled "Ship it"
      And I added the step "Write the tests" to task 2, done
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

    @gui-a-lit-offer-holds-its-words-in-its-middle
    Scenario: the offer under the pointer is lit around its words, which sit in its middle
      When I open the window
      And I rest the pointer on the offer "Search"
      Then the offer "Search" is lit around its words, in their middle

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

  Rule: A task opens on its steps, ticked, renamed, added and taken away where they are

    @gui-the-arrows-walk-the-steps
    Scenario: the arrows walk the steps, and space ticks the one they are on
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1
      And I added the step "Get the nod" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "steps"
      Then the step the keyboard is on is "Draft the columns"
      When I press "down"
      Then the step the keyboard is on is "Get the nod"
      When I press "space"
      Then the step "Get the nod" is ticked
      And task 1 carries the step "Get the nod", done

    @gui-a-task-with-steps-chooses-from-a-list
    Scenario: a task with steps opens a list from its form, and the form stays drawn under it
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the field "column"
      Then the window is choosing
      And the list reads "backlog", "todo", "current", "blocked", "watching", "paused", "done"

    @gui-a-step-is-renamed-in-its-tile
    Scenario: enter renames the step the keyboard is on, in its tile
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "steps"
      And I press "enter"
      And I type " the columns"
      And I press "enter"
      And I choose "Save"
      Then the steps of task 1 read "Draft the columns"

    @gui-steps-are-added-one-after-another
    Scenario: the last line adds a step, and the next is waiting behind it
      Given I added a task titled "Ship it"
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "steps"
      And I press "enter"
      And I type "Draft the columns"
      And I press "enter"
      And I type "Get the nod"
      And I press "enter"
      And I choose "Save"
      Then the steps of task 1 read "Draft the columns", "Get the nod"

    @gui-a-step-is-deleted-once-agreed
    Scenario: a delete key on the step clicked asks, and agreed takes it away at once, the form unsaved
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I press "backspace"
      Then the window asks "Delete the step “Draft”?"
      When I press "enter"
      Then the steps of task 1 read "Keep"
      And the form's steps read "Keep"
      And the form is open on task 1
      And the last act on task 1 changed its steps from "[ ] Draft; [ ] Keep" to "[ ] Keep"

    @gui-the-form-stays-behind-the-question-that-deletes-a-step
    Scenario: the form is still drawn while the window asks to delete a step
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I press "backspace"
      Then the window asks "Delete the step “Draft”?"
      And the form is drawn behind the question

    @gui-a-question-stands-out-from-what-is-behind-it
    Scenario: a question is outlined in the colour of what is chosen, over a window dimmed behind it
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I press "backspace"
      Then the question is outlined in the colour of what is chosen
      And what is behind the question is dimmed

    @gui-a-step-is-kept-when-its-delete-is-cancelled
    Scenario: the forward delete asks too, and cancelled leaves the step where it was
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Keep"
      And I press "delete"
      Then the window asks "Delete the step “Keep”?"
      When I press "esc"
      Then the steps of task 1 read "Draft", "Keep"
      And the form's steps read "Draft", "Keep"
      And the form is open on task 1

    @gui-steps-taken-together-are-deleted-at-once
    Scenario: steps ⌘-clicked together are deleted in one go, once agreed
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I ⌘-click the step "Review"
      And I press "backspace"
      Then the window asks "Delete 2 steps?"
      When I press "enter"
      Then the steps of task 1 read "Keep"
      And the form's steps read "Keep"
      And the last act on task 1 changed its steps from "[ ] Draft; [ ] Keep; [ ] Review" to "[ ] Keep"

    @gui-steps-stretched-to-are-taken-together
    Scenario: a ⇧-click takes every step from the one chosen to the one clicked
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      And I added the step "Ship" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Keep"
      And I ⇧-click the step "Ship"
      And I press "delete"
      Then the window asks "Delete 3 steps?"
      When I press "esc"
      Then the steps of task 1 read "Draft", "Keep", "Review", "Ship"
      When I press "backspace"
      And I press "enter"
      Then the steps of task 1 read "Draft"

    @gui-a-step-taken-is-let-go-alone
    Scenario: a ⌘-click on a step taken lets that one go, and keeps the others
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I ⌘-click the step "Keep"
      And I ⌘-click the step "Review"
      And I ⌘-click the step "Keep"
      And I press "backspace"
      Then the window asks "Delete 2 steps?"
      When I press "enter"
      Then the steps of task 1 read "Keep"

    @gui-the-last-step-taken-can-be-let-go
    Scenario: a ⌘-click lets the last step taken go, and a delete key then deletes nothing
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I ⌘-click the step "Draft"
      And I press "backspace"
      Then the form is open on task 1
      And the steps of task 1 read "Draft"

    @gui-every-step-is-taken-at-once
    Scenario: ⌘A takes every step of the list
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Keep"
      And I press "cmd+a"
      And I press "backspace"
      Then the window asks "Delete 3 steps?"

    @gui-steps-taken-outlast-a-save
    Scenario: the steps taken are still taken once the form has saved itself
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "title"
      And I type " now"
      And I click the step "Keep"
      And I press "cmd+a"
      And the window has looked at the store 2 times
      Then task 1 is called "Ship it now"
      When I press "backspace"
      Then the window asks "Delete 3 steps?"

    @gui-steps-taken-together-move-together
    Scenario: steps taken together are dragged together, in their order
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      And I added the step "Ship" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I ⌘-click the step "Keep"
      And I drag the step "Keep" onto the step "Ship"
      Then the form's steps read "Review", "Ship", "Draft", "Keep"
      When I choose "Save"
      Then the steps of task 1 read "Review", "Ship", "Draft", "Keep"

    @gui-a-step-deleted-leaves-the-rest-of-the-form-as-it-was
    Scenario: what was typed in the form and not saved is still there once a step is deleted
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "title"
      And I type " now"
      And I click the step "Keep"
      And I press "backspace"
      And I press "enter"
      Then the field "title" reads "Ship it now"
      And the steps of task 1 read "Draft"
      And task 1 is titled "Ship it"

    @gui-a-step-in-edit-loses-a-letter-not-itself
    Scenario: a double click puts a step in edit, where backspace takes a letter and asks nothing
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I double-click the step "Draft"
      And I press "backspace"
      Then the form is open on task 1
      When I press "enter"
      And I choose "Save"
      Then the steps of task 1 read "Draf"

    @gui-a-step-never-saved-is-taken-out-of-the-list
    Scenario: a step of a task not yet added is taken out of its list, once agreed
      When I open the window
      And I choose "New"
      And I put the keyboard on "steps"
      And I press "enter"
      And I type "Draft"
      And I press "enter"
      And I type "Keep"
      And I press "enter"
      And I click the step "Draft"
      And I press "backspace"
      Then the window asks "Delete the step “Draft”?"
      When I press "enter"
      Then the form's steps read "Keep"

    @gui-the-wheel-scrolls-the-steps
    Scenario: the wheel scrolls the steps, wherever the keyboard is
      Given a window of 1024 by 640
      And I added a task titled "Ship it"
      And I added 12 steps to task 1
      When I open the window
      And I double-click the card "Ship it"
      Then the steps are at their beginning
      When I scroll down over the steps
      Then the steps have moved
      When I scroll down over the steps 99 times
      Then the last step is in sight

    @gui-a-step-in-hand-shows-where-it-lands
    Scenario: a step in hand follows the pointer, and says where it would land
      Given I added a task titled "Ship it"
      And I added the step "First" to task 1
      And I added the step "Second" to task 1
      And I added the step "Third" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I take hold of the step "First" and hold it over the step "Third"
      Then a step is in hand, under the pointer
      And it would land below "Third"

    @gui-a-step-is-dragged-to-its-place
    Scenario: a step dragged onto another takes its place
      Given I added a task titled "Ship it"
      And I added the step "First" to task 1
      And I added the step "Second" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I drag the step "Second" onto the step "First"
      Then the steps are drawn as a list to tick, "Second" first
      When I choose "Save"
      And I double-click the card "Ship it"
      Then the steps are drawn as a list to tick, "Second" first

    @gui-steps-done-and-left-in-turn-are-saved
    Scenario: steps done and left in turn are saved, each kept among its own
      Given I added a task titled "Ship it"
      And I added the step "One" to task 1, done
      And I added the step "Two" to task 1, done
      And I added the step "Three" to task 1
      And I added the step "Four" to task 1, done
      And I added the step "Five" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "steps"
      And I press "down"
      And I press "space"
      And I choose "Save"
      Then the window is on the board
      And task 1 carries the step "Five", done

    @gui-a-tick-is-kept-whatever-becomes-of-the-form
    Scenario: a step ticked is done at once, and nothing is left to save for it
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "steps"
      And I press "space"
      Then task 1 carries the step "Draft", done
      When I press "esc"
      Then the window is on the board
      And task 1 carries the step "Draft", done

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

    @gui-a-run-of-the-trace-spans-its-lines
    Scenario: a run dragged across the trace takes in every line it passes over
      Given I added a task titled "Write the spec"
      When I open the window
      And I double-click the card "Write the spec"
      And I drag across the trace from the start of its first line to the end of its second
      And I press "cmd+c"
      Then the clipboard reads the first 2 lines of the trace

    @gui-the-steps-left-come-first
    Scenario: what is left comes first
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1, done
      And I added the step "Get the client's nod" to task 1
      When I open the window
      And I double-click the card "Ship it"
      Then the steps are drawn as a list to tick, "Get the client's nod" first

  Rule: The objectives a task serves are picked from a list, not typed

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

    @gui-the-labels-field-opens-a-list
    Scenario: the labels field opens a list of every label, those the task carries ticked
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 1 with the labels "backend" and "sp26-20"
      And I edited task 2 with the labels "release" and "sp26-20"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      Then the window is choosing
      And the panel is headed "labels for task 1"
      And the list reads "backend", "release", "sp26-20"
      And the label "backend" is ticked
      And the label "release" is not ticked

    @gui-labels-are-created-several-at-a-time
    Scenario: a label nobody carries is created under the name it will have, and the list stays open for the next
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the labels "backend" and "release"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      And I type "Sp26 21"
      Then the list offers to create "sp26-21"
      When I press "enter"
      And I type "front"
      And I press "enter"
      And I press "cmd+enter"
      Then the field "labels" reads "front sp26-21"
      When I choose "Save"
      Then task 1 carries the labels "front, sp26-21"

    @gui-a-label-spelt-otherwise-is-the-one-there-is
    Scenario: a label typed another way is the label there is, and nothing is created
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 2 with the labels "backend" and "sp26-20"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      And I type "SP26_20"
      Then the list reads "sp26-20"
      And the list offers to create nothing
      When I press "enter"
      And I press "cmd+enter"
      And I choose "Save"
      Then task 1 carries the labels "sp26-20"

    @gui-a-label-the-rule-refuses-is-not-created
    Scenario: a label no label may be is said why, and not created
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      And I type "ado:77594"
      Then the list says a label has no family
      And the list offers to create nothing
      When I press "enter"
      And I press "cmd+enter"
      Then the field "labels" reads ""

    @gui-a-label-taken-off-is-offered-no-more
    Scenario: a label taken off its last task is offered no more
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 1 with the labels "backend" and "sp26-20"
      And I edited task 2 with the labels "release" and "sp26-20"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      And I press "enter"
      And I press "cmd+enter"
      And I choose "Save"
      Then task 1 carries the labels "sp26-20"
      And I click the card "Write the spec"
      And I choose "Edit"
      And I put the keyboard on "labels"
      And I press "enter"
      Then the list reads "release", "sp26-20"

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
      Then the panel is headed "Archive every task in DONE?"
      When I choose "Agree"
      Then "DONE" does not carry "Write the spec"
      And "DONE" does not carry "Ship it"

  Rule: What is on screen says what it is

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
    Scenario: the form of a task with steps fits above its offers, even in the smallest window
      Given a window of 1024 by 640
      And I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1
      When I open the window
      And I click the card "Ship it"
      And I choose "Edit"
      Then no field of the form reaches what it offers
      And no two fields of the form overlap
      And the title runs the width of the form, above everything else
      And the description runs the width of the form, under the title
      And the form groups its fields under "STATE", "TIME", "BLOCKED", "LINKS"

    @gui-a-new-task-goes-in-with-its-steps
    Scenario: a new task is added with the steps already written down for it
      When I open the window
      And I choose "New"
      And I set the field "title" to "Ship it"
      And I set the field "steps" to "Draft the columns"
      And I choose "Save"
      Then the steps of task 1 read "Draft the columns"

    @gui-the-steps-read-as-a-list-to-tick
    Scenario: the steps of a task being edited read as a list to tick, written as lines underneath
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1
      When I open the window
      And I click the card "Ship it"
      And I choose "Edit"
      Then the steps are drawn as a list to tick, "Draft the columns" among them
      When I click the field "steps"
      Then the field "steps" reads "#1 [ ] Draft the columns"

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
      And I added the step "Draft the columns" to task 1
      And I added a task titled "Other"
      When I open the window
      And I click the card "Ship it"
      And I ⌘-click the card "Other"
      And I press "d"
      Then the question says of task 1 "its 1 step goes with it"
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

  Rule: An open task shows what is done to it elsewhere

    @gui-an-open-task-shows-a-step-added-elsewhere
    Scenario: a step added elsewhere appears on the task open in the window, and in its trace
      Given I added a task titled "Ship it"
      When I open the window
      And I double-click the card "Ship it"
      And somebody else adds the step "Draft" to task 1
      And the window looks again
      Then the form's steps read "Draft"
      And the trace begins with the last thing done to task 1

    @gui-an-open-task-keeps-what-was-typed-when-it-changes-elsewhere
    Scenario: what was typed stays, and what was not takes what was done elsewhere
      Given I added a task titled "Ship it"
      When I open the window
      And I double-click the card "Ship it"
      And I put the keyboard on "title"
      And I type " now"
      And somebody else renames task 1 "Ship it today"
      And somebody else adds the step "Draft" to task 1
      And the window looks again
      Then the field "title" reads "Ship it now"
      And the form's steps read "Draft"
