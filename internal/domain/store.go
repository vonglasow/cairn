package domain

import "time"

// Held is what a store, or a backup of one, contains. Tasks and objectives are
// counted apart for the same reason they are everywhere else: they are not the
// same thing, and one total would hide which of the two somebody is about to
// overwrite.
type Held struct {
	Tasks      int
	Objectives int
}

// Backup is a backup of the store, kept beside it and referred to by its number.
// It lives here rather than beside the code that writes it, because a surface
// has to name one without being allowed to know how a store is made.
type Backup struct {
	Number int
	Made   time.Time
	Path   string
}

// Migration is one step in the shape of the store: what it does, how it is
// undone, what undoing it costs, and whether it has been applied yet. It
// carries its SQL so somebody can read what a tool is about to do to their data
// before letting it, and Loses says in plain words what the older shape cannot
// hold — the SQL says `DROP COLUMN parent`, which is only the same thing to
// somebody who already knows the schema.
type Migration struct {
	Number  int
	Name    string
	Does    string
	Undoes  string
	Loses   string
	Applied bool
	Made    time.Time
}
