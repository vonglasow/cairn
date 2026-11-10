Feature: Tasks carried from one store to another

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: The tasks are written to a file, with the objectives they serve

    @exchange-export
    Scenario: the tasks are written to a file of my choosing
      Given I added an objective titled "Release 1.0"
      And I added a task titled "Write the spec" with the objective "1"
      And I added a task titled "Ship it"
      When I export the tasks to a file of my own
      Then it succeeds
      And it says the file written holds 2 tasks and 1 objective

    @exchange-an-objective-nothing-serves
    Scenario: an objective no task serves is written too, and comes back
      Given a file of my own exported from a store where these were run:
        """
        cairn objective new "Release 1.0"
        cairn objective new "Someday"
        cairn task new "Ship it" --objective 1
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 1 task and 2 objectives were added
      And `cairn objective show 2` says "Someday"

    @exchange-export-never-overwrites
    Scenario: exporting onto a file that is already there is refused
      Given I added a task titled "Write the spec"
      And a file of my own that holds no tasks
      When I export the tasks to that file
      Then it fails
      And that file is as it was

  Rule: The tasks of a file are added, as new tasks

    @exchange-import-adds
    Scenario: the tasks read in come after the ones already here
      Given I added a task titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn task new "Write the spec"
        cairn task new "Ship it"
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 2 tasks and 0 objectives were added
      And `cairn task list` lists 3 tasks
      And `cairn task list` lists task 1 titled "Already here"
      And `cairn task list` lists task 2 titled "Write the spec"
      And `cairn task list` lists task 3 titled "Ship it"

    @exchange-import-twice
    Scenario: a file read in twice adds its tasks twice
      Given a file of my own exported from a store where these were run:
        """
        cairn task new "Write the spec"
        """
      When I import the tasks of that file
      And I import the tasks of that file
      Then it succeeds
      And `cairn task list` lists 2 tasks
      And `cairn task list` lists task 2 titled "Write the spec"

    @exchange-import-carries
    Scenario: what a task carries comes with it, pointing at the new numbers
      Given I added a task titled "Already here"
      And I added an objective titled "Already served"
      And a file of my own exported from a store where these were run:
        """
        cairn objective new "Release 1.0"
        cairn task new "Ship it" --objective 1 --priority high --label release
        cairn step new 1 "Write the spec" --done
        cairn task new "Review" --state todo
        cairn task edit 1 --blocker 2
        cairn task new "Old idea"
        cairn task archive 3
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 3 tasks and 1 objective were added
      And `cairn task show 2` says "objectives   #2"
      And `cairn task show 2` says "priority     high"
      And `cairn task show 2` says "release"
      And `cairn task show 2` says "blocked by   #3"
      And `cairn step list 2` names step 1 "Write the spec" as done
      And `cairn task show 3` says "state        todo"
      And task 4 is archived
      And task 1 is not archived

    @exchange-import-reads-sub-tasks-as-steps
    Scenario: a file written before steps brings its sub-tasks in as steps, what is left first
      Given a file of my own written before steps, holding "Ship it" and its sub-tasks "Write the spec" done and "Write the tests"
      When I import the tasks of that file
      Then it succeeds
      And it says 1 task and 0 objectives were added
      And `cairn step list 1` names step 1 "Write the tests" as not done
      And `cairn step list 1` names step 2 "Write the spec" as done

    @exchange-import-not-an-export
    Scenario: a file that holds no tasks is refused, and nothing is added
      Given I added a task titled "Already here"
      And a file of my own that holds no tasks
      When I import the tasks of that file
      Then it fails
      And `cairn task list` lists 1 task

    @exchange-import-objectives-without-tasks
    Scenario: a file that holds objectives and no task adds its objectives
      Given a file of my own exported from a store where these were run:
        """
        cairn objective new "Someday"
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 0 tasks and 1 objective were added
      And `cairn objective show 1` says "Someday"

    @exchange-import-broken-link
    Scenario: a file whose tasks point outside it is refused before anything is added
      Given I added a task titled "Already here"
      And a file of my own holding a task blocked by a task it does not hold
      When I import the tasks of that file
      Then it fails
      And it says task 9 is not in the file
      And `cairn task list` lists 1 task

    @exchange-a-link-comes-with-its-tasks
    Scenario: linked tasks come back linked, at their new numbers
      Given I added a task titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn task new "Ship it"
        cairn task new "Write the tests"
        cairn task new "It does not hold" --link 1 --link 2
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn task show 4` says "linked       #2, #3"
      And `cairn task show 2` says "linked       #4"
      And `cairn task show 3` says "linked       #4"

    @exchange-a-link-outside-the-file
    Scenario: a file whose task is linked to a task it does not hold is refused before anything is added
      Given I added a task titled "Already here"
      And a file of my own holding a task linked to a task it does not hold
      When I import the tasks of that file
      Then it fails
      And it says task 9 is not in the file
      And `cairn task list` lists 1 task

  Rule: A file says what a board needs to be shown, every part of it optional

    @exchange-time-comes-day-by-day
    Scenario: the time spent comes with the days it was spent on
      Given a file of my own exported from a store where these were run:
        """
        cairn task new "Ship it"
        cairn timesheet add 1 2h --on 2026-10-01
        cairn timesheet add 1 45m --on 2026-10-02
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn task show 1` says "2h45m"
      And `cairn timesheet show --since 2026-10-01 --until 2026-10-01` says "2h"

    @exchange-the-focus-comes-with-it
    Scenario: a task in focus comes back in focus
      Given I added a task titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn task new "Ship it"
        cairn task focus 1
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn task focus` says "Ship it"

    @exchange-labels-come-with-their-colours
    Scenario: a label comes with its colour, even one nothing carries
      Given a file of my own exported from a store where these were run:
        """
        cairn label new release
        cairn label colour release teal
        """
      When I import the tasks of that file
      Then it succeeds
      And I run `cairn label list --json`
      And the JSON gives the label "release" the colour "teal"

    @exchange-a-day-counted-from-today
    Scenario: a day written from today is read as that day
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 4, "objectives": [], "tasks": [
          {"number": 1, "title": "Ship it", "state": "todo", "priority": "normal", "deadline": "+3d",
           "spent": {"-1d": "2h"}}]}
        """
      When I import the tasks of that file
      Then it succeeds
      And task 1 is due 3 days from today
      And the time sheet of yesterday gives task 1 "2h"

    @exchange-a-day-counted-wrongly
    Scenario: a day that cannot be read is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 4, "objectives": [], "tasks": [
          {"number": 1, "title": "Ship it", "state": "todo", "priority": "normal", "deadline": "+3x"}]}
        """
      When I import the tasks of that file
      Then it fails
      And `cairn task list` lists 0 tasks

    @exchange-a-total-of-time-from-version-3
    Scenario: a file of version 3 brings the total time spent, on the day it is read
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 3, "objectives": [], "tasks": [
          {"number": 1, "title": "Ship it", "state": "todo", "priority": "normal", "time_spent": "1h30m"}]}
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn task show 1` says "1h30m"

  Rule: A file carries which objective sits under which, what each waits on, and what each label is for

    @exchange-objectives-come-with-their-tree-and-what-they-wait-on
    Scenario: a sub-objective comes back under its parent, still waiting on what it waited on
      Given I added an objective titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn objective new "Home automation"
        cairn objective new "Firmware" --parent 1
        cairn objective new "Features" --parent 1 --waits-on 2
        cairn task new "Dimmer scenes" --objective 3
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn objective show 4` says "#2 Home automation"
      And `cairn objective show 4` says "#3 Firmware"
      And `cairn objective show 3` says "#4 Features waits on it"
      And `cairn task show 1` says "#4"

    @exchange-a-label-comes-with-its-scope
    Scenario: a label comes with what it may be put on
      Given a file of my own exported from a store where these were run:
        """
        cairn label new area --scope objectives
        """
      When I import the tasks of that file
      Then it succeeds
      And I run `cairn label list --json`
      And the JSON gives the label "area" the scope "objectives"

    @exchange-a-parent-outside-the-file
    Scenario: a file whose objective sits under one it does not hold is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 5, "objectives": [
          {"number": 1, "title": "Firmware", "state": "active", "priority": "normal", "parent": 9}],
         "tasks": []}
        """
      When I import the tasks of that file
      Then it fails
      And it says objective 9 is not in the file
      And `cairn objective list` lists 0 objectives

    @exchange-waiting-outside-the-file
    Scenario: a file whose objective waits on one it does not hold is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 5, "objectives": [
          {"number": 1, "title": "Features", "state": "active", "priority": "normal", "waits_on": [9]}],
         "tasks": []}
        """
      When I import the tasks of that file
      Then it fails
      And it says objective 9 is not in the file
      And `cairn objective list` lists 0 objectives

  Rule: A file carries the series, and what an import makes of them is the same series, linked to their tasks

    @exchange-a-series-comes-with-its-occurrences
    Scenario: a series comes back with its rule, its template, its state and its trace, at the new numbers
      Given recurrence is on for this store
      And today is "2026-11-09"
      And I added a task titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn task new "Team stand-up" --every "every weekday 10:30" --label work --estimate 15m
        cairn task edit 1 --state done
        cairn task skip 2
        cairn task new "Expense report" --every "every 4 weeks on monday"
        cairn series stop 2
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 4 tasks and 0 objectives were added
      And `cairn series show 1` says "every weekday at 10:30, on the calendar, from Mon 9 Nov 2026"
      And `cairn series show 1` says "Team stand-up · normal · work · 15m"
      And `cairn series show 1` says "1 done · 1 skipped · 0 missed"
      And `cairn series show 1` says "open      #4"
      And `cairn series show 1` says "skipped #3, made for Tue 10 Nov"
      And `cairn series show 2` says "stopped since"
      And `cairn task list --series 1 --with-archived` lists 3 tasks
      And `cairn task show 2` says "↻1"

    @exchange-an-imported-series-goes-on
    Scenario: the open occurrence of a series read in makes the next when it is done
      Given recurrence is on for this store
      And today is "2026-11-09"
      And a file of my own exported from a store where these were run:
        """
        cairn task new "Team stand-up" --every "every weekday 10:30"
        """
      When I import the tasks of that file
      And I run `cairn task edit 1 --state done`
      Then it succeeds
      And the answer says "↻1 made #2 for Tue 10 Nov"

    @exchange-a-series-outside-the-file
    Scenario: a file whose series serves an objective it does not hold is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 6, "objectives": [], "tasks": [],
         "series": [{"number": 1, "rule": "every day", "from": "2026-11-09",
           "template": {"title": "Water", "priority": "normal", "objectives": [9]}}]}
        """
      When I import the tasks of that file
      Then it fails
      And it says objective 9 is not in the file
      And `cairn task list` lists 0 tasks

    @exchange-a-series-that-cannot-be-read
    Scenario: a file whose series' rule cannot be read is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 6, "objectives": [],
         "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}],
         "series": [{"number": 1, "rule": "every blue moon", "from": "2026-11-09",
           "template": {"title": "Water", "priority": "normal"}}]}
        """
      When I import the tasks of that file
      Then it fails
      And `cairn task list` lists 0 tasks

  Rule: A file carries the views, and an import gives them back naming the objectives at their new numbers

    @exchange-views-come-with-their-filters
    Scenario: a view comes back with its filters, naming its objective at its new number
      Given I added an objective titled "Already here"
      And a file of my own exported from a store where these were run:
        """
        cairn objective new "Home automation"
        cairn view new Home --objective 1 --state todo
        cairn view new Shopping --label shopping --search milk
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn view show Home` says "objective 2 · state todo"
      And `cairn view show Shopping` says "label shopping · search"
      And `cairn objective show 2` says "Home automation"

    @exchange-a-view-named-here-stays-as-it-is
    Scenario: a view whose name a view here already has is left out, and the one here stays as it was
      Given I run `cairn view new home --state current`
      And a file of my own exported from a store where these were run:
        """
        cairn view new Home --state todo
        cairn view new Errands --state blocked
        """
      When I import the tasks of that file
      Then it succeeds
      And `cairn view show Home` says "state current"
      And `cairn view show Errands` says "state blocked"

    @exchange-a-view-outside-the-file
    Scenario: a file whose view names an objective it does not hold is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 7, "objectives": [],
         "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}],
         "views": [{"name": "Home", "objective": [9]}]}
        """
      When I import the tasks of that file
      Then it fails
      And it says objective 9 is not in the file
      And `cairn task list` lists 0 tasks
      And `cairn view list` says "there is no view"

    @exchange-a-view-that-cannot-be-read
    Scenario: a file whose view holds no filter is refused before anything is added
      Given a file of my own holding:
        """
        {"cairn": "tasks", "version": 7, "objectives": [],
         "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}],
         "views": [{"name": "Everything"}]}
        """
      When I import the tasks of that file
      Then it fails
      And `cairn task list` lists 0 tasks
