Feature: Usage
  Asked for nothing in particular, cairn says what it can do rather than
  failing, and says it where a result belongs — a person who typed the name
  alone is looking for a way in, not for an error.

  @usage-bare
  Scenario: cairn on its own says what it can do
    When I run `cairn`
    Then it succeeds
    And it shows how to use it
    And it writes nothing to standard error
