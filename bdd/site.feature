Feature: the documentation site

  cairn's documentation, published as a site with each release: what it is,
  how to install it, a guide per surface, and a reference of every command,
  every tool an agent sees, every key of the configuration and every scheme.
  Every page comes from docs/ or is read off cairn itself, so nothing on the
  site is written twice.

  Background:
    Given a changelog that says:
      """
      # Changelog

      ## v0.12.0 — 2026-11-13

      ### Features

      - **task:** give on each task the facts that help choose it (e1d9b63)
      """
    When the site is built for version "v0.12.0"

  Rule: A page of docs/ is published when it is for the person using cairn

    @site-pages
    Scenario: the pages for a person using cairn are published
      Then the site has a page for "INSTALL.md"
      And the site has a page for "docs/usage.md"
      And the site has a page for "docs/tui.md"
      And the site has a page for "docs/gui.md"
      And the site has a page for "docs/config.md"
      And the site has a page for "docs/decisions.md"

    @site-developer-pages
    Scenario: a page only whoever develops cairn reads is not published
      Then the site has no page for "docs/workflow.md"
      And the site has no page for "docs/checks.md"
      And the site has no page for "docs/definition-of-done.md"
      And the site has no page for "docs/architecture.md"
      And the site has no page for "docs/design-principles.md"
      And the site has no page for "docs/development.md"
      And the site has no page for "docs/dependencies.md"
      And the site has no page for "docs/release.md"
      And the site has no page for "docs/debug.md"

    @site-every-page-placed
    Scenario: every page of docs/ is either published or kept for developers
      Then every page of docs/ is published or named as a developer's

    @site-links
    Scenario: a link between two pages lands on a page and a heading that exist
      Then every link from one page of the site to another lands on a page that exists
      And every link to a heading lands on a heading of that page

    @site-links-out
    Scenario: a link to what the site does not publish goes to it on GitHub
      Then no link on the site points at a Markdown file of its own
      And the link from "docs/tui.md" to "docs/architecture.md" goes to it on GitHub

  Rule: The reference is read off cairn itself

    @site-commands
    Scenario: every command is in the reference, with what its help says
      Then the reference lists every command cairn has, with the first line of its help
      And the page of `cairn task edit` lists each of its flags with what its help says

    @site-tools
    Scenario: every tool an agent sees is in the reference
      Then the reference lists every MCP tool cairn serves, with what it does

    @site-configuration
    Scenario: every key of the configuration is in the reference
      Then the configuration reference shows the file `cairn config init` writes

    @site-schemes
    Scenario: every scheme cairn carries can colour the site
      Then the site offers every scheme cairn carries
      And a first visit wears "default" when the system is dark, and "catppuccin-latte" when it is light
      And the scheme chosen is read back from the browser before the page is drawn

  Rule: Every page says which version it documents

    @site-version
    Scenario: the version is on every page
      Then every page says "v0.12.0"

    @site-changelog
    Scenario: the changelog is a page of the site
      Then the changelog page says "give on each task the facts that help choose it"

    @site-search
    Scenario: every page can be found from the search
      Then every page is in what the search reads

  Rule: Each board is shown as cairn draws it, in every scheme

    @site-window-shots
    Scenario: the window is drawn over the fixtures in every scheme cairn carries
      Then the site carries a picture of the window in every scheme cairn carries
      And each picture is drawn on the back of its own scheme

    @site-terminal
    Scenario: the terminal board is shown as text, coloured by the scheme worn
      Then the home page shows the terminal board drawn over the fixtures
      And each scheme gives the terminal its sixteen colours
