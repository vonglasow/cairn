package gui

// ownsConsole says whether the window started on a console of its own: the one
// Windows opens for a console program started from the Explorer, with nothing
// but this process on it. That one is let go of, and its window goes with it;
// a console shared with the shell the command was typed in stays.
func ownsConsole(attached uint32) bool { return attached == 1 }
