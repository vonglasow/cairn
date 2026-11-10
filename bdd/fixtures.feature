Feature: A store full of believable work, from the binary alone

  Somebody looking at a board — to try it, to show it, to see a bug — needs a
  store that holds every case, without the repository or Go.

  Rule: The fixtures are made in a folder of their own, never the store in use

    @fixtures-made
    Scenario: the fixtures fill a new store in the folder named
      When I make the fixtures in a folder of my own
      Then it succeeds
      And it says where the store is, and the line to open it with
      And the store it made holds tasks

    @fixtures-kept
    Scenario: a folder made before is opened again as it was left
      Given I made the fixtures in a folder of my own
      And I added a task titled "Mine" to the store it made
      When I make the fixtures in that folder again
      Then it succeeds
      And the store it made holds a task titled "Mine"

    @fixtures-made-again
    Scenario: --new makes them again from nothing
      Given I made the fixtures in a folder of my own
      And I added a task titled "Mine" to the store it made
      When I make the fixtures in that folder again with --new
      Then it succeeds
      And the store it made holds no task titled "Mine"

    @fixtures-a-folder-it-did-not-make
    Scenario: a folder it did not make is left alone
      Given a folder of my own holding a file
      When I make the fixtures in that folder
      Then it fails
      And it says the folder was not made by cairn fixtures
      And the file in that folder is as it was

    @fixtures-from-a-file
    Scenario: --from fills the store from another file
      Given a file of my own holding a task titled "From my file"
      When I make the fixtures in a folder of my own from that file
      Then it succeeds
      And the store it made holds a task titled "From my file"

    @fixtures-open-the-window
    Scenario: --gui opens the window on the store it made
      When I make the fixtures in a folder of my own with --gui
      Then it succeeds
      And the window was opened on the store it made

  Rule: The fixtures hold every case a board draws

    @fixtures-a-tree-of-objectives
    Scenario: an objective holds sub-objectives, one waiting on another
      When I make the fixtures in a folder of my own
      Then the store it made has an objective holding sub-objectives
      And the store it made has an objective waiting on another

    @fixtures-every-column
    Scenario: every column and every lane holds work
      When I make the fixtures in a folder of my own
      Then the store it made has a task in every column
      And the store it made has tasks in focus and tasks archived

    @fixtures-series
    Scenario: a series comes back, another is stopped, and recurrence is on in their store
      When I make the fixtures in a folder of my own
      Then the store it made has a series with an occurrence open and a day missed
      And the store it made has a series stopped

    @fixtures-most-tasks-serve-an-objective
    Scenario: most tasks serve an objective, and a few serve none on purpose
      When I make the fixtures in a folder of my own
      Then most tasks in the store it made serve an objective
      And the store it made has tasks serving no objective

    @fixtures-objectives-typed
    Scenario: an area holds projects, and labels are for tasks, objectives or both
      When I make the fixtures in a folder of my own
      Then the store it made has an objective labelled "area" holding sub-objectives labelled "project"
      And the store it made has a label for tasks only, one for objectives only, and one carried by both

    @fixtures-a-parent-proposed-as-reached
    Scenario: a parent every sub-objective of which is reached is proposed as reached
      When I make the fixtures in a folder of my own
      Then the store it made has an objective proposed as reached

    @fixtures-a-view-on-a-sub-objective
    Scenario: a view names a sub-objective
      When I make the fixtures in a folder of my own
      Then the store it made has a view naming a sub-objective

    @fixtures-series-every-way
    Scenario: a series counted after done, one ended by its count, and occurrences skipped
      When I make the fixtures in a folder of my own
      Then the store it made has a series counted after done
      And the store it made has a series ended by its count
      And the store it made has a series with an occurrence skipped

    @fixtures-a-colour-of-its-own
    Scenario: the store's settings say the colour it wears
      When I make the fixtures in a folder of my own
      Then the configuration of the store it made says its colour is set in the store's settings
