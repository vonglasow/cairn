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
