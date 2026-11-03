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
