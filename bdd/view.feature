Feature: Views, the filters of a listing kept under a name

  Background:
    Given a store of my own, somewhere that is not the default place
    And nothing in it yet

  Rule: A view keeps the filters task list takes, under a name

    @view-new-keeps-the-filters
    Scenario: a view kept is a listing opened again by its name
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Wiring diagram" with the objective "1"
      And I added a task titled "Read the sensors" with the objective "2"
      And I added a task titled "Buy milk"
      When I run `cairn view new "Home automation" --objective 1`
      Then it succeeds
      And the answer says "kept view Home automation"
      And the answer says "keeps 2 of 3 tasks"
      When I run `cairn task list --view "Home automation" --json`
      Then it succeeds
      And the JSON lists the tasks 1, 2

  Rule: Flags given beside a view narrow it further

    @view-flags-narrow-further
    Scenario: a flag beside a view keeps less of what the view keeps
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Read the spec" with the labels "doc" and "review"
      And I edited task 2 with the state "todo"
      And I ran `cairn view new Docs --label doc`
      When I run `cairn task list --view Docs --state todo --json`
      Then it succeeds
      And the JSON lists the tasks 2

  Rule: A task outside a view that touches one inside it is named once, and never counted

    @view-outside-named-once-with-every-why
    Scenario: a task outside that blocks one inside and is linked to it is named once, with both
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Upgrade the printer"
      And I edited task 1 with the blocker "2"
      And I edited task 1 with the link "2"
      And I ran `cairn view new Docs --label doc`
      When I run `cairn task list --view Docs --json`
      Then the JSON lists the tasks 1
      And the JSON names task 2 outside, because it "blocks" task 1 and is "linked to" task 1

  Rule: A view is renamed, changed, moved and deleted, and no task changes

    @view-edit-renames-and-changes-the-filters
    Scenario: a view is renamed and its filters changed
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Buy milk" with the labels "shopping" and "home"
      And I ran `cairn view new Shopping --label shopping`
      When I run `cairn view edit Shopping --name Docs --label doc`
      Then it succeeds
      And the answer says "changed view Docs (was Shopping)"
      When I run `cairn task list --view Docs --json`
      Then the JSON lists the tasks 1

  Rule: A view holds an objective and a label by what they are, not by how they are written

    @view-follows-a-label-renamed
    Scenario: a label renamed is renamed in the view holding it
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I ran `cairn view new Sprint --label sp26-20`
      When I run `cairn label rename sp26-20 sprint-20`
      And I run `cairn view show Sprint`
      Then the answer says "label sprint-20"
      When I run `cairn task list --view Sprint --json`
      Then the JSON lists the tasks 1

  Rule: What a view names may be gone, and the view says so

    @view-label-deleted-is-gone
    Scenario: a label deleted is a condition gone, which narrows nothing and is said
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Buy milk" with the state "todo"
      And I ran `cairn view new Docs --label doc --state backlog`
      When I run `cairn label delete doc --yes`
      Then it succeeds
      And the answer says "1 view holds the label doc"
      When I run `cairn task list --view Docs`
      Then it succeeds
      And the answer says "view Docs: label doc is gone and narrows nothing"
      And the answer says "cairn view edit Docs"
      And the answer says "#1  Write the spec"
