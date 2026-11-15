Feature: the documentation site

  cairn's documentation, published as a site with each release: what it does,
  where to download it, what it gives an agent, the guides, and a reference of
  every command, every tool an agent sees, every key of the configuration and
  every scheme.
  Every page comes from docs/ or is read off cairn itself, so nothing on the
  site is written twice.

  Background:
    When the site is built for version "v0.12.0"

  Rule: docs/ is written for the person using cairn, and the site publishes it whole

    @site-pages
    Scenario: the pages for a person using cairn are published
      Then the site has a page for "docs/download.md"
      And the site has a page for "docs/usage.md"
      And the site has a page for "docs/backups.md"
      And the site has a page for "docs/tui.md"
      And the site has a page for "docs/gui.md"
      And the site has a page for "docs/config.md"
      And the site has no page "decisions.html"
      And the site has no page "changelog.html"
      And no page says "make build"

    @site-words
    Scenario: no page speaks of how cairn is built
      Then no page says "surface"
      And no page says "notion"

    @site-nav
    Scenario: the top bar says what cairn does, what it gives an agent, and where to download it
      Then the top bar reads "Features, For your agent, Docs, GitHub, Download"

    @site-home
    Scenario: the home page sends to every guide by task
      Then the home page links to "plan-a-week.html", "follow-an-objective.html" and "plug-in-an-agent.html"

    @site-docs
    Scenario: the docs page leads to every page of docs/ the site publishes
      Then the page "docs.html" links to every page of docs/ the site publishes

    @site-agents
    Scenario: the page for an agent says how to plug in the version it documents
      Then the page "agents.html" says "/plugin marketplace add vonglasow/cairn@v0.12.0"

  Rule: The reference is read off cairn itself

    @site-commands
    Scenario: every command is in the reference, with what its help says
      Then the reference lists every command cairn has, with the first line of its help
      And the page of `cairn task edit` lists each of its flags with what its help says
      And the reference lists "app", "debug", "fixtures", "migration" and "plugin" under "Upkeep", after the others

    @site-keys
    Scenario: every key of the configuration is in the reference, with what it takes
      Then the reference lists every key of the configuration, with what it takes and what it says

  Rule: Every page says which version it documents

    @site-version
    Scenario: the version is on every page
      Then every page says "v0.12.0"

  Rule: Each board is shown as cairn draws it, in every scheme

    @site-window-shots
    Scenario: the window is drawn over the fixtures in every scheme cairn carries
      Then the site carries a picture of the window in every scheme cairn carries
      And each picture is drawn on the back of its own scheme
      And every guide by task opens on a picture of the window of its own

  Rule: The README starts somebody, and sends them to the site for the rest

    @site-readme-points-to-the-site
    Scenario: every link from the README to the site lands on one of its pages
      Then the README links to the site's page for "docs/config.md"
      And every link from the README to the site lands on a page of the site
