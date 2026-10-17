Feature: The Claude Code plugin comes with the binary

  @plugin-written-beside-the-store
  Scenario: the binary writes the plugin it carries beside the store
    When I run `cairn plugin write`
    Then it succeeds
    And beside the store is a marketplace offering the plugin "cairn"
    And the plugin written there is the plugin in this repository
    And the plugin written there says the version of the cairn that wrote it
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

  Rule: The plugin written beside the store follows the cairn that serves it

    @plugin-kept-current
    Scenario: the MCP server brings a plugin an older cairn wrote to its own version
      Given the plugin was written beside the store by an older cairn
      When the MCP server starts
      Then the plugin written there says the version of the cairn that wrote it
      And the plugin written there is the plugin in this repository

    @plugin-not-written-unasked
    Scenario: the MCP server writes no plugin where nobody wrote one
      When the MCP server starts
      Then nothing is written beside the store
