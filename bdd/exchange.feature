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
        cairn task new "Write the spec" --parent 1 --state done
        cairn task new "Review" --state todo
        cairn task edit 1 --blocker 3
        cairn task new "Old idea"
        cairn task archive 4
        """
      When I import the tasks of that file
      Then it succeeds
      And it says 4 tasks and 1 objective were added
      And `cairn task show 2` says "objectives   #2"
      And `cairn task show 2` says "priority     high"
      And `cairn task show 2` says "release"
      And `cairn task show 2` says "blocked by   #4"
      And `cairn task show 3` says "parent       2"
      And `cairn task show 3` says "state        done"
      And task 5 is archived
      And task 1 is not archived

    @exchange-import-not-an-export
    Scenario: a file that holds no tasks is refused, and nothing is added
      Given I added a task titled "Already here"
      And a file of my own that holds no tasks
      When I import the tasks of that file
      Then it fails
      And `cairn task list` lists 1 task

    @exchange-import-broken-link
    Scenario: a file whose tasks point outside it is refused before anything is added
      Given I added a task titled "Already here"
      And a file of my own holding a task that hangs under a task it does not hold
      When I import the tasks of that file
      Then it fails
      And it says task 9 is not in the file
      And `cairn task list` lists 1 task
