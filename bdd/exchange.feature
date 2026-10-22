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
