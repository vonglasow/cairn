package gui

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

func holdingTo(walk func(Written) Written) func(Written) Written {
	return func(t Written) Written { return t.Selects(walk(t).At) }
}

func backTo(walk func(Written) Written) func(Written) Written {
	return func(t Written) Written { return t.BackTo(walk(t).At) }
}

func (w Window) holdingLines(by int) Window {
	writing := w.form.Writing()
	to := w.walksLines(by).form.Writing().At
	w.form = w.form.Writes(writing.Selects(to))
	return w.inSightOfTheCaret()
}
