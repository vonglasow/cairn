Feature: The window wears a colour scheme

  The window's colours come from a scheme: one of the nine cairn carries, or a
  base16 file of one's own in ~/.config/cairn/themes. `theme` under [window]
  names it, in the global configuration or in a store's settings, so each
  workspace can wear its own. A scheme cairn cannot use leaves the window in
  the default, and says why; the window always opens.

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet
    And the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"

  Rule: The window wears the scheme its store's settings name, the default when none is named

    @theme-the-default-is-the-look-it-always-had
    Scenario: with no scheme named, the window wears the colours it always wore
      When I open the window
      Then the window wears the colours it always wore

    @theme-a-shipped-scheme-colours-its-workspace
    Scenario: a scheme named in a store's settings colours that workspace's window, and no other
      Given the settings in "~/cairn/home" say:
        """
        [window]
        theme = "gruvbox-dark"
        """
      When I open the window on the workspace "home"
      Then the window's back is "#282828"
      And the window's ink is "#ebdbb2"
      When I open the window on the workspace "work"
      Then the window wears the colours it always wore

    @theme-a-scheme-of-my-own-wins-over-a-shipped-one
    Scenario: a scheme of my own named as a shipped one is the one worn
      Given a scheme of my own "nord" whose back is "#101010"
      And the global configuration also says:
        """
        [window]
        theme = "nord"
        """
      When I open the window
      Then the window's back is "#101010"

  Rule: A scheme cairn cannot use is said, and the window opens in the default

    @theme-an-unknown-scheme-falls-back-and-says-so
    Scenario: a scheme cairn does not see leaves the window in the default
      Given the global configuration also says:
        """
        [window]
        theme = "neon"
        """
      When I open the window
      Then the window wears the colours it always wore
      And it says on standard error that "neon" is not a scheme it sees

    @theme-a-bad-colour-is-said
    Scenario: a scheme holding a colour that is not one leaves the window in the default
      Given a scheme of my own "mine" whose back is "#12345"
      And the global configuration also says:
        """
        [window]
        theme = "mine"
        """
      When I open the window
      Then the window wears the colours it always wore
      And it says on standard error that "#12345" in "mine" is not a colour

  Rule: config show says every scheme cairn sees, and the one in force

    @theme-config-show-lists-the-schemes
    Scenario: config show lists the schemes cairn carries and mine, and says which is worn
      Given a scheme of my own "mine" whose back is "#101010"
      And the settings in "~/cairn/work" say:
        """
        [window]
        theme = "gruvbox-dark"
        """
      When I run `cairn config show`
      Then it says "theme" is "gruvbox-dark", set in the store's settings
      And it lists the schemes "default", "solarized-dark", "solarized-light", "gruvbox-dark", "gruvbox-light", "catppuccin-mocha", "catppuccin-latte", "dracula" and "nord" as shipped
      And it lists the scheme "mine" from my themes folder
      And it says "gruvbox-dark" is the one in force

    @theme-config-show-says-what-cannot-be-used
    Scenario: config show says why the scheme named cannot be worn
      Given the global configuration also says:
        """
        [window]
        theme = "neon"
        """
      When I run `cairn config show`
      Then it says "neon" is not a scheme it sees, and the window wears "default"

    @theme-a-scheme-hard-to-read-is-said
    Scenario: config show warns of a scheme of my own that falls short of a contrast
      Given a scheme of my own "dim" whose back is "#222222" and ink "#444444"
      When I run `cairn config show`
      Then it warns that "dim" falls short, naming "ink.text" on "surface.back" and the ratio it reads at
      And it lists the scheme "dim" from my themes folder

  Rule: Every scheme cairn carries can be read

    @theme-every-shipped-scheme-passes-contrast
    Scenario: no scheme cairn carries falls short of a contrast it is held to
      When I run `cairn config show`
      Then it warns of no scheme falling short
