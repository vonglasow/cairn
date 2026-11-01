Feature: Every label at once

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: Every label is listed with what carries it

    @label-list-says-what-carries-each
    Scenario: the labels are listed, each with how many tasks and objectives carry it
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Ship it" with the labels "sp26-20" and "release"
      And I ran `cairn objective new "Ship cairn" --label release`
      When I run `cairn label list`
      Then it succeeds
      And the label "sp26-20" is carried by 2 tasks and 0 objectives
      And the label "release" is carried by 1 task and 1 objective
      And the label "doc" is carried by 1 task and 0 objectives

    @label-list-travels-in-the-json
    Scenario: the labels read as JSON
      Given I added a task titled "Ship it" with the labels "sp26-20" and "release"
      When I run `cairn label list --json`
      Then it writes JSON to standard output
      And the JSON gives the label "release" carried by 1 task and 0 objectives

    @label-list-says-when-there-is-none
    Scenario: a store with no label says so
      Given I added a task titled "Write the spec"
      When I run `cairn label list`
      Then it succeeds
      And it says there is no label

  Rule: A label is deleted from everything carrying it, once that is confirmed

    @label-delete-takes-it-off-everything
    Scenario: deleting a label takes it off every task and every objective
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Ship it" with the labels "sp26-20" and "release"
      And I ran `cairn objective new "Ship cairn" --label sp26-20`
      When I run `cairn label delete sp26-20 --yes`
      Then it succeeds
      And task 1 carries exactly the labels "doc"
      And task 2 carries exactly the labels "release"
      And objective 1 carries exactly the labels ""

    @label-delete-asks-first
    Scenario: deleting a label is not done until it is confirmed
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      When I run `cairn label delete sp26-20`
      Then it fails
      And it says how to confirm
      And task 1 carries exactly the labels "doc, sp26-20"

    @label-delete-refuses-what-nothing-carries
    Scenario: deleting a label nothing carries is refused
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      When I run `cairn label delete nope --yes`
      Then it fails
      And it says nothing carries the label "nope"

  Rule: A label is renamed on everything carrying it

    @label-rename-everywhere
    Scenario: renaming a label renames it on every task and every objective
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I ran `cairn objective new "Ship cairn" --label sp26-20`
      When I run `cairn label rename sp26-20 sprint-20`
      Then it succeeds
      And task 1 carries exactly the labels "doc, sprint-20"
      And objective 1 carries exactly the labels "sprint-20"

    @label-rename-writes-the-new-name-the-one-way
    Scenario: the new name is written the one way a label is written
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      When I run `cairn label rename sp26-20 "Sprint 20"`
      Then task 1 carries exactly the labels "doc, sprint-20"

    @label-rename-onto-another-joins-them
    Scenario: renaming a label to one already there leaves one
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      When I run `cairn label rename doc sp26-20`
      Then task 1 carries exactly the labels "sp26-20"

  Rule: A label can be made before anything carries it

    @label-new-stands-on-its-own
    Scenario: a label made on its own is listed, carried by nothing
      When I run `cairn label new "Release Train"`
      Then it succeeds
      When I run `cairn label list`
      Then the label "release-train" is carried by 0 tasks and 0 objectives

    @label-new-stays-once-carried-and-dropped
    Scenario: a label made on its own stays when the last task carrying it lets it go
      Given I ran `cairn label new release`
      And I added a task titled "Ship it" with the labels "release" and "doc"
      When I run `cairn task edit 1 --label -release`
      And I run `cairn label list`
      Then the label "release" is carried by 0 tasks and 0 objectives

    @label-new-goes-with-delete-and-follows-rename
    Scenario: a label made on its own is renamed and deleted like any other
      Given I ran `cairn label new release`
      When I run `cairn label rename release train`
      And I run `cairn label list`
      Then the label "train" is carried by 0 tasks and 0 objectives
      When I run `cairn label delete train --yes`
      Then it succeeds
      And it says there is no label

    @label-new-refuses-what-the-rule-refuses
    Scenario: a label the rule refuses is not made
      When I run `cairn label new "a:b"`
      Then it fails
