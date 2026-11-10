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

    @settings-config-show-says-the-global-value
    Scenario: beside a key the store's settings set, config show says what the global configuration would give
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

        [time]
        weekends = "hidden"
        """
      When I run `cairn config show`
      Then it says "session" is "25m", set in the store's settings, the global configuration giving "1h"
      And it says "weekends" is "hidden", set in the store's settings, the global configuration giving "shown"
      When I run `cairn config show --json`
      Then the structured answer says the global configuration gives "focus.session" as "1h"
      And the structured answer says the global configuration gives "time.weekends" as "shown"

    @settings-config-show-says-the-colour
    Scenario: config show says the colour of the workspace, and where it comes from
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "work"
      And the settings in "~/cairn/home" say:
        """
        [workspace]
        colour = "teal"
        """
      When I run `cairn config show --workspace home`
      Then it says "colour" is "teal", set in the store's settings
      When I run `cairn config show`
      Then it says the colour is read off the workspace's name
      When I run `cairn config show --workspace home --json`
      Then the structured answer says the workspace wears "teal"

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

  Rule: A key is set, or taken out, from the command line, the file edited line by line

    The configuration stays one, in one place: the window, the terminal and
    the command line are several ways in to it, and the engine keeps it. A
    surface sends what was asked; the engine checks it and writes it, one line
    of the file and nothing else, every comment kept.

    @config-set-writes-the-settings-of-the-workspace
    Scenario: a key set is written in the settings of the workspace in use, at the end of its table, every comment kept
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the settings in "~/cairn/home" say:
        """
        # home: what this board does differently
        [pages]
        # no objectives at home
        objectives = false

        [window]
        theme = "nord"   # told apart from work at a glance
        """
      When I run `cairn config set pages.time hidden`
      Then it succeeds
      And it says it wrote the settings in "~/cairn/home"
      When I run `cairn config set window.theme gruvbox-dark`
      Then it succeeds
      And the settings in "~/cairn/home" read:
        """
        # home: what this board does differently
        [pages]
        # no objectives at home
        objectives = false
        time = "hidden"

        [window]
        theme = "gruvbox-dark"   # told apart from work at a glance
        """
      When I run `cairn config show`
      Then it says "time" is "hidden", set in the store's settings

    @config-set-global-takes-a-key-out-of-its-comment
    Scenario: --global writes the global configuration, a key written commented taken out of its comment in place
      Given `cairn config init` wrote the configuration
      When I run `cairn config set focus.session 45m --global`
      Then it succeeds
      And the global configuration holds `session = "45m"` right under `[focus]`
      When I run `cairn config show`
      Then it says "session" is "45m"

    @config-unset-takes-the-line-out
    Scenario: a key taken out loses its line and nothing else, and the value under it answers again
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        # no objectives at home
        objectives = false
        """
      When I run `cairn config unset pages.objectives`
      Then it succeeds
      And it says "pages.objectives is no longer said"
      And it says "true, by default"
      And the settings in "~/cairn/home" read:
        """
        [pages]
        # no objectives at home
        """

    @config-unset-refuses-a-key-not-said-there
    Scenario: a key the file does not say is not taken out of it
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      When I run `cairn config unset pages.time`
      Then it fails
      And it says on standard error "pages.time is not said in"

    @config-set-refuses-what-the-key-does-not-take
    Scenario Outline: a value the key does not take is refused in words naming the key, and nothing is written
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        objectives = false
        """
      When I run `cairn config set <key> <value>`
      Then it fails
      And it says on standard error "<key>"
      And it says on standard error "<said>"
      And the settings in "~/cairn/home" are as they were

      Examples:
        | key              | value  | said                                             |
        | pages.time       | hiden  | is not a word pages.time takes (on, hidden, off) |
        | focus.session    | 1hour  | is not a duration — write 1h, 45m, 1h30m         |
        | recurrence.on    | yes    | is true or false                                 |
        | archive.days     | 0      | is not a number of days, one or more             |
        | archive.start    | monday | is not a day — write 2026-10-21                  |
        | window.theme     | neon   | is not a scheme cairn sees                       |
        | workspace.colour | puce   | puce                                             |
        | pages.tiem       | on     | is not a key cairn sets                          |

    @config-set-a-key-of-the-machine-goes-in-the-global-file
    Scenario: a key of the machine is written in the global configuration, whatever workspace is in use
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      When I run `cairn config set plugin ~/plugins/cairn`
      Then it succeeds
      And the global configuration holds `plugin = "~/plugins/cairn"`

    @config-set-a-key-of-a-store-is-not-global
    Scenario: the colour a workspace wears is refused in the global configuration
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      When I run `cairn config set workspace.colour green --global`
      Then it fails
      And it says on standard error "workspace.colour belongs to a store"

    @config-set-without-a-file-writes-the-table-alone
    Scenario: with no global file, a key set writes a file holding its table alone
      Given no configuration file
      When I run `cairn config set recurrence.on true --global`
      Then it succeeds
      And the global configuration reads:
        """
        [recurrence]
        on = true
        """

    @config-set-a-loose-store-goes-in-the-global-file
    Scenario: for a store no workspace names and with no settings beside it, a key is set in the global configuration
      Given no configuration file
      When I run `cairn config set pages.time off`
      Then it succeeds
      And the global configuration holds `time = "off"` right under `[pages]`

    @config-set-refuses-a-file-that-does-not-read
    Scenario: a file that cannot be read is not written into, and is said
      Given the workspaces "work" at "~/cairn/work" and "home" at "~/cairn/home", the default being "home"
      And the settings in "~/cairn/home" say:
        """
        [pages]
        time = "hiden"
        """
      When I run `cairn config set pages.objectives false`
      Then it fails
      And it names on standard error the file "~/cairn/home/settings.toml"
      And the settings in "~/cairn/home" are as they were
