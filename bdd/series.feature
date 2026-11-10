Feature: A series, the tasks that come back

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet
    And recurrence is on for this store
    And today is "2026-11-09"

  Rule: A rule is written in plain words, and read back as a sentence with the next dates

    @series-new-makes-the-first-occurrence
    Scenario: a task made with a rule is the first occurrence of a series
      When I run `cairn task new "Team stand-up" --every "every weekday 10:30" --label work`
      Then it succeeds
      And the answer says "↻1 every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending"
      And the answer says "day      Mon 9 Nov"
      And the answer says "then     Tue 10 Nov · Wed 11 Nov · Thu 12 Nov"
      And `cairn task show 1` says "deadline     2026-11-09"
      And `cairn series list` says "↻1  Team stand-up"

    @series-grammar-reads-other-and-a-day
    Scenario: every other thursday is every two weeks on a Thursday
      When I run `cairn task new Retrospective --every "every other thursday 14:00"`
      Then it succeeds
      And the answer says "every 2 weeks on Thursday at 14:00, on the calendar"
      And the answer says "day      Thu 12 Nov"
      And the answer says "then     Thu 26 Nov · Thu 10 Dec · Thu 24 Dec"

    @series-grammar-after-done
    Scenario: a rule can count from the day the last one is done
      When I run `cairn task new "Water the plants" --every "every 3 days after done"`
      Then it succeeds
      And the answer says "every 3 days, after the last is done"
      And the answer says "then     3 days after it is done"

    @series-grammar-from-a-day
    Scenario: a series can start on a later day
      When I run `cairn task new Review --every "every friday" --from 2026-11-20`
      Then it succeeds
      And the answer says "day      Fri 20 Nov"

    @series-grammar-ends
    Scenario: a rule may end after so many times
      When I run `cairn task new Course --every "every monday 3 times"`
      Then it succeeds
      And the answer says "every week on Monday, on the calendar, from Mon 9 Nov 2026, 3 times"

    @series-grammar-refuses-two-ends
    Scenario: an end is a number of times or a day, never both
      When I run `cairn task new Course --every "every day 3 times until 2026-12-01"`
      Then it fails
      And it says on standard error "an end is a number of times or a day, never both"
      And `cairn task list` says "there is no task"

    @series-grammar-refuses-a-word-it-does-not-know
    Scenario: a word the grammar does not know is named
      When I run `cairn task new Moon --every "every blue moon"`
      Then it fails
      And it says on standard error "blue"

  Rule: One occurrence is open at a time, and the next is made when it is done or skipped

    @series-done-makes-the-next
    Scenario: putting an occurrence in done makes the next one
      Given I ran `cairn task new "Team stand-up" --every "every weekday 10:30"`
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And the answer says "↻1 made #2 for Tue 10 Nov"
      And `cairn task show 2` says "deadline     2026-11-10"
      And `cairn task list --series 1` lists 2 tasks

    @series-one-open-at-a-time
    Scenario: nothing more is made while one is open
      Given I ran `cairn task new "Team stand-up" --every "every weekday 10:30"`
      And it is 3 days from now
      When I run `cairn task list --series 1`
      Then it succeeds
      And `cairn task list --series 1` lists 1 task

    @series-missed-are-not-made-up
    Scenario: the days passed while one was open are missed, never made
      Given I ran `cairn task new "Team stand-up" --every "every weekday 10:30"`
      And today is "2026-11-12"
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And the answer says "↻1 made #2 for Thu 12 Nov"
      And the answer says "↻1 missed Tue 10 Nov, Wed 11 Nov"
      And `cairn task list --series 1` lists 2 tasks
      And `cairn series show 1` says "2 missed"

    @series-skip-makes-the-next
    Scenario: skipping an occurrence takes it off the board and makes the next
      Given I ran `cairn task new "Team stand-up" --every "every weekday 10:30"`
      When I run `cairn task skip 1`
      Then it succeeds
      And the answer says "#1 skipped"
      And the answer says "↻1 made #2 for Tue 10 Nov"
      And `cairn task list` lists 1 task
      And `cairn series show 1` says "1 skipped"

    @series-skip-refuses-an-ordinary-task
    Scenario: skipping a task that is no occurrence is refused
      Given I added a task titled "Write the spec"
      When I run `cairn task skip 1`
      Then it fails
      And it says on standard error "task 1 is no occurrence of a series"

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

    @series-an-occurrence-serves-the-objectives
    Scenario: an occurrence serves what the series serves, and counts there
      Given I added an objective titled "Team"
      And I ran `cairn task new "Team stand-up" --every "every weekday" --objective 1`
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And `cairn task show 2` says "objectives   #1"
      And `cairn objective show 1` says "1 of 2 tasks"

  Rule: A series is stopped, resumed, changed and deleted on its own

    @series-stop-keeps-the-open-task
    Scenario: stopping a series leaves its open occurrence as an ordinary open task
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn series stop 1`
      Then it succeeds
      And the answer says "↻1 stopped; #1 stays open, and nothing more is made"
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And `cairn task list --series 1 --with-archived` lists 1 task
      And `cairn series list --state stopped` says "↻1"

    @series-resume-catches-nothing-up
    Scenario: resuming makes the next from today, catching nothing up
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn series stop 1`
      And I ran `cairn task edit 1 --state done`
      And today is "2026-11-16"
      When I run `cairn series resume 1`
      Then it succeeds
      And the answer says "↻1 resumed · next Mon 16 Nov, made as #2"
      And `cairn series show 1` says "0 missed"

    @series-list-narrows
    Scenario: the series are listed, narrowed by their state and a search
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task new "Expense report" --every "every 4 weeks on monday"`
      And I ran `cairn series stop 2`
      When I run `cairn series list --state stopped --search expense`
      Then it succeeds
      And the answer says "↻2  Expense report"
      And the answer does not say "Team stand-up"

    @series-show-says-everything
    Scenario: showing a series says its rule, its template, what it did and its history
      Given I ran `cairn task new "Team stand-up" --every "every weekday 10:30" --label work --estimate 15m`
      And I ran `cairn task skip 1`
      When I run `cairn series show 1`
      Then it succeeds
      And the answer says "rule      every weekday at 10:30, on the calendar, from Mon 9 Nov 2026, never ending"
      And the answer says "template  Team stand-up · normal · work · 15m"
      And the answer says "open      #2"
      And the answer says "skipped #1"
      And the answer says "made #2 for Tue 10 Nov"

    @series-edit-changes-what-comes-next
    Scenario: changing a series changes the ones to come, never the ones made
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn series edit 1 --title Stand-up --every "every tuesday"`
      Then it succeeds
      When I run `cairn task edit 1 --state done`
      Then the answer says "↻1 made #2 for Tue 10 Nov"
      And `cairn task show 1` says "Team stand-up"
      And `cairn task show 2` says "Stand-up"
      And `cairn series show 1` says "changed title"

    @series-task-edit-reaches-the-series
    Scenario: an occurrence changed with --series changes the series too
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn task edit 1 --title Daily --priority high --series`
      Then it succeeds
      When I run `cairn task edit 1 --state done`
      Then `cairn task show 2` says "Daily"
      And `cairn task show 2` says "priority     high"

    @series-task-edit-series-refuses-an-ordinary-task
    Scenario: --series on a task that is no occurrence is refused
      Given I added a task titled "Write the spec"
      When I run `cairn task edit 1 --title Spec --series`
      Then it fails
      And it says on standard error "task 1 is no occurrence of a series"

    @series-every-makes-a-task-the-first
    Scenario: a rule given to a task makes it the first occurrence
      Given I added a task titled "Water the plants"
      When I run `cairn task edit 1 --every "every 3 days after done"`
      Then it succeeds
      And the answer says "↻1 every 3 days, after the last is done"
      And `cairn task list --repeats` lists 1 task

    @series-every-refuses-an-occurrence
    Scenario: a rule given to an occurrence is refused, the series' own is changed instead
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      When I run `cairn task edit 1 --every "every day"`
      Then it fails
      And it says on standard error "cairn series edit 1"

    @series-delete-asks-what-becomes-of-the-tasks
    Scenario: deleting a series is refused until what becomes of its tasks is said
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task edit 1 --state done`
      When I run `cairn series delete 1 --yes`
      Then it fails
      And it says on standard error "↻1 made 2 tasks"
      And it says on standard error "--keep-tasks"
      And it says on standard error "--delete-open"
      And it says on standard error "--delete-tasks"

    @series-delete-keeps-the-tasks
    Scenario: deleting a series and keeping its tasks leaves them ordinary tasks
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task edit 1 --state done`
      When I run `cairn series delete 1 --keep-tasks --yes`
      Then it succeeds
      And `cairn task list` lists 2 tasks
      And `cairn task list --repeats` lists 0 tasks
      And `cairn series list` says "there is no series"

    @series-delete-the-open-one
    Scenario: deleting a series can take its open occurrence with it, and keep the done
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task edit 1 --state done`
      When I run `cairn series delete 1 --delete-open --yes`
      Then it succeeds
      And `cairn task list` lists 1 task
      And `cairn task show 1` says "done"

    @series-delete-every-task
    Scenario: deleting a series can take every task it made with it
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task edit 1 --state done`
      When I run `cairn series delete 1 --delete-tasks --yes`
      Then it succeeds
      And `cairn task list --with-archived` says "there is no task"

  Rule: A listing keeps the occurrences when asked

    @series-list-repeats
    Scenario: a listing keeps only the occurrences, or those of one series
      Given I added a task titled "Write the spec"
      And I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn task new Review --every "every friday"`
      When I run `cairn task list --repeats`
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

    @series-off-leaves-the-tasks-ordinary
    Scenario: with recurrence off, an occurrence is an ordinary task, and nothing is made
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And recurrence is off for this store
      When I run `cairn task edit 1 --state done`
      Then it succeeds
      And the answer does not say "↻"
      And `cairn task list --with-archived` lists 1 task

    @series-on-again-brings-it-back
    Scenario: turned on again, every series is as it was, and nothing is caught up
      Given I ran `cairn task new "Team stand-up" --every "every weekday"`
      And I ran `cairn series stop 1`
      And recurrence is off for this store
      And recurrence is on for this store
      When I run `cairn series show 1`
      Then it succeeds
      And the answer says "stopped"
      And `cairn task list` says "↻1"
