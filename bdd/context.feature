Feature: Where a task's outside context lives

  Background:
    Given a home of my own, made for this run
    And a store of my own, somewhere that is not the default place
    And nothing in it yet
    And a configuration that says the notes live at "~/vault"

  Rule: Showing a task says where its context is

    @context-an-answer-says-where-to-look
    Scenario: looking at a task says where its context lives
      Given I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then it says the context lives at "~/vault/1"

    @context-travels-in-the-json
    Scenario: an agent gets the place in the shape a machine reads
      Given I added a task titled "Write the spec"
      When I ask the server to show task 1
      Then the structured answer names where the context lives

    @context-a-listing-says-nothing-of-it
    Scenario: listing tasks says nothing about where their context is
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I run `cairn task list`
      Then no line says where a context lives

  Rule: The place is worked out from the task, never remembered

    @context-the-same-task-always-gives-the-same-place
    Scenario: the place does not move when the task does
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      And I edited task 1 with the state "blocked"
      And I moved task 1 under task 2
      When I run `cairn task show 1`
      Then it says the context lives at "~/vault/1"

    @context-two-tasks-never-share-a-place
    Scenario: two tasks are told apart
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship it"
      When I run `cairn task show 1`
      And I run `cairn task show 2`
      Then the two places are not the same

    @context-two-stores-are-not-told-apart
    Scenario: the same number in two stores is sent to the same place
      Given I added a task titled "Write the spec"
      And another store of my own, somewhere else again
      And I added a task titled "Ship it"
      When I run `cairn task show 1` against each store
      Then both say the context lives at the same place

  Rule: Cairn says the place, and does nothing about it

    @context-a-place-that-holds-nothing-is-answered-the-same
    Scenario: nothing was ever written there, and the answer does not change
      Given I added a task titled "Write the spec"
      And nothing was ever written under "~/vault/1"
      When I run `cairn task show 1`
      Then it says the context lives at "~/vault/1"
      And it says nothing about whether anything is there

    @context-naming-a-place-makes-nothing
    Scenario: asking about a task leaves the disk as it was
      Given I added a task titled "Write the spec"
      When I run `cairn task show 1`
      Then nothing was made under "~/vault"

    @context-a-task-that-goes-leaves-what-was-written
    Scenario: deleting a task does not take its notes with it
      Given I added a task titled "Write the spec"
      And notes were written under "~/vault/1"
      When I run `cairn task delete 1 --yes`
      Then it succeeds
      And what was written under "~/vault/1" is still there
