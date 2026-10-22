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

  @plugin-has-one-place
  Scenario: the plugin is written in one place, and no command names another
    When I run `cairn plugin write --dir somewhere`
    Then it fails, saying the flag is unknown

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

    @plugin-kept-current-where-the-configuration-says
    Scenario: the MCP server keeps current the plugin in the folder the configuration names
      Given a configuration that says the plugin lives in a folder of my own
      And the plugin was written there by an older cairn
      When the MCP server starts
      Then the plugin in that folder says the version of the cairn that wrote it

  @plugin-the-skill-says-open-a-linked-task
  Scenario: the skill tells an agent to open a linked task rather than reopen one done
    When I read the cairn skill
    Then the skill says "A task done whose work does not hold is not reopened: open a new task linked to it"
