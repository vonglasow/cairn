Feature: Version

  @version-degraded
  Scenario: a binary built without stamps degrades cleanly
    Given a binary built with no version stamps
    When I run `cairn --version`
    Then it reports the version as "dev"
    And it reports the commit as "none"
    And it reports the build time as "unknown"
    And it reports a real Go toolchain
    And it reports a real platform

  @version-stamped
  Scenario: a stamped binary reports where it came from
    Given a binary stamped with version "26-2-preview2-65-g4c5a15f70"
    And stamped with commit "4c5a15f700dcabcdef0123"
    And stamped with build time "2026-08-15T13:59:48Z"
    And stamped with commit date "2026-08-15T13:59:37Z"
    When I run `cairn --version`
    Then it reports the version as "26-2-preview2-65-g4c5a15f70"
    And it reports the commit as "4c5a15f700dc"
    And it reports the build time as "2026-08-15T13:59:48Z"
    And it reports the commit date as "2026-08-15T13:59:37Z"
    And it reports a real Go toolchain
    And it reports a real platform

  @version-from-a-source-archive
  Scenario: a build from a release's source archive says the release it came from
    Given the source archive of this repository made as version "v9.9.9"
    When I unpack it somewhere git has never been
    Then a build there would be stamped with version "v9.9.9"
    And a build there would be stamped with the commit the archive was made from
    And a build there would be stamped with the date of that commit
    And the archive says how to build and install it

  @version-brew-formula
  Scenario: the archive comes with a Homebrew formula that builds it
    Given the source archive of this repository made as version "v9.9.9"
    Then beside the archive is a Homebrew formula for version "9.9.9"
    And the formula names the archive's checksum
    And the formula fetches the archive from where it was made

  @version-brew-formula-hosted
  Scenario: the formula fetches the archive from where it is hosted, when told
    Given the source archive of this repository made as version "v9.9.9" to be fetched from "https://files.example.org/cairn"
    Then the formula fetches "https://files.example.org/cairn/cairn-v9.9.9-src.tar.gz"

  @version-brew-formula-for-its-minor
  Scenario: the archive comes with a formula kept at its minor version too, as the tap keeps every one
    Given the source archive of this repository made as version "v9.9.9"
    Then beside the archive is a Homebrew formula "cairn@9.9" for version "9.9.9"

  @version-binary-archive
  Scenario: a binary is handed over in an archive named for the system it runs on
    Given a binary archive of this repository made as version "v9.9.9" for this system
    Then the archive is named for version "v9.9.9" and this system
    And the binary in it says it is version "v9.9.9"
    And the checksums name it

  @version-install-reads-anywhere
  Scenario: the archive's install guide reads the same in any tool, on any system
    Given the source archive of this repository made as version "v9.9.9"
    When I unpack it somewhere git has never been
    Then the archive's install guide is written in ASCII alone

  @version-is-made-on-main
  Scenario: a version is made from main and nowhere else
    Given a repository on a branch called "work"
    When I make a release there
    Then it refuses, saying "a version is made on main"
    And no tag was made

  @version-is-made-from-a-clean-tree
  Scenario: a version is made from what is committed and nothing else
    Given a repository on a branch called "main"
    And a file changed and not committed
    When I make a release there
    Then it refuses, saying "a version is made from a clean tree"
    And no tag was made

  @version-archive-says-what-changed
  Scenario: a version's archive says what changed, version by version
    Given a repository on a branch called "main"
    And a version "v0.1.0" made there
    And the commits "fix: keep the focus", "feat(gui): choose a label", "feat!: drop families" and "docs: say how"
    When I make its source archive as version "v0.2.0"
    Then the archive's changelog says, under "v0.2.0", "choose a label" among the features
    And the archive's changelog says, under "v0.2.0", "drop families" among the breaking changes
    And the archive's changelog says, under "v0.2.0", "keep the focus" among the fixes
    And the archive's changelog says nothing of "say how"
    And the archive's changelog says, under "v0.1.0", "begin" among the features

  @version-sums-kept
  Scenario: the checksums of every version made are kept beside each other
    Given a repository on a branch called "main"
    And a version "v0.1.0" made there
    And its source archive made as version "v0.1.0"
    And the commits "feat: more", "fix: less" and "docs: words"
    When I make its source archive as version "v0.2.0"
    Then the checksums name "cairn-v0.1.0-src.tar.gz" and "cairn-v0.2.0-src.tar.gz"
