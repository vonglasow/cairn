Feature: A Claude Code session tied to a cairn task

  @claude-session-focus-writes-the-entry
  Scenario: taking the focus writes the session's entry
    Given a task numbered 1
    When session "s1", working in "/work/project", puts task 1 in focus
    Then the notes of task 1 hold an open entry for session "s1"
    And that entry says "cd /work/project" and "claude --resume s1"
    And the hook said nothing and succeeded

  @claude-session-unfocus-closes-the-link
  Scenario: unfocusing its task closes the link
    Given session "s1" holds task 1
    When session "s1" lets go of task 1
    Then the entry of session "s1" in the notes of task 1 says it was released
    And after a message in session "s1", task 1 is not in focus

  @claude-session-other-acts-keep-the-link
  Scenario Outline: what is not the session's own act leaves the link alone
    Given session "s1" holds task 1
    When session "s1" <does something else>
    Then the entry of session "s1" in the notes of task 1 is still open

    Examples:
      | does something else                          |
      | asks which tasks are in focus                |
      | puts task 99, which does not exist, in focus |
      | lets go of task 2                            |
      | lets go of every task at once                |

  @claude-session-one-task-at-a-time
  Scenario: one session, one task: taking another lets go of the first
    Given a task numbered 2
    And session "s1" holds task 1
    When session "s1" puts task 2 in focus
    Then the entry of session "s1" in the notes of task 1 says it was released
    And task 1 is not in focus
    And the notes of task 2 hold an open entry for session "s1"

  @claude-session-unattached-is-left-alone
  Scenario Outline: a session tied to nothing is left alone
    Given a task numbered 1 in focus
    When session "s9", tied to no task, <has an event>
    Then the hook said nothing and succeeded
    And no notes hold an entry for session "s9"
    And task 1 is in focus

    Examples:
      | has an event       |
      | receives a message |
      | answers            |
      | ends               |

  @claude-session-idle-and-back
  Scenario: an idle session lets go of its task, and the next message takes it back
    Given session "s1" holds task 1
    When session "s1" answers and no message follows within the idle delay
    Then task 1 is not in focus
    And the entry of session "s1" in the notes of task 1 is still open
    When a message comes in session "s1"
    Then task 1 is in focus
    And the hook said nothing and succeeded

  @claude-session-its-own-task-again
  Scenario: taking its own task again keeps it in focus and its link open
    Given session "s1" holds task 1
    When session "s1" puts task 1 in focus
    Then task 1 is in focus
    And the entry of session "s1" in the notes of task 1 is still open

  @claude-session-a-message-keeps-it
  Scenario: a message within the idle delay keeps the focus
    Given session "s1" holds task 1
    When session "s1" answers and a message follows within the idle delay
    Then once the idle delay has passed, task 1 is in focus

  @claude-session-end-releases-only-its-own
  Scenario: a session's end lets go of its own task and nobody else's
    Given session "s1" holds task 1
    And session "s2" holds task 2
    When session "s1" ends
    Then task 1 is not in focus
    And the entry of session "s1" in the notes of task 1 says it was released
    And task 2 is in focus

  @claude-session-shared-task-stays
  Scenario Outline: a task two sessions work on stays in focus while one of them still does
    Given session "s1" holds task 1
    And session "s2" holds task 1
    When session "s1" <leaves>
    Then task 1 is in focus

    Examples:
      | leaves                                               |
      | ends                                                 |
      | answers and no message follows within the idle delay |

  @claude-session-acts-are-the-sessions
  Scenario: what the hook does is written down as the session's doing
    Given session "s1" holds task 1
    When session "s1" answers and no message follows within the idle delay
    Then the trace of task 1 says the focus was let go by "claude session s1"
