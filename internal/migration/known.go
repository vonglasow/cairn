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
	{
		name: "what happened to a task is written down",
		does: `
CREATE TABLE IF NOT EXISTS trace (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	task_number INTEGER NOT NULL,
	did         TEXT    NOT NULL,
	field       TEXT    NOT NULL DEFAULT '',
	was         TEXT    NOT NULL DEFAULT '',
	became      TEXT    NOT NULL DEFAULT '',
	through     TEXT    NOT NULL DEFAULT '',
	at          TEXT    NOT NULL
);
CREATE INDEX IF NOT EXISTS trace_by_task ON trace (task_number, id);`,
		undoes: `
DROP INDEX IF EXISTS trace_by_task;
DROP TABLE IF EXISTS trace;`,
		loses: "everything that was ever done to a task",
	},
	{
		name: "a task says what blocks it",
		does: `
ALTER TABLE tasks ADD COLUMN because TEXT NOT NULL DEFAULT '';
CREATE TABLE IF NOT EXISTS blockers (
	task_number    INTEGER NOT NULL,
	blocker_number INTEGER NOT NULL,
	PRIMARY KEY (task_number, blocker_number)
);
CREATE INDEX IF NOT EXISTS blockers_by_blocker ON blockers (blocker_number);`,
		undoes: `
DROP INDEX IF EXISTS blockers_by_blocker;
DROP TABLE IF EXISTS blockers;
ALTER TABLE tasks DROP COLUMN because;`,
		loses: "what blocks each task",
	},
	{
		name: "a task can serve more than one objective",
		does: `
CREATE TABLE IF NOT EXISTS task_objectives (
	task_number      INTEGER NOT NULL,
	objective_number INTEGER NOT NULL,
	PRIMARY KEY (task_number, objective_number)
);
INSERT OR IGNORE INTO task_objectives (task_number, objective_number)
	SELECT number, objective FROM tasks WHERE objective <> 0;
ALTER TABLE tasks DROP COLUMN objective;`,
		undoes: `
ALTER TABLE tasks ADD COLUMN objective INTEGER NOT NULL DEFAULT 0;
UPDATE tasks SET objective = COALESCE((
	SELECT MIN(objective_number) FROM task_objectives
	WHERE task_objectives.task_number = tasks.number), 0);
DROP TABLE IF EXISTS task_objectives;`,
		loses: "every objective a task served beyond the first",
	},
	{
		name: "more than one task can be in focus",
		does: `
CREATE TABLE IF NOT EXISTS focused (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	task_number INTEGER NOT NULL UNIQUE
);
INSERT OR IGNORE INTO focused (task_number)
	SELECT task_number FROM focus WHERE task_number IS NOT NULL;
DROP TABLE IF EXISTS focus;`,
		undoes: `
CREATE TABLE IF NOT EXISTS focus (
	id          INTEGER PRIMARY KEY CHECK (id = 1),
	task_number INTEGER REFERENCES tasks(number) ON DELETE SET NULL
);
INSERT OR IGNORE INTO focus (id, task_number)
	VALUES (1, (SELECT task_number FROM focused ORDER BY id LIMIT 1));
DROP TABLE IF EXISTS focused;`,
		loses: "every task in focus beyond the first one taken",
	},
	{
		name: "a sub-task is done or it is not",
		does: `
UPDATE tasks SET state = 'todo'
	WHERE parent IS NOT NULL AND parent != 0 AND state != 'done';`,
		// Nothing to undo: what column each sub-task stood in is gone by then,
		// and putting them all back in the backlog would be inventing an
		// answer rather than restoring one. The shape is unchanged either way,
		// so an older cairn reads this store as it always did.
		undoes: "",
		loses:  "which column a sub-task stood in, for any that was not done",
	},
	{
		name: "a task can be taken off the board and kept",
		does: `
ALTER TABLE tasks ADD COLUMN archived INTEGER NOT NULL DEFAULT 0;`,
		// The column goes with it: rolling back and applying again is a thing
		// this store does, and the second apply fails on a column left behind.
		undoes: `
ALTER TABLE tasks DROP COLUMN archived;`,
		loses: "which tasks had been taken off the board",
	},
	{
		// docs/labels.md: a label has no family, so one carrying a colon is
		// split at every colon into labels of its own; then every label is
		// written the one way — accents off, spaces, underscores and slashes
		// made dashes, a run of dashes made one, none at either end — and a
		// task or an objective left carrying one twice carries it once.
		name: "a label is written one way, and has no family",
		does: `
WITH RECURSIVE pieces(owner, piece, rest) AS (
	SELECT task_number, '', label || ':' FROM labels WHERE instr(label, ':') > 0
	UNION ALL
	SELECT owner, substr(rest, 1, instr(rest, ':') - 1), substr(rest, instr(rest, ':') + 1)
	FROM pieces WHERE rest <> ''
)
INSERT OR IGNORE INTO labels (task_number, label) SELECT owner, piece FROM pieces WHERE piece <> '';
DELETE FROM labels WHERE instr(label, ':') > 0;
WITH RECURSIVE pieces(owner, piece, rest) AS (
	SELECT objective_number, '', label || ':' FROM objective_labels WHERE instr(label, ':') > 0
	UNION ALL
	SELECT owner, substr(rest, 1, instr(rest, ':') - 1), substr(rest, instr(rest, ':') + 1)
	FROM pieces WHERE rest <> ''
)
INSERT OR IGNORE INTO objective_labels (objective_number, label) SELECT owner, piece FROM pieces WHERE piece <> '';
DELETE FROM objective_labels WHERE instr(label, ':') > 0;
CREATE TEMP TABLE written AS
	SELECT label AS was, lower(label) AS becomes FROM labels
	UNION SELECT label, lower(label) FROM objective_labels;
UPDATE written SET becomes =
	replace(replace(replace(replace(replace(replace(becomes, 'à', 'a'), 'á', 'a'), 'â', 'a'), 'ã', 'a'), 'ä', 'a'), 'å', 'a');
UPDATE written SET becomes =
	replace(replace(replace(replace(replace(becomes, 'è', 'e'), 'é', 'e'), 'ê', 'e'), 'ë', 'e'), 'ç', 'c');
UPDATE written SET becomes =
	replace(replace(replace(replace(replace(becomes, 'ì', 'i'), 'í', 'i'), 'î', 'i'), 'ï', 'i'), 'ñ', 'n');
UPDATE written SET becomes =
	replace(replace(replace(replace(replace(becomes, 'ò', 'o'), 'ó', 'o'), 'ô', 'o'), 'õ', 'o'), 'ö', 'o');
UPDATE written SET becomes =
	replace(replace(replace(replace(replace(replace(becomes, 'ù', 'u'), 'ú', 'u'), 'û', 'u'), 'ü', 'u'), 'ý', 'y'), 'ÿ', 'y');
UPDATE written SET becomes = replace(replace(replace(becomes, ' ', '-'), '_', '-'), '/', '-');
UPDATE written SET becomes = replace(becomes, '--', '-') WHERE becomes LIKE '%--%';
UPDATE written SET becomes = replace(becomes, '--', '-') WHERE becomes LIKE '%--%';
UPDATE written SET becomes = replace(becomes, '--', '-') WHERE becomes LIKE '%--%';
UPDATE written SET becomes = replace(becomes, '--', '-') WHERE becomes LIKE '%--%';
UPDATE written SET becomes = replace(becomes, '--', '-') WHERE becomes LIKE '%--%';
UPDATE written SET becomes = trim(becomes, '-');
INSERT OR IGNORE INTO labels (task_number, label)
	SELECT labels.task_number, written.becomes FROM labels JOIN written ON labels.label = written.was
	WHERE written.becomes <> written.was AND written.becomes <> '';
DELETE FROM labels WHERE label IN (SELECT was FROM written WHERE becomes <> was);
INSERT OR IGNORE INTO objective_labels (objective_number, label)
	SELECT objective_labels.objective_number, written.becomes FROM objective_labels JOIN written ON objective_labels.label = written.was
	WHERE written.becomes <> written.was AND written.becomes <> '';
DELETE FROM objective_labels WHERE label IN (SELECT was FROM written WHERE becomes <> was);
DROP TABLE written;`,
		// Nothing to undo: which spellings were one word is gone once they
		// are one label, and the shape is unchanged, so an older cairn reads
		// this store as it always did.
		undoes: "",
		loses:  "how each label was spelt before, and which labels had a family",
	},
}
