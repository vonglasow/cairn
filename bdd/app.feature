Feature: cairn as an application

  The window, opened like any other application: from Spotlight or the Dock
  on a Mac, from the applications menu on a Linux desktop. Either way it runs
  the cairn installed on the machine.

  Background:
    Given a store of my own, somewhere that is not the default place

  Rule: On a Mac, the application is written in the person's Applications folder

    @app-installed
    Scenario: the application is written, and says where
      When I run `cairn app install`
      Then it succeeds
      And it says the application is in my Applications folder
      And the application runs the window of the cairn that wrote it
      And the application carries cairn's icon and its version
      And the application was signed for this machine

    @app-icon-drawn-to-the-edge
    Scenario: the application's icon is drawn to its edges, for macOS to round
      When I run `cairn app install`
      Then it succeeds
      And every image of the application's icon is opaque to its corners

    @app-installed-again
    Scenario: installing again replaces what was there
      Given I ran `cairn app install`
      When I run `cairn app install`
      Then it succeeds
      And the application runs the window of the cairn that wrote it

  Rule: On a Linux desktop, an entry is written in the person's applications menu

    Background:
      Given this is a Linux desktop

    @app-installed-on-linux
    Scenario: the entry is written in the applications menu, and says where
      When I run `cairn app install`
      Then it succeeds
      And it says the entry is in my applications folder
      And the entry starts the window of the cairn that wrote it
      And cairn's logo is in my icon theme at every size
      And the desktop's menu was refreshed

    @app-installed-on-linux-where-the-desktop-keeps-its-data
    Scenario: the entry goes where the desktop says it keeps its data
      Given my desktop keeps its data in a folder of its own
      When I run `cairn app install`
      Then it succeeds
      And it says the entry is in that folder
      And the entry starts the window of the cairn that wrote it

    @app-installed-on-linux-again
    Scenario: installing again on a Linux desktop leaves one entry, the same
      Given I ran `cairn app install`
      When I run `cairn app install`
      Then it succeeds
      And the entry starts the window of the cairn that wrote it
      And the desktop's menu was refreshed twice

    @app-on-linux-an-entry-of-somebody-elses-is-kept
    Scenario: an entry of the same name somebody else wrote is left alone
      Given a desktop entry of cairn's name written by somebody else
      When I run `cairn app install`
      Then it fails
      And it says that entry is left alone
      And the entry is still the one somebody else wrote

    @app-on-linux-without-a-menu-tool
    Scenario: a desktop with no tool to refresh its menu still gets the entry
      Given the desktop has no tool to refresh its menu
      When I run `cairn app install`
      Then it succeeds
      And the entry starts the window of the cairn that wrote it

  Rule: Elsewhere there is no application to install

    @app-only-on-a-mac-or-linux
    Scenario: elsewhere than on a Mac or a Linux desktop, there is no application to install
      Given this is neither a Mac nor Linux
      When I run `cairn app install`
      Then it fails
      And it says an application is installed on a Mac or a Linux desktop only
