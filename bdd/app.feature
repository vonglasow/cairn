Feature: cairn as a Mac application

  The window, opened from Spotlight or the Dock like any other application:
  a cairn.app in the person's Applications folder that runs the cairn
  installed on the machine.

  Background:
    Given a store of my own, somewhere that is not the default place

  Rule: The application is written in the person's Applications folder

    @app-installed
    Scenario: the application is written, and says where
      When I run `cairn app install`
      Then it succeeds
      And it says the application is in my Applications folder
      And the application runs the window of the cairn that wrote it
      And the application carries cairn's icon and its version
      And the application was signed for this machine

    @app-installed-again
    Scenario: installing again replaces what was there
      Given I ran `cairn app install`
      When I run `cairn app install`
      Then it succeeds
      And the application runs the window of the cairn that wrote it

    @app-only-on-a-mac
    Scenario: elsewhere than on a Mac, there is no application to install
      Given this is not a Mac
      When I run `cairn app install`
      Then it fails
      And it says an application is installed on a Mac only
