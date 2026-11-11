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

    @view-new-holds-a-search
    Scenario: a view holds a search beside its filters
      Given I added a task titled "Write the spec"
      And I added a task titled "Ship the spec"
      And I added a task titled "Buy milk"
      When I run `cairn view new Spec --search spec --state backlog`
      Then it succeeds
      When I run `cairn task list --view spec --json`
      Then the JSON lists the tasks 1, 2

    @view-new-holds-the-focus
    Scenario: a view holds the tasks in focus, or the ones that are not
      Given I added a task titled "Write the spec"
      And I added a task titled "Fix the gate"
      And I put task 2 in focus
      When I run `cairn view new Now --focus`
      Then it succeeds
      When I run `cairn task list --view now --json`
      Then the JSON lists the tasks 2
      When I run `cairn view edit Now --no-focus`
      Then it succeeds
      When I run `cairn task list --view now --json`
      Then the JSON lists the tasks 1

    @view-new-refuses-a-name-taken
    Scenario: a name already taken is refused, whatever its case or its accents
      Given I added a task titled "Write the spec"
      And I ran `cairn view new Groceries --state todo`
      When I run `cairn view new grocéries --priority high`
      Then it fails
      And it says on standard error "a view is already named Groceries"
      And `cairn view list` says "Groceries"

    @view-new-refuses-no-filter
    Scenario: a view that would hold no filter is refused
      When I run `cairn view new Everything`
      Then it fails
      And it says on standard error "a view holds at least one filter or a search"
      And `cairn view list` says "there is no view"

    @view-list-names-them-in-their-places
    Scenario: the views are listed in their places, with what each holds and keeps
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Buy milk" with the state "todo"
      And I ran `cairn view new Docs --label doc`
      And I ran `cairn view new Next --state todo`
      When I run `cairn view list`
      Then it succeeds
      And the view "Docs" is listed in place 1, holding "label doc", keeping 1
      And the view "Next" is listed in place 2, holding "state todo", keeping 1

    @view-show-says-what-it-holds-and-keeps
    Scenario: showing a view says what it holds, what it keeps and what lies outside it
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Review the spec"
      And I edited task 1 with the blocker "2"
      And I ran `cairn view new Docs --label doc`
      When I run `cairn view show Docs`
      Then it succeeds
      And the answer says "label doc"
      And the answer says "keeps 1 of 2 tasks"
      And the answer says "#2 blocks #1"

    @view-show-gives-the-objectives-figures-the-stores
    Scenario: showing a view on an objective gives that objective's figures, the store's and not the view's
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Wiring diagram" with the objective "1"
      And I added a task titled "Read the sensors" with the objective "2"
      And I edited task 1 with the state "done"
      And I ran `cairn view new Home --objective 1 --state backlog`
      When I run `cairn view show Home`
      Then it succeeds
      And the answer says "keeps 1 of 2 tasks"
      And the answer says "objective #1 Home automation  1/2 · 50%"
      And the answer says "(the store's)"
      When I run `cairn view show Home --json`
      Then it succeeds
      And the JSON gives objective 1 as 1 of 2 tasks done, the store's

    @view-unknown-is-refused
    Scenario Outline: a view nobody kept is refused by name
      When I run `<command>`
      Then it fails
      And it says on standard error "no view is named Gardn"

      Examples:
        | command                       |
        | cairn task list --view Gardn  |
        | cairn view show Gardn         |
        | cairn view edit Gardn --state todo |
        | cairn view delete Gardn       |

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

    @view-outside-waits-on
    Scenario: a task outside that waits on one inside is named outside
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Publish the spec"
      And I edited task 2 with the blocker "1"
      And I ran `cairn view new Docs --label doc`
      When I run `cairn task list --view Docs`
      Then the answer says "outside this view"
      And the answer says "#2  Publish the spec"
      And the answer says "waits on #1"

    @view-outside-not-without-a-view
    Scenario: a listing without a view names nothing outside
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I added a task titled "Publish the spec"
      And I edited task 2 with the blocker "1"
      When I run `cairn task list --label doc`
      Then the answer does not say "outside this view"

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

    @view-edit-moves-its-place
    Scenario: a view moved to another place moves the others along
      Given I added a task titled "Write the spec" with the state "todo"
      And I ran `cairn view new First --state todo`
      And I ran `cairn view new Second --state backlog`
      And I ran `cairn view new Third --state done`
      When I run `cairn view edit Third --place 1`
      Then it succeeds
      When I run `cairn view list`
      Then the view "Third" is listed in place 1, holding "state done", keeping 0
      And the view "First" is listed in place 2, holding "state todo", keeping 1
      And the view "Second" is listed in place 3, holding "state backlog", keeping 0

    @view-edit-refuses-nothing-to-change
    Scenario: editing a view without naming anything to change is refused
      Given I ran `cairn view new Next --state todo`
      When I run `cairn view edit Next`
      Then it fails
      And it says on standard error "nothing was named to change"

    @view-delete-touches-no-task
    Scenario: deleting a view changes no task, and the places close up
      Given I added a task titled "Write the spec" with the state "todo"
      And I ran `cairn view new First --state todo`
      And I ran `cairn view new Second --state backlog`
      When I run `cairn view delete First`
      Then it succeeds
      And the answer says "deleted view First · no task changed"
      And `cairn task list` lists 1 task
      And `cairn view list` says "1  Second"

    @view-change-writes-no-trace
    Scenario: keeping, changing and deleting a view writes nothing in any task's trace
      Given I added a task titled "Write the spec" with the state "todo"
      And I ran `cairn view new Next --state todo`
      And I ran `cairn view edit Next --name Soon`
      And I ran `cairn view delete Soon`
      When I run `cairn trace show 1`
      Then it carries 1 act

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

    @view-follows-the-tree
    Scenario: a view on an objective reaches what sits under it the day it is opened
      Given I added an objective titled "Home automation"
      And I ran `cairn view new Home --objective 1`
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Read the sensors" with the objective "2"
      When I run `cairn task list --view Home --json`
      Then the JSON lists the tasks 1

    @view-holds-an-objective-itself
    Scenario: a view on an objective itself leaves out its sub-objectives
      Given I added an objective titled "Home automation"
      And I added an objective titled "Firmware" under objective 1
      And I added a task titled "Wiring diagram" with the objective "1"
      And I added a task titled "Read the sensors" with the objective "2"
      And I ran `cairn view new Home --objective 1 --itself`
      When I run `cairn task list --view Home --json`
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

    @view-every-condition-gone-keeps-nothing
    Scenario: a view whose every condition is gone keeps nothing, and says why
      Given I added a task titled "Write the spec" with the labels "doc" and "sp26-20"
      And I ran `cairn view new Docs --label doc`
      And I ran `cairn label delete doc --yes`
      When I run `cairn task list --view Docs --json`
      Then it succeeds
      And the JSON lists no task
      And the JSON says the view lost "label doc"
      When I run `cairn task list --view Docs`
      Then the answer says "every condition view Docs holds is gone, so it keeps nothing"
      And the answer says "cairn view delete Docs"

    @view-objective-delete-lists-the-views
    Scenario: deleting an objective a view names lists the view and asks what becomes of it
      Given I added an objective titled "Home automation"
      And I ran `cairn view new Home --objective 1`
      When I run `cairn objective delete 1 --yes`
      Then it fails
      And it says on standard error "view Home"
      And it says on standard error "--views off|delete"
      And `cairn objective list` lists 1 objective

    @view-objective-delete-takes-it-off-the-views
    Scenario: the objective is taken off the views naming it
      Given I added an objective titled "Home automation"
      And I added a task titled "Write the spec" with the state "todo"
      And I ran `cairn view new Home --objective 1 --state todo`
      When I run `cairn objective delete 1 --views off --yes`
      Then it succeeds
      When I run `cairn view show Home`
      Then the answer says "state todo"
      And the answer does not say "objective"

    @view-objective-delete-keeps-a-view-holding-nothing-else
    Scenario: a view that held only the objective keeps it, gone, and keeps nothing
      Given I added an objective titled "Home automation"
      And I added a task titled "Write the spec"
      And I ran `cairn view new Home --objective 1`
      When I run `cairn objective delete 1 --views off --yes`
      Then it succeeds
      When I run `cairn task list --view Home --json`
      Then the JSON lists no task
      And the JSON says the view lost "objective 1"

    @view-objective-delete-deletes-the-views
    Scenario: the views naming the objective are deleted with it
      Given I added an objective titled "Home automation"
      And I ran `cairn view new Home --objective 1`
      And I ran `cairn view new Next --state todo`
      When I run `cairn objective delete 1 --views delete --yes`
      Then it succeeds
      And `cairn view list` says "1  Next"
      When I run `cairn view show Home`
      Then it fails
