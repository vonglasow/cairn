Feature: A page turned off in the configuration is gone from every surface

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: The command line offers nothing of a page turned off

    @pages-off-takes-its-commands
    Scenario: the time page turned off takes the timesheet away, and leaves the rest
      Given a configuration that turns the page "time" off
      When I run `cairn --help`
      Then the help does not offer "timesheet"
      And the help offers "task"
      And the help offers "objective"
      When I run `cairn timesheet show`
      Then it fails
      When I run `cairn task list`
      Then it succeeds

  Rule: Only the pages cairn has are turned off

    @pages-an-unknown-page-is-passed-over
    Scenario: a page cairn does not have is passed over, and said
      Given a configuration that turns the page "calendar" off
      When I run `cairn config show`
      Then it succeeds
      And it says "pages.calendar" is ignored

  Rule: Showing the configuration says which pages are off

    @pages-config-show-says-what-is-off
    Scenario: config show names the pages turned off
      Given a configuration that turns the page "time" off
      When I run `cairn config show`
      Then it says the pages turned off are "time"
