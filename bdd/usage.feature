Feature: Usage

  @usage-bare
  Scenario: cairn on its own says what it can do
    When I run `cairn`
    Then it succeeds
    And it shows how to use it
    And it writes nothing to standard error
