Feature: The board

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a window of 100 by 44

  Rule: The work is laid out in four columns and a band

    @tui-a-task-is-taken-off-the-board-from-the-cursor
    Scenario: a task is archived, and unarchived
      Given I added a task titled "Migration Jenkins"
      And I added a task titled "Spike Argo"
      When I open the board
      Then the cursor is on the task "Migration Jenkins"
      When I press "x"
      Then the board carries 1 task
      And the board says nothing of "Migration Jenkins"
      When I press "shift+a"
      Then the board carries 2 tasks
      When I press "x"
      Then the board carries 2 tasks
      And "BACKLOG" carries "Migration Jenkins"

  Rule: A window too narrow is told so, rather than drawn badly

    @tui-too-narrow-says-what-it-needs
    Scenario: a window under sixty says how many it needs
      Given a window of 47 by 30
      When I open the board
      Then it says it needs 60 columns
      And it says the window is 47

  Rule: hjkl walks the cards, and ctrl+h and ctrl+l go to another column at once

    @tui-the-board-opens-on-the-first-card
    Scenario: the board opens on the first card of the backlog, in no mode
      Given I added a task titled "Write the spec"
      When I open the board
      Then the cursor is on the task "Write the spec"
      And the board does not say "NORMAL"

  Rule: Walking sideways keeps the level, and the ends are one key away

    @tui-sideways-keeps-the-level
    Scenario: changing column stays at the same level in it
      Given I added a task titled "First" with the state "todo"
      And I added a task titled "Second" with the state "todo"
      And I added a task titled "Also first" with the state "current"
      And I added a task titled "Also second" with the state "current"
      When I open the board
      And I press "ctrl+l", "j"
      Then the cursor is on the task "Second"
      When I press "l"
      Then the cursor is on the task "Also second"

  Rule: The band is three lanes, one under each of the last three columns

    @tui-ctrl-k-leaves-the-band
    Scenario: ctrl+j goes into the band at once, and ctrl+k out of it
      Given I added a task titled "Above" with the state "todo"
      And I added a task titled "Further" with the state "todo"
      And I added a task titled "Below" with the state "blocked"
      When I open the board
      And I press "l", "ctrl+j"
      Then the cursor is on the task "Below"
      When I press "ctrl+k"
      Then the cursor is on the task "Further"

  Rule: A page turned off in the configuration is not drawn

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

    @tui-done-asked-about-can-be-kept
    Scenario: saying no leaves the task and its steps as they were
      Given I added a task titled "Ship it" with the state "current"
      And I added the step "Write the spec" to task 1
      When I open the board
      And I press "ctrl+l", "ctrl+l", "shift+l"
      And the board looks again
      And I press "esc"
      Then task 1 is in the column "current"

    @tui-out-of-the-band-returns-where-the-cursor-came-from
    Scenario: coming back out returns the task to the column it was pushed from
      Given I added a task titled "Write the spec" with the state "current"
      When I open the board
      And I press "ctrl+l", "ctrl+l", "shift+j"
      And I choose "blocked"
      And I press "shift+k"
      Then task 1 is in the column "current"

    @tui-moving-a-task-down-among-its-kind-in-the-band
    Scenario: pushing a task down in the band puts it one place lower among the tasks of its kind
      Given I added a task titled "First" with the state "blocked"
      And I added a task titled "Second" with the state "blocked"
      And I added a task titled "Elsewhere" with the state "watching"
      When I open the board
      And I press "ctrl+l", "ctrl+j", "shift+j"
      Then task 1 is in the column "blocked"
      And the band reads "Second", "First", "Elsewhere"

    @tui-moving-a-task-up-among-its-kind-in-the-band
    Scenario: pushing a task up in the band puts it one place higher among the tasks of its kind
      Given I added a task titled "Elsewhere" with the state "blocked"
      And I added a task titled "First" with the state "paused"
      And I added a task titled "Second" with the state "paused"
      When I open the board
      And I press "ctrl+l", "ctrl+j", "l", "j", "shift+k"
      Then task 3 is in the column "paused"
      And the band reads "Elsewhere", "Second", "First"

  Rule: The bar shows only what works here

    @tui-the-bar-follows-the-mode
    Scenario: the bar changes with what the cursor holds
      Given I added a task titled "Write the spec"
      When I open the board
      Then the bar offers "move"
      When I press "ctrl+l"
      Then the bar does not offer "move"

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

  Rule: A card says what its task carries, and no more

    @tui-a-card-carries-the-number-and-the-title
    Scenario: a card is its number and its title
      Given I added a task titled "Write the spec"
      When I open the board
      Then the card for task 1 says "1"
      And the card for task 1 says "Write the spec"

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

  Rule: Two tabs, and an objective opens a board of its own

    @tui-the-tab-key-changes-tab
    Scenario: tab moves to the objectives
      When I open the board
      And I press "tab"
      Then the tab is "OBJECTIVES"

  Rule: Objectives are a tree, and an objective's board walks its sub-objectives

    @tui-objective-shift-n-adds-a-sub-objective
    Scenario: N adds a sub-objective under the objective the cursor is on
      Given I added an objective titled "Home automation"
      When I open the board
      And I press "tab", "shift+n"
      And I set the field "title" to "Firmware"
      And I save the form
      Then objective 2 sits under objective 1

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

  Rule: An objective's form says where it sits, and a task's what it counts in

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

    @tui-objective-the-task-form-refuses-a-double-count
    Scenario: a task put on a parent and on one of its sub-objectives is refused at the foot of the form
      Given I added an objective titled "Home automation"
      And I added an objective titled "Enclosure" under objective 1
      And I added a task titled "Pin map" serving objective 2
      When I open the board
      And I press "e"
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

  Rule: Work started early is warned about

    @tui-objective-work-started-early-is-warned-about
    Scenario: putting a task in focus says when its objective waits on one not reached
      Given I added an objective titled "Firmware"
      And I added an objective titled "Features"
      And objective 2 waits on objective 1
      And I added a task titled "OTA updates" serving objective 2
      When I open the board
      And I press "f"
      Then the board says "#2 Features waits on #1 Firmware, not reached"

  Rule: Nothing is written until it is saved

    @tui-a-form-cancelled-changes-nothing
    Scenario: leaving the form throws the change away
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "e"
      And I set the field "title" to "Something else"
      And I press "esc"
      Then task 1 is titled "Write the spec"

  Rule: Two tasks are linked from the form, and a card opens a task linked to it

    @tui-the-form-links-a-task
    Scenario: a link typed in the form links the two tasks
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I open the board
      And I press "j", "e"
      And I set the field "links" to "1"
      And I save the form
      Then task 2 is linked to task 1
      And task 1 is linked to task 2

  Rule: A form is filled in the same way as the window's

    @tui-enter-writes-a-line-in-the-description
    Scenario: enter writes a line in the description rather than saving
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "e", "tab"
      And I type "Deux colonnes"
      And I press "enter"
      And I type "et une bande"
      Then the mode is "EDIT"
      And task 1 is titled "Write the spec"

  Rule: Losing work is asked about first

    @tui-m-moves-by-a-columns-letter
    Scenario: m then a column's first letter moves the chosen task there
      Given I added a task titled "Write the spec" with the state "todo"
      When I open the board
      And I press "ctrl+l", "m", "d"
      And the board looks again
      Then "DONE" carries "Write the spec"

    @tui-a-whole-column-goes-off-the-board-at-once
    Scenario: the column the cursor is on is archived in one gesture
      Given I added a task titled "Write the spec" with the state "done"
      And I added a task titled "Ship it" with the state "done"
      When I open the board
      And I press "ctrl+l", "ctrl+l", "ctrl+l"
      And I press "shift+x"
      Then the board says "archive every task in DONE?"
      When I press "enter"
      And the board looks again
      Then the board carries 0 tasks

  Rule: The detail shows what a card cannot

    @tui-the-detail-edits-what-it-is-showing
    Scenario: a task being read is edited without closing it first
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "enter"
      And I press "e"
      Then the mode is "EDIT"
      And the board says "Write the spec"

    @tui-the-detail-focuses-what-it-is-showing
    Scenario: the key the detail offers for focus focuses what it is showing
      Given I added a task titled "Write the spec"
      When I open the board
      And I press "enter"
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

  Rule: An open task shows what is done to it elsewhere

    @tui-an-open-task-shows-a-step-added-elsewhere
    Scenario: a step added elsewhere appears on the task open on the board
      Given I added a task titled "Ship it"
      When I open the board
      And I press "enter"
      And somebody else adds the step "Draft" to task 1
      And the board looks again
      Then the board says "[ ] 1  Draft"

  Rule: The board is filtered by column, priority, label, objective and due day, and every filter is cleared at once

    Background:
      Given I added a task titled "Write the spec" in "todo" at the priority "high" carrying the label "backend"
      And I added a task titled "Review the plan" in "todo" at the priority "normal" carrying the label "backend"
      And I added a task titled "Ship it" in "current" at the priority "high" carrying the label "sprint"

    @tui-every-filter-is-cleared-at-once
    Scenario: one key takes every filter away
      When I open the board
      And I press "shift+f", "l", "j", "j", "j", "l", "j", "backspace"
      Then the board says "Review the plan"
      And the board says "Ship it"

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

  Rule: The board names the workspace it shows, and opens another in its place

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

    @tui-workspace-a-recent-folder-opens-in-place
    Scenario: a recent folder chosen opens in place of the board
      Given the workspace "work", the default
      And the folder "scratch" holds a task titled "Plant the bulbs"
      And the recent folder "scratch", opened yesterday
      When I open the board
      And I press "ctrl+o", "down", "enter"
      Then the header opens with "◌ " and the folder "scratch"
      And "BACKLOG" carries "Plant the bulbs"

  Rule: The board wears the colour scheme the window wears

    @tui-colours-follow-the-scheme
    Scenario: on the fixtures, every screen is drawn in the scheme's colours alone
      Given the fixtures
      When I open the board
      Then the board says "Done this week"
      And no screen wears a colour the scheme does not have

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

    @tui-view-0-goes-back-to-the-whole-board
    Scenario: 0 after v leaves the view for the whole board, and esc leaves what is in force
      Given I kept the view "Sprint" holding the label "sprint"
      When I open the board
      And I press "v", "1", "v", "esc"
      Then the board carries 1 task
      When I press "v", "0"
      Then the board carries 3 tasks
      And the board does not say "view Sprint"

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

  @tui-series-the-trace-says-made-from-series
  Scenario: an occurrence its series made says so in its trace
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    And task 1 was skipped
    And a window of 140 by 44
    When I open the board
    And I press "ctrl+l", "enter"
    Then the board says "made from series  ↻1"

  @tui-series-r-s-skips
  Scenario: r s skips the occurrence under the cursor, and its series makes the next
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "ctrl+l", "r", "s"
    Then task 1 is archived
    And task 2 is titled "Stand-up"
    And task 2 is in the column "todo"

  @tui-series-r-x-stops-and-r-r-resumes
  Scenario: r x stops the series, leaving the task on the board, and r r resumes it
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "ctrl+l", "r", "x"
    Then series 1 is stopped
    And task 1 is in the column "todo"
    When I press "r", "r"
    Then series 1 is active

  @tui-series-r-e-opens-its-series
  Scenario: r e opens the series of the occurrence in its tab
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "ctrl+l", "r", "e"
    Then the tab is "SERIES"
    And the board says "TEMPLATE"

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

  @tui-series-saving-an-occurrence-asks-this-one-or-all
  Scenario: changing the title of an occurrence asks whether the series changes too
    Given recurrence is on on the board
    And I added a task titled "Stand-up" coming back "every day"
    When I open the board
    And I press "ctrl+l", "e"
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
    And I press "ctrl+l", "e"
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

  Rule: The configuration is read again on the board's beat, and a key is set from the : line

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
      When I press "ctrl+l"
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
