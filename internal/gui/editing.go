package gui

// The keys that walk and cut a text the way every text field does: to either
// end of a line, a word at a time, holding as they go with shift, and taking
// out what is after the caret, or a word or a line in one stroke. A field
// that knew only one character at a time to either side was a field nobody
// could correct in a hurry. Both the form's fields and the search read them.

// textKey is what a key does to a text, and whether it does anything.
func textKey(name string) (func(Written) Written, bool) {
	switch name {
	case "home", "cmd+left":
		return Written.LineStart, true
	case "end", "cmd+right":
		return Written.LineEnd, true
	case "alt+left":
		return Written.WordBack, true
	case "alt+right":
		return Written.WordOn, true
	case "shift+home", "cmd+shift+left":
		return holdingTo(Written.LineStart), true
	case "shift+end", "cmd+shift+right":
		return holdingTo(Written.LineEnd), true
	case "alt+shift+left":
		return holdingTo(Written.WordBack), true
	case "alt+shift+right":
		return holdingTo(Written.WordOn), true
	case "delete-forward":
		return Written.Forward, true
	case "cmd+backspace":
		return backTo(Written.LineStart), true
	case "alt+backspace":
		return backTo(Written.WordBack), true
	}
	return nil, false
}

// holdingTo holds from where the anchor is to where a walk would put the caret.
func holdingTo(walk func(Written) Written) func(Written) Written {
	return func(t Written) Written { return t.Selects(walk(t).At) }
}

// backTo takes out everything from where a walk would put the caret up to it.
func backTo(walk func(Written) Written) func(Written) Written {
	return func(t Written) Written { return t.BackTo(walk(t).At) }
}

// holdingLines holds from the anchor to the line above or below, in a field
// with room for several: the arrows walk its lines, and shift holds on the way.
func (w Window) holdingLines(by int) Window {
	writing := w.form.Writing()
	to := w.walksLines(by).form.Writing().At
	w.form = w.form.Writes(writing.Selects(to))
	return w.inSightOfTheCaret()
}
