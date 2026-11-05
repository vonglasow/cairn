Feature: cairn upgrades itself

  The newest version of cairn, taken from its releases on GitHub and put in
  the place of the cairn that was run, so whatever runs it from there — the
  terminal, the Mac application, a server Claude Code started — runs the new
  one.

  Background:
    Given a store of my own, somewhere that is not the default place

  Rule: The newest release takes the place of the cairn that was run

    @upgrade-to-the-newest
    Scenario: a newer release is put in its place, and it says so
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it succeeds
      And it says cairn went from v0.10.0 to v0.11.0
      And the cairn in its place is the one released as v0.11.0

    @upgrade-already-the-newest
    Scenario: the newest already, nothing is changed
      Given cairn v0.10.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it succeeds
      And it says v0.10.0 is the newest already
      And the cairn in its place is the one that was there

    @upgrade-answers-json
    Scenario: asked for JSON, it says what it went from, to, and where
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade --json`
      Then it succeeds
      And it says in JSON it went from v0.10.0 to v0.11.0, in its place
      And the cairn in its place is the one released as v0.11.0

    @upgrade-already-the-newest-json
    Scenario: asked for JSON, the newest already went from itself to itself
      Given cairn v0.10.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade --json`
      Then it succeeds
      And it says in JSON it went from v0.10.0 to v0.10.0, in its place
      And the cairn in its place is the one that was there

  Rule: Nothing is put in its place that cannot be trusted to be cairn

    @upgrade-checks-the-sum
    Scenario: an archive that does not match the sum its release gives is refused
      Given cairn v0.11.0 is released, its binary not matching the sum it gives
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it fails
      And it says the archive does not match its sum
      And the cairn in its place is the one that was there

    @upgrade-no-binary-for-this-system
    Scenario: a release with no binary for this system is refused
      Given cairn v0.11.0 is released, with no binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it fails
      And it says the release has no binary for this system
      And the cairn in its place is the one that was there

    @upgrade-releases-out-of-reach
    Scenario: releases that cannot be reached are said to be out of reach
      Given the releases cannot be reached
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it fails
      And it says the releases could not be read
      And the cairn in its place is the one that was there

  Rule: A cairn somebody else keeps is left to them

    @upgrade-built-from-source
    Scenario: a cairn built from its source is built again, not upgraded
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0-3-g4c5a15f, built from its source
      When I run `cairn upgrade`
      Then it fails
      And it says a cairn built from its source is built again
      And the cairn in its place is the one that was there

    @upgrade-kept-by-homebrew
    Scenario: a cairn Homebrew installed is upgraded by Homebrew
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, installed by Homebrew
      When I run `cairn upgrade`
      Then it fails
      And it says to run brew upgrade cairn
      And the cairn in its place is the one that was there

    @upgrade-place-not-writable
    Scenario: a place that cannot be written is said, and left as it was
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release, in a folder I cannot write
      When I run `cairn upgrade`
      Then it fails
      And it says its place cannot be written
      And the cairn in its place is the one that was there

  Rule: Its place is tried before anything is fetched, and left as it was

    @upgrade-place-not-writable-fetches-nothing
    Scenario: a place that cannot be written is refused before anything is fetched
      Given cairn v0.11.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release, in a folder I cannot write
      When I run `cairn upgrade`
      Then it fails
      And nothing was asked of the releases

    @upgrade-tries-its-place-and-leaves-nothing
    Scenario: trying its place leaves nothing beside the cairn
      Given cairn v0.10.0 is released, with a binary for this system
      And the cairn I run is v0.10.0, from a release
      When I run `cairn upgrade`
      Then it succeeds
      And nothing but the cairn is in its folder
