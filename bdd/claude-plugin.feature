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

  Rule: The plugin is installed from GitHub, at the tag of the cairn that runs it

    @plugin-installed-from-its-tag
    Scenario: a cairn released says how to install the plugin of its own version
      Given a store of my own, somewhere that is not the default place
      And the cairn I run is v0.12.0, from a release
      When I run `cairn plugin show`
      Then it succeeds
      And the answer says "/plugin marketplace add vonglasow/cairn@v0.12.0"
      And the answer says "/plugin install cairn@cairn"
      And the answer says "/plugin marketplace remove cairn"

    @plugin-installed-from-a-build
    Scenario: a cairn built from its source says to write the plugin it carries, as no tag holds it
      Given a store of my own, somewhere that is not the default place
      And the cairn I run is v0.11.0-79-g616ccfc, built from its source
      When I run `cairn plugin show`
      Then it succeeds
      And the answer says "cairn plugin write"
      And the answer does not say "vonglasow/cairn@"

  @plugin-the-skill-says-open-a-linked-task
  Scenario: the skill tells an agent to open a linked task rather than reopen one done
    When I read the cairn skill
    Then the skill says "A task done whose work does not hold is not reopened: open a new task linked to it"

  @plugin-the-skill-holds-several-tasks
  Scenario: the skill tells an agent how to work through several tasks in one session
    When I read the cairn skill
    Then the skill says "Focus each task when its work starts, and keep it while its work runs"
    And the skill says "A subagent does not hold the focus: the session that sends it does"

  @plugin-the-skill-says-what-closing-writes
  Scenario: the skill tells an agent what closing a task writes
    When I read the cairn skill
    Then the skill says "a dated entry naming what was delivered"
    And the skill says "summary.md rewritten to say it is finished"

  @plugin-the-skill-says-what-a-decision-changes
  Scenario: the skill tells an agent what a decision that changes a task does to it
    When I read the cairn skill
    Then the skill says "the title becomes the work decided"
    And the skill says "the question asked and the decision go in a dated entry"

  @plugin-the-skill-says-the-traps-of-the-command-line
  Scenario: the skill tells an agent what the command line does that it would not expect
    When I read the cairn skill
    Then the skill says "--label =x"
    And the skill says "--replace"
    And the skill says "When the MCP server answers an error"

  @plugin-the-skill-reads-the-approach-label
  Scenario: the skill gives the approach label its meaning, only when the configuration names one
    When I read the cairn skill
    Then the skill says "[approach] label"
    And the skill says "Without one, say nothing of it"

  @plugin-the-skill-asks-what-cairn-answers
  Scenario: every call the skill writes out whole is one the MCP server answers
    When I read the cairn skill
    Then the skill writes out at least one call whole
    And the MCP server answers every call the skill writes out whole

  @plugin-the-skill-says-the-server-keeps-one-workspace
  Scenario: the skill tells an agent its server answers from one workspace, and switching is the person's
    When I read the cairn skill
    Then the skill says "answers from one workspace"
    And the skill says "`workspace_show {}`"
    And the skill says "Switching is the person's"

  @plugin-the-skill-says-what-abandoning-writes
  Scenario: the skill tells an agent that a task abandoned is deleted, its reasons kept in its notes
    When I read the cairn skill
    Then the skill says "A task abandoned is deleted"
    And the skill says "a dated entry saying it is abandoned and why"

  Rule: Reading back a day or a week, and where things stand, are skills of the plugin

    @plugin-the-recap-skill-reads-back-the-last-day-worked
    Scenario: the recap skill reads back the last day worked, or a week from Monday to Sunday
      When I read the recap skill
      Then the skill says "the last day before today with at least one act in the trace"
      And the skill says "Monday to Sunday, the previous one when asked on a Monday"
      And the skill says "the time per task"
      And the skill says "the objectives that moved"

    @plugin-the-recap-skill-proposes-the-sessions-no-task-names
    Scenario: the recap skill proposes a task for each Claude Code session the trace does not name, and attaches it on the person's word
      When I read the recap skill
      Then the skill says "the task proposed"
      And the skill says "Attach nothing until the person says"
      And the skill says "sessions-ignored"
      And the skill says "timesheet_add"

    @plugin-the-status-skill-says-where-things-stand
    Scenario: the status skill says where things stand
      When I read the status skill
      Then the skill says "their open steps"
      And the skill says "what is left in the version"
      And the skill says "what waits on the person"

    @plugin-the-other-skills-ask-what-cairn-answers
    Scenario Outline: every call the <skill> skill writes out whole is one the MCP server answers
      When I read the <skill> skill
      Then the skill writes out at least one call whole
      And the MCP server answers every call the skill writes out whole

      Examples:
        | skill  |
        | recap  |
        | status |
