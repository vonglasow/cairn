Feature: Settings that belong to a store, kept beside it

  A workspace's folder holds its store, its notes, and settings.toml. What that
  file says is used for that store instead of the global configuration, key by
  key; the global configuration keeps the workspaces and what belongs to the
  machine.

  Background:
    Given a home of my own, made for this run
    And nothing configured anywhere yet

  Rule: The store's settings win over the global configuration, key by key

    @settings-the-store-wins
    Scenario: a key the store's settings say wins over the global one
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [focus]
        session = "1h"
        """
      And the settings in "~/cairn/work" say:
        """
        [focus]
        session = "25m"
        """
      When I run `cairn config show`
      Then it says "session" is "25m"

    @settings-a-key-not-said-comes-from-the-global-file
    Scenario: a key the store's settings do not say comes from the global configuration
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [time]
        deleted = "taken"
        """
      And the settings in "~/cairn/work" say:
        """
        [time]
        weekends = "hidden"
        """
      When I run `cairn config show`
      Then it says "deleted" is "taken"
      And it says "weekends" is "hidden"

    @settings-two-workspaces-answer-differently
    Scenario: two workspaces with settings of their own answer differently
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        objectives = false
        """
      When I run `cairn objective list`
      Then it succeeds
      When I run `cairn objective list --workspace home`
      Then it fails
      When I run `cairn config show --workspace home`
      Then it says "objectives" is "false"

    @settings-beside-a-store-named
    Scenario: a store named by --store reads the settings beside it
      Given the settings in "~/loose" say:
        """
        [pages]
        time = false
        """
      When I run `cairn config show --store ~/loose/cairn.db`
      Then it says "time" is "hidden"

  Rule: What belongs to the machine stays in the global configuration

    @settings-a-machine-key-is-refused
    Scenario Outline: a key of the machine in the store's settings is refused, naming where it goes
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/work" say:
        """
        <line>
        """
      When I run `cairn task list`
      Then it refuses, and names "<key>"
      And it names the global configuration

      Examples:
        | key        | line                                 |
        | store      | store = "~/elsewhere/cairn.db"       |
        | notes      | notes = "~/vault"                    |
        | plugin     | plugin = "~/plugin"                  |
        | workspaces | workspaces = { play = "~/cairn/play" } |

  Rule: An empty string says nothing, except where nothing is a value

    @settings-an-empty-approach-label-is-none-here
    Scenario: an empty approach label in the store's settings means no label for this store
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [approach]
        label = "discuss"
        """
      And the settings in "~/cairn/work" say:
        """
        [approach]
        label = ""
        """
      When I run `cairn config show`
      Then it says no label asks for the approach to be agreed first
      When I run `cairn config show --workspace home`
      Then it says "label" is "discuss"

    @settings-another-empty-string-is-refused
    Scenario: any other empty string in the store's settings is refused, naming the key
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/work" say:
        """
        [focus]
        session = ""
        """
      When I run `cairn task list`
      Then it refuses, and names "focus.session"

  Rule: A key cairn does not know is passed over, and said

    @settings-an-unknown-key-is-passed-over
    Scenario: a key of the store's settings cairn does not know is passed over, and said
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/work" say:
        """
        [pages]
        calendar = false
        """
      When I run `cairn task list`
      Then it succeeds
      And it says on standard error that "pages.calendar" is ignored
      And it names on standard error the file "~/cairn/work/settings.toml"
      When I run `cairn config show`
      Then it says "pages.calendar" is ignored

  Rule: Showing the configuration says where each value comes from

    @settings-config-show-says-where-each-comes-from
    Scenario: config show says of each key whether the store's settings, the global configuration or the default set it
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [focus]
        session = "1h"
        """
      And the settings in "~/cairn/work" say:
        """
        [time]
        weekends = "hidden"
        """
      When I run `cairn config show`
      Then it names the store's settings "~/cairn/work/settings.toml"
      And it says "weekends" is "hidden", set in the store's settings
      And it says "session" is "1h", from the global configuration
      And it says "deleted" is "kept", by default

    @settings-config-show-as-json
    Scenario: config show tells a machine which keys the store's settings write
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the global configuration also says:
        """
        [focus]
        session = "1h"
        """
      And the settings in "~/cairn/work" say:
        """
        [time]
        weekends = "hidden"
        """
      When I run `cairn config show --json`
      Then the structured answer says "time.weekends" is written in the store's settings
      And the structured answer says "focus.session" is written in the global configuration

  Rule: A workspace added starts with settings that say nothing

    @settings-workspace-add-writes-empty-settings
    Scenario: workspace add writes settings that say nothing yet
      When I run `cairn workspace add play ~/cairn/play`
      Then it succeeds
      And the settings in "~/cairn/play" say nothing, and that everything comes from the global configuration until they say otherwise
      When I run `cairn config show --workspace play`
      Then it succeeds

    @settings-workspace-add-leaves-settings-there
    Scenario: workspace add leaves settings already in the folder as they were
      Given the settings in "~/cairn/play" say:
        """
        [pages]
        time = false
        """
      When I run `cairn workspace add play ~/cairn/play`
      Then it succeeds
      When I run `cairn config show --workspace play`
      Then it says "time" is "hidden"
