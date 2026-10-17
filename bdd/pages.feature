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

    @pages-all-off-leave-no-command-of-theirs
    Scenario: every page turned off leaves no command of theirs, and the rest
      Given a configuration that turns every page off
      When I run `cairn --help`
      Then the help does not offer "task"
      And the help does not offer "objective"
      And the help does not offer "timesheet"
      And the help offers "config"
      And the help offers "backup"

    @pages-off-takes-the-steps-with-the-tasks
    Scenario: the tasks turned off take their steps with them
      Given a configuration that turns the page "tasks" off
      When I run `cairn --help`
      Then the help does not offer "step"
      And the help offers "label"

    @pages-labels-go-with-the-last-page-they-label
    Scenario: the labels go once neither tasks nor objectives are shown
      Given a configuration that turns the pages "tasks" and "objectives" off
      When I run `cairn --help`
      Then the help does not offer "label"
      And the help offers "timesheet"

    @pages-on-by-default
    Scenario: a configuration that says nothing of the pages shows them all
      Given a configuration that says the notes live at "~/vault"
      When I run `cairn --help`
      Then the help offers "task"
      And the help offers "objective"
      And the help offers "timesheet"

  Rule: Only the pages cairn has are turned off

    @pages-an-unknown-page-is-refused
    Scenario: a page cairn does not have is refused, not passed over
      Given a configuration that turns the page "calendar" off
      When I run `cairn config show`
      Then it succeeds
      And it says the file cannot be made sense of

  Rule: Showing the configuration says which pages are off

    @pages-config-show-says-what-is-off
    Scenario: config show names the pages turned off
      Given a configuration that turns the page "time" off
      When I run `cairn config show`
      Then it says the pages turned off are "time"
