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

  Rule: Cards are the same height, whatever they carry

    @gui-cards-share-one-height
    Scenario: a card with a great deal on it is as tall as an empty one
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it" with the description "Two columns, the client says"
      When I open the window
      Then every card in "BACKLOG" is the same height

  Rule: Each column and the band stand on a ground of their own

    @gui-a-columns-head-names-and-counts-it
    Scenario: each column's head names it at its left, in words, and counts its cards at its right
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      Then the column heads read "Backlog 0", "Todo 1", "Current 0", "Done 0"
      And every column's name stands at the left of its head
      And every column's name stands clear of the edge above it

  Rule: A column, or a lane, is widened by dragging the gutter beside it

    @gui-a-gutter-shows-it-can-be-dragged
    Scenario: the pointer over a column's gutter says the column can be widened
      When I open the window
      Then the pointer over the gutter after "todo" is the one for widening sideways

    @gui-a-gutter-dragged-picks-no-card
    Scenario: dragging a gutter takes hold of no card
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I drag the gutter between "todo" and "current" 60 to the right
      Then "todo" carries "Write the spec"
      And no card is being dragged

  Rule: The band is made taller or shorter by dragging the edge above it

    @gui-the-band-grows-when-its-edge-is-dragged
    Scenario: the edge between the columns and the band, dragged up, gives the band what it takes from the columns
      When I open the window
      And I drag the edge above the band 100 up
      Then the band is 100 taller than it opened
      And "todo" is 100 shorter than it opened
      And "backlog" is as tall as it opened

  Rule: A size given is taken back by a double-click or by the menu

    @gui-a-double-click-on-a-gutter-evens-the-columns
    Scenario: a double-click on a gutter between columns makes every column as wide as the others
      When I open the window
      And I drag the gutter between "todo" and "current" 60 to the right
      And I double-click the gutter between "todo" and "current"
      Then every column is as wide as the others

    @gui-a-double-click-on-a-lane-gutter-evens-the-lanes
    Scenario: a double-click on a gutter between lanes makes every lane as wide as the others
      When I open the window
      And I drag the gutter between the "blocked" and "watching" lanes 40 to the right
      And I drag the gutter between "todo" and "current" 60 to the right
      And I double-click the gutter between the "blocked" and "watching" lanes
      Then every lane is as wide as the others
      And "todo" is 60 wider than it opened

    @gui-a-double-click-on-the-band-edge-gives-it-back-its-height
    Scenario: a double-click on the edge above the band gives the band the height it opened with
      When I open the window
      And I drag the edge above the band 100 up
      And I double-click the edge above the band
      Then the band is as tall as it opened

    @gui-reset-sizes-takes-every-size-back
    Scenario: Reset sizes, in the command bar, takes back every width and the band's height
      When I open the window
      And I drag the gutter between "todo" and "current" 60 to the right
      And I drag the gutter between the "blocked" and "watching" lanes 40 to the right
      And I drag the edge above the band 100 up
      And I press "cmd+k"
      And I type "reset sizes"
      And I press "enter"
      Then every column is as wide as the others
      And every lane is as wide as the others
      And the band is as tall as it opened

  Rule: The sizes are kept from one window to the next only when the configuration says so

    @gui-the-sizes-are-kept-when-the-configuration-says-so
    Scenario: with keep_sizes on, a window opened again has the sizes the last one was left with
      Given the configuration keeps the window's sizes
      When I open the window
      And I drag the gutter between "todo" and "current" 60 to the right
      And I drag the edge above the band 100 up
      And the window is closed and opened again
      Then "todo" is 60 wider than it opened
      And the band is 100 taller than it opened

  Rule: The board is filtered by column, priority, label, objective and due day, and every filter is cleared at once

    Background:
      Given I added a task titled "Write the spec" in "todo" at the priority "high" carrying the label "backend"
      And I added a task titled "Review the plan" in "todo" at the priority "normal" carrying the label "backend"
      And I added a task titled "Ship it" in "current" at the priority "high" carrying the label "sprint"

    @gui-every-filter-is-cleared-at-once
    Scenario: one act takes every filter away
      When I open the window
      And I set the filter "priority" to "high"
      And I set the filter "label" to "backend"
      And I clear every filter
      Then "todo" carries "Review the plan"
      And "current" carries "Ship it"
      And the board says it is filtered by ""

    @gui-a-filters-list-is-left-untouched
    Scenario: left closes a filter's column and leaves the filter as it was
      When I open the window
      And I press "shift+f"
      And I click the filter "priority"
      And I press "down"
      And I press "left"
      Then the board says it is filtered by ""
      And the window shows the filters
      And the menu of filters shows 1 column

    @gui-esc-closes-the-whole-menu
    Scenario: space takes a value and stays, esc closes the menu from any column, what was taken kept
      When I open the window
      And I press "shift+f"
      And I press "p"
      And I press "down" 2 times
      And I press "space"
      Then the board says it is filtered by "priority high"
      And the filter "priority" lists its values
      When I press "esc"
      Then the window is on the board
      And the board says it is filtered by "priority high"

    @gui-the-filter-columns-slide-in-a-narrow-window
    Scenario: in a narrow window the oldest column slides behind a stub, which a click brings back
      Given a window of 720 by 560
      And I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      When I open the window
      And I press "shift+f"
      And I press "o"
      And I press "right"
      Then the menu of filters shows 2 columns and a stub
      When I click the stub of the menu of filters
      Then the menu of filters shows 2 columns

    @gui-a-filter-is-taken-off-by-its-key
    Scenario: backspace on a line of the menu of filters takes that filter off, and that one alone
      When I open the window
      And I set the filter "priority" to "high"
      And I set the filter "label" to "backend"
      And I press "up"
      And I press "backspace"
      Then the board says it is filtered by "label backend"
      And the window shows the filters

    @gui-backspace-takes-the-last-pill-off-the-board
    Scenario: backspace on the board takes off the last filter in force, then the search, a pill at a time
      When I open the window
      And I set the filter "priority" to "high"
      And I set the filter "label" to "backend"
      And I press "esc"
      And I press "/"
      And I type "spec"
      And I press "enter"
      And I press "backspace"
      Then the board says it is filtered by "priority high"
      When I press "backspace"
      And I press "backspace"
      Then the board says it is filtered by ""
      And no search is kept

    @gui-a-value-chosen-again-takes-its-filter-off
    Scenario: the value a filter is on, chosen again, takes the filter off
      When I open the window
      And I set the filter "priority" to "high"
      And I press "esc"
      And I click the pill "priority: high"
      And I press "enter"
      Then the board says it is filtered by ""

    @gui-a-kept-search-stands-first-among-the-pills
    Scenario: a search kept stands first among the filters' pills, and Clear takes both away
      When I open the window
      And I set the filter "priority" to "high"
      And I press "esc"
      And I press "/"
      And I type "spec"
      And I press "enter"
      Then the pills after the tabs read "/ spec, priority: high, Clear, + Filter"
      When I click the pill "Clear"
      Then the board says it is filtered by ""
      And no search is kept

    @gui-the-filters-narrow-the-objectives
    Scenario: the filters narrow the objectives by what an objective is, and a filter meaning nothing to one is drawn faint
      Given I added an objective titled "Ship cairn"
      And I added an objective titled "Plan the trip"
      When I open the window
      And I set the filter "objective" to "2"
      And I set the filter "column" to "todo"
      And I press "esc"
      And I click the tab "OBJECTIVES"
      Then the objectives shown read "Plan the trip"
      And the pill "column: todo" is drawn faint
      And the pill "objective: #2 Plan the trip" is not drawn faint

    @gui-the-filters-narrow-the-time
    Scenario: the filters narrow the time page to the tasks they keep, its totals with them
      Given I spent "1h" on task 1 on "2026-09-29"
      And I spent "2h" on task 2 on "2026-09-29"
      And today is "2026-09-30"
      When I open the window
      And I set the filter "priority" to "high"
      And I press "esc"
      And I click the tab "TIME"
      Then the day "2026-09-29" shows "Write the spec"
      And the day "2026-09-29" does not show "Review the plan"
      And the day "2026-09-29" shows a total of "1h"

  Rule: A window too small says what it needs

    @gui-too-small-says-so
    Scenario: a window under the smallest size says what it needs
      Given a window of 320 by 240
      When I open the window
      Then it says the window is too small
      And it says what size it needs

  Rule: A click outside a panel closes it, where closing loses nothing

    @gui-clicking-off-a-new-task-saves-nothing
    Scenario: clicking off a task being added asks first, and what is thrown away is not made and is kept as a draft
      When I open the window
      And I press "cmd+n"
      And I type "Plan the week"
      And I click outside the panel
      Then the panel is headed "Throw away what you typed?"
      When I choose "Throw away"
      Then the window is on the board
      And there are no tasks
      When I press "cmd+n"
      Then the window asks to pick up the task left unsaved

  Rule: The mouse chooses, opens and scrolls

    @gui-the-wheel-scrolls-one-lane
    Scenario: the wheel scrolls the lane of the band under the pointer and no other
      Given I added 6 tasks to "watching"
      And I added 6 tasks to "blocked"
      When I open the window
      And I scroll down over the lane "watching"
      Then the lane "watching" is scrolled
      And the lane "blocked" is not scrolled

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

  Rule: Dragging a card moves the work

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

    @gui-a-drop-in-done-asks-about-open-steps
    Scenario: a card dropped in DONE with steps still open asks to tick them too
      Given I added a task titled "Ship it" with the state "todo"
      And I added the step "Write the spec" to task 1
      When I open the window
      And I drag the card "Ship it" onto "DONE"
      Then the window asks to put task 1 in done, ticking its steps
      When I click "Finish"
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

  Rule: A card says when it is wanted and what it has cost

    @gui-a-card-says-when-it-is-due
    Scenario: a day close enough to matter is written on the card
      Given I added a task titled "Write the spec"
      And I edited task 1 with the deadline "2026-10-01"
      When I open the window
      Then the card "Write the spec" says "2d"

  Rule: A form is the whole of a task

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

  Rule: A form is moved by dragging it where it holds no text

    @gui-the-form-moves-when-dragged-by-its-ground
    Scenario: the form moves with the pointer dragged from a place that holds no text
      When I open the window
      And I press "cmd+n"
      And I drag the form from a place that holds no text 80 to the right and 40 down
      Then the form stands 80 to the right and 40 lower than it opened

    @gui-the-form-stays-when-its-text-is-dragged
    Scenario: dragging from one of the form's texts leaves the form where it opened
      When I open the window
      And I press "cmd+n"
      And I drag the form from its heading 80 to the right and 40 down
      Then the form stands where it opened

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

  Rule: The board arrives by the intro its settings name, and none when the system asks for less motion

    @gui-the-board-arrives-by-the-intro-named
    Scenario: the window opens on its board arriving by the intro its settings name, over a breath later
      Given the workspace "home", the default
      And the workspace "home" holds a task titled "Write the spec"
      And the settings of the workspace "home" are written by hand:
        """
        [window]
        intro = "deck"
        """
      When I open the window on the workspace "home"
      And the window draws its first frame
      Then the board is arriving by "deck"
      When the window draws a frame 2 seconds later
      Then no cairn is being built

  Rule: The window carries cairn's logo

    @gui-x11-is-handed-the-logo
    Scenario: on X11 the window hands its logo over itself, at the sizes a bar asks for
      When I open the window
      Then the icon it hands X11 holds cairn's logo at 16, 32, 48, 64 and 128 pixels, its corners clear

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

  Rule: The time spent is seen a week or a month at a time, every task of it

    @gui-a-deleted-tasks-time-says-so
    Scenario: the time of a deleted task says it was deleted, and a click on it says so
      Given I added a task titled "Write the spec"
      And I spent "1h30m" on task 1 on "2026-09-29"
      And today is "2026-09-30"
      And I deleted task 1
      When I open the window
      And I click the tab "TIME"
      And I click the card "Write the spec (deleted)" on the time page
      Then the window tells "task 1 was deleted"
      And no card is chosen

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

    @gui-the-time-month-is-chosen-from-the-twelve
    Scenario: the month and the year of the time page are buttons opening the grids the date picker has
      Given today is "2026-09-16"
      When I open the window
      And I click the tab "TIME"
      And I click "Month" on the time page
      And I click "the month" on the time page
      Then the time page offers the twelve months
      And every cell of the time page is drawn as a button
      When I click the cell "Mar" on the time page
      Then the month shown is "March 2026"
      When I click "the year" on the time page
      And I click "‹" on the time page
      And I click the cell "2014" on the time page
      Then the time page offers the twelve months
      When I click the cell "Mar" on the time page
      Then the month shown is "March 2014"

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

  Rule: Tasks are exported from the window, and imported into it, as a file

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

  Rule: The window's menu bar holds cairn and Help; every other act has its key, the bar under the board, the right click or the palette

    @gui-the-menu-bar-acts-from-the-board
    Scenario: a line of the menu bar does what its key does
      When I open the window
      And I pick "Help" in the menu bar
      Then the panel is headed "Help"

  Rule: A page turned off in the configuration is not drawn

    @gui-every-page-off-leaves-what-belongs-to-none
    Scenario: every page turned off leaves the menu bar nothing but the settings
      Given the pages turned off are "tasks, objectives, time"
      When I open the window
      Then the menu bar offers "Settings…"
      And the menu bar offers no "Help", "Upgrade", "About cairn"
      When I pick "Help" from the menu bar
      Then the window is on the board
      When I pick "Settings…" from the menu bar
      Then the panel is headed "Settings"

    @gui-the-first-page-shown-opens
    Scenario: with the tasks turned off the window opens on the objectives
      Given the pages turned off are "tasks"
      When I open the window
      Then the window is on the page "OBJECTIVES"

    @gui-every-page-off-is-an-empty-window
    Scenario: every page turned off opens a window that says so and leads to the settings
      Given the pages turned off are "tasks, objectives, time"
      When I open the window
      Then the window says every page is turned off
      When I press "cmd+n"
      Then the window is on the board
      When I click the offer to open the settings
      Then the panel is headed "Settings"
      When I press "esc"
      And I press "cmd+,"
      Then the panel is headed "Settings"

  Rule: Labels are seen and managed apart from any task

    @gui-a-label-is-deleted-everywhere
    Scenario: a label is taken off everything carrying it once that is agreed to
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 1 with the labels "doc" and "sp26-20"
      And I edited task 2 with the labels "release" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the label "sp26-20"
      And I press "cmd+backspace"
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

    @gui-a-label-is-made-from-the-labels
    Scenario: a label is made from the list of labels, with nothing carrying it
      When I open the window
      And I press "shift+l"
      Then the labels are listed
      And the bar offers "New label" under "n"
      When I press "n"
      And I type "Release Train"
      And I press "enter"
      Then the labels list "release-train" carried by 0 tasks and 0 objectives
      And the panel is headed "Labels"

    @gui-a-colour-is-walked-to-with-the-keys
    Scenario: the arrows walk the swatches and enter takes one, esc none
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      And I coloured the label "doc" "red"
      When I open the window
      And I press "shift+l"
      And I click the label "doc"
      And I press "c"
      And I press "right"
      And I press "esc"
      Then the label "doc" wears "red"
      When I press "c"
      And I press "right"
      And I press "enter"
      Then the label "doc" wears "maroon"

    @gui-the-labels-are-sought
    Scenario: the labels are narrowed by what is typed after a slash, and esc lets them all back
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "release"
      And I made the label "sp26-20"
      When I open the window
      And I press "shift+l"
      And I press "/"
      And I type "re"
      Then the labels in sight are "release"
      When I press "esc"
      Then the labels in sight are "doc, release, sp26-20"
      And the labels are listed

    @gui-a-label-sought-and-missing-is-made
    Scenario: a label sought that does not exist is offered to be made, and enter makes it
      When I open the window
      And I press "shift+l"
      And I press "/"
      And I type "Release Train"
      Then the labels offer to create "Release Train"
      When I press "enter"
      Then the labels list "release-train" carried by 0 tasks and 0 objectives

    @gui-the-chosen-label-carries-its-actions
    Scenario: the label chosen carries its actions in its row, and the foot only makes one or closes
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the label "sp26-20"
      Then the row of the label "sp26-20" offers "Rename", "Colour" and "Delete"
      And the row of the label "doc" offers nothing
      And the panel offers "New label" and "Close"
      When I click "Colour" in the row of the label "sp26-20"
      Then the labels offer 14 colours

    @gui-a-label-is-coloured-from-its-dot
    Scenario: the dot at the start of a label's row opens its colours
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I press "shift+l"
      And I click the dot of the label "doc"
      Then the labels offer 14 colours
      When I click the colour "peach"
      Then the label "doc" wears "peach"

  Rule: A new backup is made, and a backup restored, from the window

    @gui-putting-a-backup-back-from-the-window
    Scenario: a backup chosen from the list is restored once agreed to
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "cmd+b"
      And I added a task titled "Ship the board"
      And the window has looked at the store 1 time
      And I press "shift+b"
      Then the backups listed are 1
      And the panel is headed "Backups"
      And the bar offers "Restore" under "⏎"
      And the bar offers "Close" under "esc"
      When I press "enter"
      Then the window is asking
      When I press "enter"
      Then the window is on the board
      And "BACKLOG" carries "Write the spec"
      And "BACKLOG" does not carry "Ship the board"
      And the window tells "backup 1 was restored"
      And there are 2 backups

  Rule: A store behind the window is offered the changes it waits for

    @gui-a-store-from-a-newer-cairn-at-open
    Scenario: a window opened on a store from a newer cairn says so
      Given a store a newer cairn changed, holding a task titled "Write the spec"
      When I open the window
      Then the window says the store comes from a newer cairn
      And the only answer it offers is "Quit"

  Rule: A right click says what can be done to what is under it

    @gui-a-menu-acts
    Scenario: what is clicked in the menu is done to the card
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I right-click the card "Write the spec"
      And I click "Focus" in the menu
      Then task 1 is in focus
      And the window is on the board
      When I right-click the card "Write the spec"
      Then the menu offers "Open", "Edit", "New linked task", "Move to…", "Priority…", "Labels…", "Archive", "Unfocus", "Delete"

    @gui-a-recurring-cards-menu-skips-it
    Scenario: the menu of a card that comes back skips it, and its series makes the next
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I right-click the card "Stand-up"
      Then the menu offers "Open", "Edit", "New linked task", "Move to…", "Priority…", "Labels…", "Archive", "Focus", "Skip this one", "Delete"
      When I click "Skip this one" in the menu
      Then task 1 is archived
      And task 2 is in the column "todo"

    @gui-a-cards-menu-opens-a-linked-task
    Scenario: a card's menu opens a new task linked to it
      Given I added a task titled "Ship it" with the state "done"
      When I open the window
      And I right-click the card "Ship it"
      And I click "New linked task" in the menu
      And I set the field "title" to "It does not hold"
      And I choose "Save"
      Then task 2 is linked to task 1

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

    @gui-a-calendar-month-is-chosen-from-the-twelve
    Scenario: clicking the month in the calendar offers the twelve months
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      And I click the calendar's month
      Then the calendar offers the twelve months
      When I click the month "Mar"
      Then the calendar shows March

    @gui-a-calendar-year-is-chosen-from-a-grid
    Scenario: clicking the year in the calendar offers the years around it
      Given I added a task titled "Write the spec"
      And today is "2026-10-12"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "deadline"
      And I click the calendar's year
      Then the calendar offers the years around the one shown
      When I click the year "2028"
      Then the calendar offers the twelve months of 2028

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

  Rule: A form saves itself

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

    @gui-a-double-click-on-a-column-adds-there
    Scenario: a double click where a column holds no card adds a task to it
      When I open the window
      And I double-click the column "TODO"
      Then the window is on the form
      When I type "Plan the week"
      And I choose "Save"
      Then "TODO" carries "Plan the week"

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

    @gui-a-task-added-between-two-cards-keeps-the-column-chosen
    Scenario: a task added between two cards and given another column stays in that column
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship the board" with the state "todo"
      When I open the window
      And I double-click between the cards "Write the spec" and "Ship the board"
      And I type "Plan the week"
      And I click the field "column"
      And I press "down"
      And I press "enter"
      And I choose "Save"
      Then "CURRENT" carries "Plan the week"

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

    @gui-a-new-task-with-no-title-is-not-made
    Scenario: a form adding a task with no title yet makes nothing
      When I open the window
      And I press "cmd+n"
      And the window has looked at the store 3 times
      Then there are no tasks

  Rule: What the board can do, the window can do

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

    @gui-two-ways-into-the-search
    Scenario: the search opens the way a person expects
      When I open the window
      And I press "cmd+f"
      Then the search box is open
      And I press "esc"
      And I press "/"
      Then the search box is open

  Rule: The window follows the store

    @gui-somebody-else-added-something
    Scenario: a task added elsewhere appears
      When I open the window
      And somebody else adds a task titled "Client's answer"
      And the window looks again
      Then "BACKLOG" carries "Client's answer"

  Rule: A task's trace stays inside the panel that shows it

    @gui-opening-a-task-opens-it-to-be-edited
    Scenario: opening a task opens it to be edited, its trace a tab away
      Given I added a task titled "Write the spec"
      And it was worked on 3 times
      When I open the window
      And I double-click the card "Write the spec"
      Then the window is on the form
      When I press "cmd+t"
      Then the trace begins with the last thing done to task 1

  Rule: What the arrows reach, they bring into sight

    @gui-the-arrows-bring-an-objective-into-sight
    Scenario: walking down the objectives brings the next one into sight
      Given I added 12 objectives
      When I open the window
      And I click the tab "OBJECTIVES"
      And I press "down" 11 times
      Then what is chosen is in sight

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

  Rule: An objective opens on the work that serves it

    @gui-an-objectives-work-is-walked-with-the-keyboard
    Scenario: tab from the description reaches the work, the arrows walk it, enter opens a task
      Given I added an objective titled "Ship the board"
      And I added a task titled "Write the spec" with the state "current" serving objective 1
      And I added a task titled "Ship it" with the state "todo" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the board"
      And I press "enter"
      And I put the keyboard on "description"
      And I press "tab"
      Then the keyboard is on the work "Write the spec"
      When I press "down"
      Then the keyboard is on the work "Ship it"
      When I press "down"
      Then the keyboard is on the work "+ New task serving it"
      When I press "up"
      And I press "enter"
      Then the form is open on task 2

    @gui-an-objective-makes-a-task-that-serves-it
    Scenario: the last line under an objective's work opens a new task serving it
      Given I added an objective titled "Ship the board"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship the board"
      And I press "enter"
      And I click the work "+ New task serving it"
      Then the window is on the form
      And the field "objectives" names "Ship the board"

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

    @gui-a-list-is-corrected-like-any-other-field
    Scenario: what was typed into a list can be taken back out
      Given I added a task titled "Write the spec" in "todo" at the priority "normal" carrying the label "backend"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "labels"
      And I type "backendx"
      Then the list does not offer "backend"
      When I press "backspace"
      Then the list offers "backend"

    @gui-a-triple-click-holds-the-paragraph
    Scenario: a triple click holds the whole paragraph it lands in
      Given I added a task titled "Write the spec" with the paragraphs "Two columns and a band" and "Then the rest"
      When I open the window
      And I double-click the card "Write the spec"
      And I triple-click the first word of "field:description" on the open task
      And I press "cmd+c"
      Then the clipboard reads "Two columns and a band"

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

    @gui-a-click-in-a-list-leaves-the-form-under-it-alone
    Scenario: a click inside the list does not reach the form behind it
      Given I added a task titled "Write the spec" in "todo" at the priority "normal" carrying the label "backend"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "labels"
      And I click the line that narrows the list
      Then the window is choosing
      And the form is still on the field "labels"

    @gui-a-click-off-a-list-closes-it
    Scenario: a click off the list closes it, as Cancel does, and leaves the form open
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I choose "Edit"
      And I click the field "priority"
      And I click the field "title"
      Then the window is on the form
      And the field "priority" reads "normal"

    @gui-the-count-of-labels-opens-them-all
    Scenario: resting the pointer on a card's count of labels shows every label it carries
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "an-altogether-overlong-label" and "another-just-as-long-one"
      When I open the window
      And I rest the pointer on the count of labels of "Write the spec"
      Then a bubble shows the labels "an-altogether-overlong-label", "another-just-as-long-one"
      When I rest the pointer on the title of "Write the spec"
      Then no bubble shows labels

  Rule: A key does what the bar offers, where the bar offers it

    @gui-a-narrow-bar-keeps-its-line
    Scenario: at the narrowest window the bar's hints give their words up, so its line stays
      Given a window of 720 by 560
      And I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      Then the command bar reads "#1 · Write the spec"
      And the bar's line has room to be read
      And every hint of the bar is its key alone, inside the window

    @gui-esc-lets-the-chosen-card-go
    Scenario: esc on the board lets go of the card chosen
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      Then the card "Write the spec" is chosen
      When I press "esc"
      Then no card is chosen
      And the command bar reads "Command, number or title…"

    @gui-the-palette-takes-a-command-typed-whole
    Scenario: a command typed whole in the command bar runs as if its row had been chosen
      Given I added a task titled "Write the spec" with the state "todo"
      And I added a task titled "Ship it" with the state "todo"
      When I open the window
      And I click the card "Write the spec"
      And I press "cmd+k"
      And I type "move to done"
      And I press "enter"
      Then task 1 is in the column "done"
      And the command bar is shut
      When I press "cmd+k"
      And I type "priority of #2 high"
      And I press "enter"
      Then task 2 carries the priority "high"

    @gui-the-palette-opens-a-task-found-by-title
    Scenario: the command bar finds a task by its title and opens it
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the window
      And I press "cmd+k"
      And I type "ship"
      Then the command bar offers the task "Ship it"
      When I press "down"
      And I press "enter"
      Then the window is on the form

    @gui-a-search-kept-is-a-pill
    Scenario: enter keeps the search as a pill above the board, and its cross clears it
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "enter"
      Then the search "spec" is kept as a pill
      When I click the cross of the search's pill
      Then no search is kept

    @gui-a-number-opens-its-task
    Scenario: a task's number typed in the command bar opens that task, with or without its #
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the window
      And I press "cmd+k"
      And I type "#2"
      Then the command bar offers "Open #2 · Ship it" first
      When I press "enter"
      Then the form is open on task 2
      And the command bar is shut
      When I press "esc"
      And I press "cmd+k"
      And I type "1"
      And I press "enter"
      Then the form is open on task 1

    @gui-esc-on-a-search-goes-back-where-it-came-from
    Scenario: esc on a search goes back to where the search was started, and never into the command bar from the board
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "/"
      And I type "spec"
      And I press "esc"
      Then the window is on the board
      And no search is kept
      When I press "cmd+k"
      And I type "/spec"
      And I press "esc"
      Then the command bar is open

    @gui-m-moves-the-chosen-task
    Scenario: m opens the columns under the chosen card, and a column's letter moves it there
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the window
      And I click the card "Write the spec"
      And I press "m"
      Then the menu offers "backlog", "current", "blocked", "watching", "paused", "done"
      When I press "d"
      Then the engine was asked "task edit" for task 1

    @gui-p-sets-the-priority
    Scenario: p opens the priorities under the chosen card
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And I press "p"
      And I press "down"
      And I press "enter"
      Then the engine was asked "task edit" for task 1

    @gui-l-picks-labels-under-the-card
    Scenario: l opens the labels under the chosen card, with no form, and each one taken is written at once
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      And I made the label "release"
      When I open the window
      And I click the card "Write the spec"
      And I press "l"
      Then the labels are picked under the card "Write the spec"
      When I type "rel"
      And I press "enter"
      Then task 1 carries exactly the labels "doc, release, sp26-20"
      When I press "esc"
      Then the window is on the board
      And task 1 carries exactly the labels "doc, release, sp26-20"

    @gui-the-labels-taken-stand-as-pills
    Scenario: the labels a task carries stand as pills at the head of the list, a cross or backspace taking one off
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "doc" and "sp26-20"
      When I open the window
      And I click the card "Write the spec"
      And I press "l"
      Then the labels taken read "doc, sp26-20"
      When I click the cross of the label taken "doc"
      Then task 1 carries exactly the labels "sp26-20"
      When I press "backspace"
      Then task 1 carries exactly the labels ""
      And the labels taken read ""

    @gui-g-goes-to-a-page
    Scenario: g then a page's letter goes to that page
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "g"
      And I press "o"
      Then the window is on the page "OBJECTIVES"
      When I press "g"
      And I press "i"
      Then the window is on the page "TIME"
      When I press "2"
      Then the time page shows a week
      When I press "3"
      Then the time page shows a month

    @gui-every-key-the-bar-offers-works
    Scenario: every key the bar offers does what the bar says, on every screen
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      Then every key the bar offers does what it offers

  Rule: A task opens on its steps, ticked, renamed, added and taken away where they are

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

    @gui-a-question-answers-to-its-verb
    Scenario Outline: a question's answer is named by what it does, and enter does it
      Given I added a task titled "Write the spec" with the state "done"
      When I open the window
      And I click the card "Write the spec"
      And I press "<key>"
      Then the panel is headed "<question>"
      And the bar offers "<verb>" under "⏎"
      And the bar offers "Cancel" under "esc"
      And the question says nothing of how to answer it
      When I press "enter"
      Then the board says nothing of "Write the spec"

      Examples:
        | key           | question                      | verb    |
        | cmd+backspace | Delete task 1?                | Delete  |
        | shift+x       | Archive every task in DONE?   | Archive |

    @gui-a-question-on-a-task-says-what-goes-with-it
    Scenario: deleting a task says the steps and the time that go with it
      Given I added a task titled "Write the spec"
      And I added the step "Draft" to task 1
      And I spent "1h30m" on task 1 on "2026-10-30"
      When I open the window
      And I click the card "Write the spec"
      And I press "cmd+backspace"
      Then the question says "Its 1 step, its labels and its 1h30m spent go with it."

    @gui-space-ticks-every-step-taken
    Scenario: space ticks every step taken, and unticks them once all are done
      Given I added a task titled "Ship it"
      And I added the step "Draft" to task 1
      And I added the step "Keep" to task 1
      And I added the step "Review" to task 1
      When I open the window
      And I double-click the card "Ship it"
      And I click the step "Draft"
      And I ⇧-click the step "Review"
      And I press "space"
      Then task 1 carries the step "Draft", done
      And task 1 carries the step "Keep", done
      And task 1 carries the step "Review", done
      When I press "space"
      Then the steps of task 1 read "Draft", "Keep", "Review"
      And task 1 carries the step "Keep", not done

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

  Rule: The objectives a task serves are picked from a list, not typed

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

  Rule: The key that opens a field does not write in it

    @gui-the-key-that-opens-a-form-writes-nothing-in-it
    Scenario: the key that opens the form does not land in the field
      Given I added a task titled "Write the spec"
      When I open the window
      And I click the card "Write the spec"
      And the keyboard says "e" as both a key and as text
      Then the form is open on task 1
      And the field "title" reads "Write the spec"

  Rule: Losing work is asked about first

    @gui-every-panel-is-closed-by-its-cross
    Scenario Outline: the cross at the top of a panel closes it, as its esc does
      Given I added a task titled "Write the spec"
      When I open the window
      And I press "<key>"
      And I click the panel's cross
      Then the window is on the board

      Examples:
        | key     |
        | ?       |
        | shift+f |
        | cmd+,   |

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

  Rule: One thing is drawn one way, wherever it stands

    @gui-the-count-of-labels-is-a-plain-pill
    Scenario: the count of the labels a card leaves out is a pill like any other, with no ring
      Given I added a task titled "Write the spec"
      And I edited task 1 with the labels "an-altogether-overlong-label" and "another-just-as-long-one"
      When I open the window
      Then the card "Write the spec" counts the labels it leaves out in a pill with no ring

  Rule: Several cards are taken at once

    @gui-what-takes-more-can-be-left-out
    Scenario: what would take more with it is named, and can be left out
      Given I added a task titled "Ship it"
      And I added the step "Draft the columns" to task 1
      And I added a task titled "Other"
      When I open the window
      And I click the card "Ship it"
      And I ⌘-click the card "Other"
      And I press "cmd+backspace"
      Then the question says "#1  Ship it — its 1 step goes with it"
      When I choose "Skip flagged"
      Then "BACKLOG" carries "Ship it"
      And "BACKLOG" does not carry "Other"

    @gui-what-cannot-go-is-left-out
    Scenario: an objective work still serves is left out, and the rest goes
      Given I added an objective titled "Ship cairn"
      And I added an objective titled "Keep it honest"
      And I added a task titled "Write the spec" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Ship cairn"
      And I ⌘-click the objective "Keep it honest"
      And I press "cmd+backspace"
      Then the question says "#1  Ship cairn — 1 task serves it: it cannot go"
      When I choose "Delete"
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

  Rule: The settings are changed in place, and read again when they change

    @gui-the-configuration-is-shown
    Scenario: ⌘, shows the configuration in use, key by key, and where each comes from
      Given a configuration file that names no workspace
      When I open the window
      And I press "cmd+,"
      Then the panel is headed "Settings"
      And the settings are written in "All workspaces"
      And the configuration reads "store" read when cairn opens
      And the configuration reads "plugin" from its default
      And the configuration reads "time.weekends" from its default

    @gui-the-configuration-file-is-opened-from-the-window
    Scenario: the file is handed to the system to open
      Given a configuration file that names no workspace
      When I open the window
      And I press "cmd+,"
      And I choose "Open the file"
      Then the window asks the system to open the configuration file

    @gui-a-missing-configuration-is-written-from-the-window
    Scenario: with no file, the window writes the whole of one
      When I open the window
      And I press "cmd+,"
      And I choose "Write the whole file"
      Then a configuration file exists
      When I choose the scope "All workspaces"
      Then the configuration reads "plugin" from the file

    @gui-settings-a-toggle-is-written-at-once
    Scenario: a switch clicked is written in the settings of the workspace at once, and the window follows it
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      Then the settings are written in "This workspace · home"
      And the configuration reads "series.on" from its default
      When I click the switch of "series.on"
      Then the settings of the workspace "home" say `on = true`
      And the configuration reads "series.on" only here
      And the banner says nothing of a change
      When I press "esc"
      Then the tabs read "TASKS", "OBJECTIVES", "TIME", "SERIES"

    @gui-settings-a-choice-is-written-in-the-global-configuration
    Scenario: with the global configuration chosen, a choice of three is written there
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I choose the scope "All workspaces"
      And I click "Counted only" on the row "pages.time"
      Then the global configuration says `time = "hidden"`
      And the settings of the workspace "home" do not say `time = "hidden"`
      And the configuration reads "pages.time" from the file

    @gui-settings-reset-takes-the-line-out
    Scenario: Use the shared value takes the key out of the settings of the workspace, and the value of all workspaces answers again
      Given the workspace "home", the default
      And the settings of the workspace "home" turn the time off
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I click "Use the shared value" on the row "pages.time"
      Then the settings of the workspace "home" do not say `time = "off"`
      And the configuration reads "pages.time" from its default
      And the tabs read "TASKS", "OBJECTIVES", "TIME"

    @gui-settings-set-here-pins-the-global-value
    Scenario: a key from the global configuration is set here with the value it has, under the pointer
      Given the workspace "home", the default
      And the global configuration says the archive lasts 30 days
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I click "Set only here" on the row "archive.days"
      Then the settings of the workspace "home" say `days = 30`
      And the configuration reads "archive.days" only here

    @gui-settings-a-field-is-written-on-enter
    Scenario: a field is written when enter is pressed, and a value it does not take is said under it and not written
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I write "1 hour" in the row "focus.session"
      Then the row "focus.session" says `focus.session: "1 hour" is not a duration — write 1h, 45m, 1h30m`
      And the settings of the workspace "home" do not say `session = "1 hour"`
      When I write "45m" in the row "focus.session"
      Then the settings of the workspace "home" say `session = "45m"`
      And the configuration reads "focus.session" only here

    @gui-settings-places-are-read-when-cairn-opens
    Scenario: the store and the notes are shown and never written from the window
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I choose the scope "All workspaces"
      Then the configuration reads "store" read when cairn opens
      And the configuration reads "notes" read when cairn opens
      And the row "store" offers nothing to change

    @gui-settings-the-theme-is-previewed-then-worn
    Scenario: the theme chooser re-skins the window as the keyboard walks it, wears the scheme on enter, and esc goes back
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I click the control of "window.theme"
      Then the panel is headed "Theme of home"
      When I walk the chooser to "nord"
      Then the window wears "nord"
      When I press "esc"
      Then the window wears "default"
      And the panel is headed "Settings"
      When I click the control of "window.theme"
      And I walk the chooser to "gruvbox-dark"
      And I press "enter"
      Then the window wears "gruvbox-dark"
      And the settings of the workspace "home" say `theme = "gruvbox-dark"`
      And the configuration reads "window.theme" only here

    @gui-settings-the-intro-is-played-then-kept
    Scenario: the intro chooser plays each intro on the board behind as the keyboard walks it, replays it, keeps one on enter, and esc goes back
      Given the workspace "home", the default
      And the workspace "home" holds a task titled "Write the spec"
      When I open the window on the workspace "home"
      And I press "cmd+,"
      And I click the control of "window.intro"
      Then the panel is headed "Intro of home"
      When I walk the intro chooser to "deck"
      Then the board is arriving by "deck"
      When the window draws a frame 2 seconds later
      Then no cairn is being built
      When I press "r"
      Then the board is arriving by "deck"
      When I press "esc"
      Then the panel is headed "Settings"
      And no cairn is being built
      And the settings of the workspace "home" do not say `intro = "deck"`
      When I click the control of "window.intro"
      And I walk the intro chooser to "flap"
      And I press "enter"
      Then the settings of the workspace "home" say `intro = "flap"`
      And the configuration reads "window.intro" only here

    @gui-settings-a-file-that-cannot-be-read-keeps-the-last-good
    Scenario: settings that cannot be read leave the last good ones in force, said until they read
      Given the workspace "home", the default
      And the settings of the workspace "home" turn the time off
      When I open the window on the workspace "home"
      And the settings of the workspace "home" are written by hand:
        """
        [pages]
        time = "hiden"
        """
      And the window looks again
      Then the banner says "settings.toml cannot be read"
      And the banner says "line 2"
      And the tabs read "TASKS", "OBJECTIVES"
      When I press "cmd+,"
      Then the row "pages.time" is not offered
      When the settings of the workspace "home" are written by hand:
        """
        [pages]
        time = "on"
        """
      And the window looks again
      Then the banner says "pages.time off → on"
      And the banner says nothing of a file that cannot be read
      And the tabs read "TASKS", "OBJECTIVES", "TIME"

    @gui-settings-rows-say-what-they-do
    Scenario: each row is named in plain words, says what it does, and shows its key only under the pointer
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      Then the sections read "Look", "Tabs", "Time", "Archive", "Agent", "Files"
      And the row "window.keep_sizes" is named "Remember column widths" over "The widths you drag stay the next time"
      When I rest the pointer on the row "window.keep_sizes"
      Then the row "window.keep_sizes" is named "Remember column widths" over "window.keep_sizes"

    @gui-settings-the-opacity-is-a-slider
    Scenario: dragged card opacity is a slider, written where it is let go, the arrows moving it by five
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      Then the row "window.ghost" shows 65 %
      When I drag the slider of "window.ghost" to 40 %
      Then the settings of the workspace "home" say `ghost = 40`
      When I walk the settings to "window.ghost"
      And I press "right"
      Then the settings of the workspace "home" say `ghost = 45`

    @gui-settings-the-time-of-deleted-tasks-is-a-switch
    Scenario: whether a deleted task's time is still counted is one switch, written in the words of the file
      Given the workspace "home", the default
      When I open the window on the workspace "home"
      And I press "cmd+,"
      Then the row "time.deleted" is named "Count the time of deleted tasks" over "On: the time spent on a task you delete stays in the Time tab. Off: it goes with the task."
      When I click the switch of "time.deleted"
      Then the settings of the workspace "home" say `deleted = "taken"`
      And the configuration reads "time.deleted" only here

  Rule: The window names its workspace, and opens another in a window of its own

    @gui-workspace-the-title-and-the-pill-name-it
    Scenario: the window's title and the pill opening the top bar name the workspace, in its colour
      Given the workspace "work", the default
      And the settings of the workspace "work" say its colour is "teal"
      When I open the window on the workspace "work"
      Then the window is titled "cairn — work"
      And the top bar opens with the workspace "work", before the tabs
      And the dot of the pill wears "teal"

    @gui-workspace-the-switcher-lists-them
    Scenario: the pill drops the switcher: the workspaces, each saying what choosing it does, then the recent folders
      Given the workspace "work", the default
      And the workspace "home"
      And the workspace "cairn-dev"
      And the recent folder "scratch", opened yesterday
      And the recent folder "trip", whose store is gone
      And a window is already open on the workspace "cairn-dev"
      When I open the window on the workspace "work"
      And I click the pill of the workspace
      Then the switcher lists the workspaces "cairn-dev", "home", "work"
      And the workspace "cairn-dev" is tagged "open · brings it forward" and "⌘⇧1"
      And the workspace "home" is tagged "opens a new window" and "⌘⇧2"
      And the workspace "work" is tagged "this window · default" and "⌘⇧3"
      And the recent folder "scratch" is tagged "opened yesterday"
      And the recent folder "trip" is tagged "not found now"
      And the switcher ends on "Open a folder…", "New workspace…", "Settings of work…"

    @gui-workspace-another-opens-in-a-window-of-its-own
    Scenario: choosing another workspace opens it in a window of its own, and this one stays
      Given the workspace "work", the default
      And the workspace "home"
      When I open the window on the workspace "work"
      And I click the pill of the workspace
      And I click the workspace "home" in the switcher
      Then a cairn is started with "gui --workspace home"
      And the window says "home opens in a window of its own"
      And the window is titled "cairn — work"

    @gui-workspace-a-folder-opens-in-a-window-of-its-own
    Scenario: a folder typed opens in a window of its own; a folder that is not there is refused
      Given the workspace "work", the default
      And a store no workspace names, in the folder "scratch"
      When I open the window on the workspace "work"
      And I press "cmd+shift+o"
      And I click "Open a folder…" in the switcher
      And I type the folder "elsewhere"
      And I press "enter"
      Then the window says "is not there"
      And no cairn is started
      When I press "esc"
      And I click "Open a folder…" in the switcher
      And I type the folder "scratch"
      And I press "enter"
      Then a cairn is started with "gui --store" and the store of the folder "scratch"

    @gui-workspace-a-new-one-is-made-and-opened
    Scenario: New workspace… opens a dialog asking its name, its folder and its colour, and the workspace made opens in a window of its own
      Given the workspace "work", the default
      When I open the window on the workspace "work"
      And I press "cmd+shift+o"
      And I click "New workspace…" in the switcher
      Then the dialog "New workspace" asks "Name", "Folder", "Colour" and "Switch to it now"
      When I type "garden"
      Then the folder of the new workspace reads "~/.cairn/garden"
      When I click the field "Folder" of the dialog
      And I type the folder "cairn/garden"
      And I click the colour "teal" in the dialog
      And I click "Create" in the dialog
      Then the configuration names the workspace "garden"
      And the settings of the workspace "garden" say `colour = "teal"`
      And a cairn is started with "gui --workspace garden"

    @gui-workspace-a-new-one-is-chosen-a-folder-and-kept-apart
    Scenario: the folder is chosen with the system's dialog, and a workspace made without switching stays closed
      Given the workspace "work", the default
      When I open the window on the workspace "work"
      And I press "cmd+shift+o"
      And I click "New workspace…" in the switcher
      And I type "garden"
      And I click "Choose…" in the dialog
      And I pick the folder "cairn/garden" in the system's dialog
      And I click "Switch to it now" in the dialog
      And I click "Create" in the dialog
      Then the configuration names the workspace "garden"
      And no cairn is started
      And the window says "garden is made"

    @gui-workspace-a-new-one-refused-stays-asked
    Scenario: a name cairn does not take is said in the dialog, and Cancel makes nothing
      Given the workspace "work", the default
      When I open the window on the workspace "work"
      And I press "cmd+shift+o"
      And I click "New workspace…" in the switcher
      And I type "My Garden"
      And I click "Create" in the dialog
      Then the dialog "New workspace" says "is named in lowercase letters"
      When I click "Cancel" in the dialog
      Then the configuration does not name the workspace "My Garden"
      And no cairn is started

    @gui-workspace-the-switcher-opens-the-settings
    Scenario: the switcher's last line opens the settings of this workspace
      Given the workspace "work", the default
      When I open the window on the workspace "work"
      And I click the pill of the workspace
      And I click "Settings of work…" in the switcher
      Then the panel is headed "Settings"
      And the settings are written in "This workspace · work"

  Rule: The objectives tab is a tree beside the objective chosen

    @gui-objective-a-task-in-the-detail-opens
    Scenario: a task clicked in the detail of an objective or a sub-objective opens over the tab, and closing it comes back to that objective
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Garden"
      And I added a task titled "Buy the relay" serving objective 1
      And I added a task titled "Flash the bootloader" serving objective 2
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I click the task "Buy the relay" in the detail
      Then the form is open on task 1
      When I press "esc"
      Then the detail is headed "Home automation"
      When I click the task "Flash the bootloader" in the detail
      Then the form is open on task 2
      When I press "esc"
      Then the detail is headed "Home automation"

    @gui-objective-the-detail-says-what-waits-on-what
    Scenario: an objective's detail says what it waits on, and what waits on it
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Features"
      Then the detail says "#1 Firmware"
      When I click the objective "Firmware"
      Then the detail says "#2 Features"

    @gui-objective-a-parent-is-proposed-as-reached
    Scenario: a parent whose every sub-objective is reached is proposed as reached, and left where it stands
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And objective 2 is reached
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      Then the detail says "every sub-objective is reached: it may be put reached too"
      And objective 1 stands "active"

    @gui-objective-the-tree-is-lit-under-the-pointer
    Scenario: an objective of the tree, a tile and an offer of the detail are lit under the pointer
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I click the chevron of "Home automation"
      And I rest the pointer on the objective "Enclosure"
      Then the objective "Enclosure" is lit under the pointer
      When I rest the pointer on the tile of "Enclosure"
      Then the tile of "Enclosure" is lit under the pointer
      When I rest the pointer on the detail's offer "Edit"
      Then the detail's offer "Edit" is lit under the pointer

  Rule: A parent of the objectives tab folds its sub-objectives away

    @gui-objective-shift-arrows-reorder
    Scenario: shift with up or down moves the chosen objective among its siblings, a parent with its sub-objectives, and stops at the ends
      Given I added an objective titled "Home automation"
      And I added an objective titled "Tidy the workshop"
      And I added an objective titled "Firmware" under objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I press "shift+down"
      Then the objectives read "Tidy the workshop", "Home automation"
      And the objective "Home automation" is chosen
      When I press "right"
      Then the objectives read "Tidy the workshop", "Home automation", "Firmware"
      When I press "shift+down"
      Then the objectives read "Tidy the workshop", "Home automation", "Firmware"
      When I press "shift+up"
      And I press "shift+up"
      Then the objectives read "Home automation", "Firmware", "Tidy the workshop"

    @gui-objective-dragged-reorders-and-changes-parent
    Scenario: an objective dragged above or below another takes that place, and dropped onto a parent goes under it
      Given I added an objective titled "Home automation"
      And I added an objective titled "Tidy the workshop"
      And I added an objective titled "Garden"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I drag the objective "Garden" above "Home automation"
      Then the objectives read "Garden", "Home automation", "Tidy the workshop"
      When I drag the objective "Garden" below "Tidy the workshop"
      Then the objectives read "Home automation", "Tidy the workshop", "Garden"
      When I drag the objective "Garden" onto "Home automation"
      And I click the objective "Home automation"
      And I press "right"
      Then the objectives read "Home automation", "Garden", "Tidy the workshop"

    @gui-objective-a-folded-parent-opens-under-a-drag
    Scenario: a folded parent held under a dragged objective unfolds after a moment, and folds back when the drop lands elsewhere
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Garden"
      And I added an objective titled "Tidy the workshop"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I hold the objective "Garden" over "Home automation" for 700 ms
      Then the objectives read "Home automation", "Firmware", "Garden", "Tidy the workshop"
      And the landing says "under #1"
      When I let the objective go below "Tidy the workshop"
      Then the objectives read "Home automation", "Tidy the workshop", "Garden"

    @gui-objective-x-chooses-the-level-under-a-last-sub-objective
    Scenario: under a parent's last sub-objective, the pointer left of the sub-objectives drops at level one
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Garden"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I press "right"
      And I hold the objective "Garden" below "Firmware", left of its indent
      Then the landing says "after #1"

    @gui-objective-a-parent-dragged-onto-another-is-refused
    Scenario: a parent dropped onto another objective is refused, and nothing moves
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Garden"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I drag the objective "Home automation" onto "Garden"
      Then the objectives read "Home automation", "Garden"

    @gui-objective-a-match-chosen-keeps-its-parent-open
    Scenario: a sub-objective chosen during a search keeps its parent unfolded once the search is cleared
      Given I added an objective titled "Home automation"
      And I added an objective titled "Tidy the workshop"
      And I added an objective titled "Firmware" under objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I press "/"
      And I type "firm"
      And I click the objective "Firmware"
      And I press "esc"
      Then the objectives read "Home automation", "Firmware", "Tidy the workshop"
      And the objective "Firmware" is chosen

    @gui-objective-a-folded-parent-says-what-it-hides
    Scenario: a folded parent says which hidden sub-objective is late or waits, the late first and the rest counted, and unfolded says nothing of it
      Given today is "2026-11-10"
      And I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added an objective titled "Documentation" under objective 1
      And objective 3 waits on objective 2
      And objective 4 is due "2026-11-06"
      When I open the window
      And I click the tab "OBJECTIVES"
      Then the objective "Home automation" ends in the late ink with "#4 Documentation late +1"
      When I rest the pointer on what "Home automation" hides
      Then the bubble holds the line "#4 Documentation, its deadline Fri 6 Nov past, not reached."
      When I click the chevron of "Home automation"
      Then the objective "Home automation" says nothing in the late ink
      And the objective "Documentation" ends in the late ink with "6 Nov, late"
      And the objective "Firmware" ends in the late ink with "waits on Enclosure"

  Rule: Show its tasks narrows the tasks board to a subtree

    @gui-objective-t-shows-its-tasks
    Scenario: t on an objective shows on the tasks board its tasks and those of its sub-objectives
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Vent slots" serving objective 2
      And I added a task titled "Buy the relay" serving objective 1
      And I added a task titled "Buy a lamp"
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I press "t"
      Then the tab is "TASKS"
      And the board says it is filtered by "objective #1 Home automation, with its sub-objectives"
      And "BACKLOG" carries "Vent slots"
      And "BACKLOG" carries "Buy the relay"
      And "BACKLOG" does not carry "Buy a lamp"

  Rule: An objective's form says where it sits and how far along it is

    @gui-objective-the-form-offers-level-one-parents
    Scenario: the parent of an objective is chosen among the objectives of level one, or none
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware"
      And I added an objective titled "Enclosure" under objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Firmware"
      And I choose "Edit"
      And I put the keyboard on "parent"
      And I press "enter"
      Then the list reads "none", "Home automation"
      When I press "down"
      And I press "enter"
      And I choose "Save"
      Then objective 2 sits under objective 1

  Rule: A task's objectives are a tree, and counting one twice is refused where it is chosen

    @gui-objective-the-task-form-refuses-a-double-count
    Scenario: ticking a parent beside its sub-objectives is refused at the foot of the list, and nothing changes
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Pin map" serving objectives 2 and 3
      When I open the window
      And I click the card "Pin map"
      And I choose "Edit"
      And I put the keyboard on "objectives"
      And I press "enter"
      And I press "up"
      And I press "enter"
      Then the list says "#1 is already counted in #1 Home automation through #2 Enclosure and #3 Firmware"
      And the list says under it "Untick Enclosure and Firmware to put it on Home automation itself. Nothing was changed."
      And objective 1 is not ticked
      And objective 2 is ticked

  Rule: Deleting an objective asks what goes with it, and decides nothing

    @gui-objective-deleting-asks-row-by-row
    Scenario: deleting an objective holding sub-objectives and tasks asks about each, and deletes once each is chosen
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "OTA updates" serving objective 3
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I choose "Delete"
      Then the deletion asks about "#2 Enclosure" and "#3 Firmware"
      And the deletion cannot be confirmed yet
      When I choose "Keep as level one" for "#2 Enclosure"
      And I choose "Delete" for "#3 Firmware"
      Then the deletion asks about the tasks "#1"
      And the deletion cannot be confirmed yet
      When I choose "Detach" for the tasks
      Then the deletion is confirmed by "Delete #1 and #3"
      When I confirm the deletion
      Then objective 1 is gone
      And objective 2 stands on its own
      And objective 3 is gone
      And task 1 serves nothing

    @gui-objective-a-deletion-asked-is-kept-on-escape
    Scenario: escape keeps an objective whose deletion was asked about
      Given I added an objective titled "Home automation"
      And I added a task titled "Buy the relay" serving objective 1
      When I open the window
      And I click the tab "OBJECTIVES"
      And I click the objective "Home automation"
      And I choose "Delete"
      And I press "esc"
      Then objective 1 is still there
      And task 1 serves objective 1

  Rule: Work started early is warned about

    @gui-objective-work-started-early-is-warned-about
    Scenario: putting a task in focus says when its objective waits on one not reached
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      And I added a task titled "OTA updates" serving objective 2
      When I open the window
      And I click the card "OTA updates"
      And I press "f"
      Then the window tells "#2 Features waits on #1 Firmware, not reached"

  Rule: A view keeps the board's filters under a name, and opens them again in one act

    Background:
      Given I added a task titled "Write the spec" in "todo" at the priority "high" carrying the label "backend"
      And I added a task titled "Review the plan" in "todo" at the priority "normal" carrying the label "backend"
      And I added a task titled "Ship it" in "current" at the priority "high" carrying the label "sprint"

    @gui-view-kept-from-the-filters
    Scenario: the filters in force are kept as a view from the foot of the filter menu, and fold into its pill
      When I open the window
      And I set the filter "label" to "backend"
      And I press "v"
      Then the name proposed for the view is "backend"
      When I press "cmd+a"
      And I type "Backend work"
      And I press "enter"
      Then the view "Backend work" holds the label "backend"
      And the pills after the tabs read "view Backend work, + Filter"
      And the tab "Tasks" counts "2/3"

    @gui-view-kept-from-the-command-bar
    Scenario: the command bar keeps the filters in force as a view named there
      When I open the window
      And I set the filter "priority" to "high"
      And I press "esc"
      And I press "cmd+k"
      And I type "view new Urgent"
      And I press "enter"
      Then the view "Urgent" holds the priority "high"
      And the pills after the tabs read "view Urgent, + Filter"

    @gui-view-opened-by-its-key
    Scenario: ⌘ and a view's place opens it, and ⌘0 goes back to the whole board
      Given I kept the view "Sprint" holding the label "sprint"
      When I open the window
      And I press "cmd+1"
      Then "current" carries "Ship it"
      And "todo" does not carry "Write the spec"
      And the tab "Tasks" counts "1/3"
      When I press "cmd+0"
      Then "todo" carries "Write the spec"
      And the pills after the tabs read "+ Filter"

    @gui-view-listed-at-the-head-of-the-filters
    Scenario: the filter menu lists the views at its head, each with how many tasks it keeps and its key
      Given I kept the view "Sprint" holding the label "sprint"
      And I kept the view "High" holding the priority "high"
      When I open the window
      And I press "shift+f"
      Then the filter menu lists the views "Sprint 1 ⌘1", "High 2 ⌘2"
      When I click the view "High"
      Then "todo" carries "Write the spec"
      And "todo" does not carry "Review the plan"

    @gui-view-a-filter-on-top-marks-it-changed
    Scenario: a filter added on top of a view marks its pill changed, and is saved into it when asked
      Given I kept the view "High" holding the priority "high"
      When I open the window
      And I press "cmd+1"
      And I set the filter "label" to "backend"
      And I press "esc"
      Then the pill of the view is marked changed
      And "current" does not carry "Ship it"
      When I click the pill "view High"
      And I choose "Save to the view"
      Then the view "High" holds the label "backend"
      And the pill of the view is not marked changed

    @gui-view-the-panel-renames-moves-and-deletes
    Scenario: the views panel renames a view, moves it to another place, and deletes it
      Given I kept the view "Sprint" holding the label "sprint"
      And I kept the view "High" holding the priority "high"
      When I open the window
      And I press "cmd+k"
      And I type "views…"
      And I press "enter"
      Then the views panel lists "Sprint", "High"
      When I press "shift+down"
      Then the views read "High", "Sprint"
      When I press "r"
      And I press "cmd+a"
      And I type "Now"
      And I press "enter"
      Then the views read "High", "Now"
      When I press "cmd+backspace"
      Then the views read "High"

    @gui-view-gone-is-said
    Scenario: a view whose label is gone says so in the panel, and still opens
      Given I kept the view "Sprint" holding the label "sprint"
      And I kept the view "High" holding the priority "high"
      And I edited task 1 with the labels "sprint" and "backend"
      And I deleted the label "sprint"
      When I open the window
      And I press "cmd+k"
      And I type "views…"
      And I press "enter"
      Then the views panel warns "label sprint is gone and narrows nothing"

    @gui-view-naming-an-objective-narrows-by-its-parts
    Scenario: a view naming an objective holding sub-objectives draws the bar narrowing to each part
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I edited task 1 with the objective "2"
      And I edited task 2 with the objective "1"
      And I kept the view "Home" holding the objective "1"
      When I open the window
      And I press "cmd+1"
      Then the narrowing bar reads "all 2", "Enclosure 1", "itself 1"
      When I click the segment "Enclosure"
      Then "todo" carries "Write the spec"
      And "todo" does not carry "Review the plan"
      When I click the segment "all"
      Then "todo" carries "Review the plan"
      And the pill of the view is not marked changed

  Rule: A task that comes back is told from another at a glance, and its series has a tab of its own

    @gui-series-r-s-skips-the-chosen-occurrence
    Scenario: r then s skips the occurrence chosen on the board, and its series makes the next
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the card "Stand-up"
      Then the bar offers "Skip this one" under "r s"
      When I press "r"
      And I press "s"
      Then task 1 is archived
      And task 2 is in the column "todo"

    @gui-series-r-x-stops-and-r-r-resumes
    Scenario: r then x stops the series of the occurrence chosen, leaving it on the board, and r then r resumes it
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the card "Stand-up"
      And I press "r"
      And I press "x"
      Then series 1 is stopped
      And task 1 is in the column "todo"
      When I press "r"
      And I press "r"
      Then series 1 is active

    @gui-series-r-e-opens-its-series
    Scenario: r then e opens the series of the occurrence chosen in its tab
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the card "Stand-up"
      And I press "r"
      And I press "e"
      Then the tab is "SERIES"
      And the series shown is series 1

    @gui-series-the-form-of-an-occurrence-says-its-series
    Scenario: the form of an occurrence says its series beside its fields, and offers to skip it
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I double-click the card "Stand-up"
      Then the form says of its series "every day, on the calendar"
      And the form offers "Skip this one"
      When I click "Skip this one"
      Then task 1 is archived
      And the window is on the board

    @gui-series-the-history-of-an-occurrence-says-its-series
    Scenario: the history of an occurrence its series made says the series
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      And task 1 was skipped
      When I open the window
      And I double-click the card "Stand-up"
      And I press "cmd+t"
      Then the trace says "made from series  ↻1"

    @gui-series-the-form-of-an-occurrence-names-the-days-missed
    Scenario: the series box of an occurrence names the days missed, what it counts from, and leads to the series
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day" from "2026-11-01"
      And task 1 was skipped
      When I open the window
      And I double-click the card "Stand-up"
      Then the form says of its series "nobody was there; not made up"
      And the form says of its series "On the calendar"
      And the form says of its series "ends never"
      When I click "Edit the series…"
      Then the series shown is series 1

    @gui-series-the-tab-narrows-to-a-state
    Scenario: the series tab narrows to a state, and says so on a pill
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      And I added a task titled "Expense report" coming back "every 4 weeks on monday"
      And I stopped series 2
      When I open the window
      And I click the tab "SERIES"
      And I set the series' state to "stopped"
      Then the series listed are "Expense report"
      And the board shows the pill "state: stopped"
      When I set the series' state to "all"
      Then the series listed are "Stand-up", "Expense report"

    @gui-series-the-picker-sets-the-day-it-counts-from
    Scenario: the picker sets the day a new series counts from
      Given recurrence is on in the window
      And today is "2026-11-09"
      When I open the window
      And I press "cmd+n"
      And I click the field "rule"
      And I click the next of "from" in the rule's picker
      Then the field "from" reads "2026-11-10"

    @gui-series-the-chip-opens-its-series
    Scenario: a click on the series' number of a card opens the series in its tab
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the series of the card "Stand-up"
      Then the tab is "SERIES"
      And the series shown is series 1

    @gui-series-the-tab-lists-them
    Scenario: the series tab lists the active series apart from the stopped ones, the chosen one beside them
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      And I added a task titled "Expense report" coming back "every 4 weeks on monday"
      And I stopped series 2
      When I open the window
      And I click the tab "SERIES"
      Then the tab "SERIES" says it holds 2
      And the list of series reads "Active", "Stand-up", "Stopped", "Expense report"
      And the series shown says "Template — what the next one is made from"
      When I click the series "Expense report"
      Then the series shown is series 2
      And the series shown offers "Resume"

    @gui-series-the-tab-skips-stops-and-resumes
    Scenario: in the tab, s skips the open occurrence, x stops the series and r resumes it
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the tab "SERIES"
      And I press "s"
      Then task 1 is archived
      And task 2 is in the column "todo"
      When I click "Stop"
      Then series 1 is stopped
      And task 2 is in the column "todo"
      When I press "r"
      Then series 1 is active

    @gui-series-the-tab-reaches-an-archived-occurrence
    Scenario: an occurrence archived long ago is reached from its series
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      And task 1 was skipped
      When I open the window
      And I click the tab "SERIES"
      And I click the occurrence of task 1
      Then the form is open on task 1

    @gui-series-deleting-keeps-the-tasks-unless-told
    Scenario: deleting a series asks what becomes of its tasks, keeping them unless told
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the tab "SERIES"
      And I click "Delete…"
      Then the window asks what becomes of the tasks of series 1
      And the choice taken is "Keep its tasks, as ordinary tasks"
      When I press "enter"
      Then there is no series 1
      And task 1 is in the column "todo"

    @gui-series-deleting-every-task-asks-once-more
    Scenario: deleting a series with every task it made asks once more, naming the count
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      And task 1 was skipped
      When I open the window
      And I click the tab "SERIES"
      And I press "backspace"
      And I click the choice "Delete all 2 tasks too"
      And I press "enter"
      Then the window asks "Delete all 2 tasks too?"
      And series 1 is active
      When I press "enter"
      Then there is no series 1
      And there is no task 1
      And there is no task 2

    @gui-series-the-form-changes-the-series
    Scenario: e in the tab opens the series' form, and saving changes what comes next
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I click the tab "SERIES"
      And I press "e"
      And I set the field "title" to "Daily stand-up"
      And I press "cmd+enter"
      Then series 1 is titled "Daily stand-up"
      And task 1 is titled "Stand-up"

    @gui-series-the-repeats-picker-builds-the-rule-by-its-parts
    Scenario: the picker under the rule field builds the rule from its parts, written in the words the field reads
      Given recurrence is on in the window
      When I open the window
      And I press "cmd+n"
      And I click the field "rule"
      Then the rule's picker offers "every", "on", "at", "ends" and "the next one comes"
      When I click "weeks" in the rule's picker
      And I click "Th" in the rule's picker
      And I click the next of "every" in the rule's picker
      Then the field "rule" reads "every 2 weeks on thu"
      When I press "enter"
      Then the rule's picker is closed
      And the field "rule" reads "every 2 weeks on thu"

    @gui-series-the-repeats-picker-ends-and-counts-after-done
    Scenario: the picker ends a rule after so many times, and counts it from the last one done
      Given recurrence is on in the window
      When I open the window
      And I press "cmd+n"
      And I click the field "rule"
      And I click "after the last is done" in the rule's picker
      And I click "after" in the rule's picker
      Then the field "rule" reads "every day after done 10 times"

    @gui-series-the-repeats-picker-walked-by-the-keyboard
    Scenario: the picker is walked by the keyboard, a part a line
      Given recurrence is on in the window
      When I open the window
      And I press "cmd+n"
      And I click the field "rule"
      And I press "right"
      And I press "down"
      And I press "right"
      Then the field "rule" reads "every 2 weeks"

    @gui-series-the-repeats-picker-cancelled-leaves-the-words
    Scenario: the picker cancelled leaves the field as it was before it opened
      Given recurrence is on in the window
      When I open the window
      And I press "cmd+n"
      And I set the field "rule" to "every monday"
      And I click the field "rule"
      And I click "months" in the rule's picker
      And I press "esc"
      Then the rule's picker is closed
      And the field "rule" reads "every monday"

    @gui-series-the-repeats-field-reads-the-rule-back
    Scenario: a rule typed is read back under its field with the days it comes on, before anything is saved
      Given recurrence is on in the window
      When I open the window
      And I press "cmd+n"
      And I set the field "rule" to "every other thursday 14:00"
      Then under the field "rule" the window reads "every 2 weeks on Thursday at 14:00, on the calendar"
      And under the field "rule" the window reads the next three days
      When I set the field "rule" to "every blue moon"
      Then under the field "rule" the window reads "not a rule cairn reads"

    @gui-series-saving-an-occurrence-asks-this-one-or-all
    Scenario: changing the title of an occurrence asks whether the series changes too
      Given recurrence is on in the window
      And I added a task titled "Stand-up" coming back "every day"
      When I open the window
      And I double-click the card "Stand-up"
      And I set the field "title" to "Daily stand-up"
      And I press "cmd+enter"
      Then the window asks "Change Stand-up?"
      When I click "The whole series"
      Then task 1 is titled "Daily stand-up"
      And series 1 is titled "Daily stand-up"
