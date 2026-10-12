Feature: The Claude Code plugin comes with the binary
  The plugin was installed from this repository, which a person who was handed
  only the binary — by brew, or as a file — does not have. The binary carries
  the plugin it was built with, and writes it out as a marketplace Claude Code
  installs from: one command, and the plugin always matches the cairn it calls.

  @plugin-written-beside-the-store
  Scenario: the binary writes the plugin it carries beside the store
    When I run `cairn plugin write`
    Then it succeeds
    And beside the store is a marketplace offering the plugin "cairn"
    And the plugin written there is the plugin in this repository
    And the hook written there can be run
    And it says to run "/plugin marketplace add" and "/plugin install cairn@cairn"

  @plugin-written-where-told
  Scenario: the plugin is written into the folder it is told
    When I write the plugin into a folder of my own
    Then it succeeds
    And that folder is a marketplace offering the plugin "cairn"

  @plugin-written-again
  Scenario: writing the plugin again leaves nothing an older cairn wrote
    Given the plugin was written beside the store with a file it no longer carries
    When I run `cairn plugin write`
    Then it succeeds
    And the plugin written there is the plugin in this repository
