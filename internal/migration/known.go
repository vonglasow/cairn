package migration

// Known names every change to the shape of the store cairn knows, oldest
// first. Appending one here is what a store then has to be told to apply;
// the name is what gets recorded, so it is written for somebody reading
// `cairn migration list` a year from now.
//
// A change to the shape of a store is SQL, so the text lives here, with the
// notion it describes — internal/store is what executes it, and only it.
var Known = []Migration{
	{
		name: "the shape a store starts with",
		does: `
CREATE TABLE IF NOT EXISTS tasks (
	number      INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	state       TEXT    NOT NULL,
	priority    TEXT    NOT NULL,
	deadline    TEXT    NOT NULL DEFAULT '',
	estimate    INTEGER NOT NULL DEFAULT 0,
	time_spent  INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS labels (
	task_number INTEGER NOT NULL REFERENCES tasks(number) ON DELETE CASCADE,
	label       TEXT    NOT NULL,
	PRIMARY KEY (task_number, label)
);`,
		// Left empty on purpose: this one is the floor, see ErrNothingToRollBack
		// in internal/store.
		undoes: "",
	},
	{
		name:   "a task can hang under another",
		does:   `ALTER TABLE tasks ADD COLUMN parent INTEGER NOT NULL DEFAULT 0`,
		undoes: `ALTER TABLE tasks DROP COLUMN parent`,
		loses:  "which task each sub-task hangs under",
	},
	{
		name: "objectives, and the tasks that serve them",
		does: `
CREATE TABLE IF NOT EXISTS objectives (
	number      INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	state       TEXT    NOT NULL,
	priority    TEXT    NOT NULL,
	deadline    TEXT    NOT NULL DEFAULT '',
	estimate    INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS objective_labels (
	objective_number INTEGER NOT NULL REFERENCES objectives(number) ON DELETE CASCADE,
	label            TEXT    NOT NULL,
	PRIMARY KEY (objective_number, label)
);
ALTER TABLE tasks ADD COLUMN objective INTEGER NOT NULL DEFAULT 0;`,
		undoes: `
DROP TABLE IF EXISTS objective_labels;
DROP TABLE IF EXISTS objectives;
ALTER TABLE tasks DROP COLUMN objective;`,
		loses: "every objective, and which one each task serves",
	},
	{
		name: "a task carries a place, and one task is in focus",
		does: `
ALTER TABLE tasks ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
UPDATE tasks SET position = (
	SELECT COUNT(*) FROM tasks AS earlier
	WHERE earlier.state = tasks.state
	  AND earlier.parent = tasks.parent
	  AND earlier.number <= tasks.number
);
CREATE TABLE IF NOT EXISTS focus (
	id          INTEGER PRIMARY KEY CHECK (id = 1),
	task_number INTEGER REFERENCES tasks(number) ON DELETE SET NULL
);
INSERT OR IGNORE INTO focus (id, task_number) VALUES (1, NULL);`,
		undoes: `
DROP TABLE IF EXISTS focus;
ALTER TABLE tasks DROP COLUMN position;`,
		loses: "the order of the tasks, and which task is in focus",
	},
}
