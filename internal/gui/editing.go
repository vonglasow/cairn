package gui

import "cairn/internal/board"

func textKey(name string) (func(board.Written) board.Written, bool) {
	switch name {
	case "home", "cmd+left":
		return board.Written.LineStart, true
	case "end", "cmd+right":
		return board.Written.LineEnd, true
	case "alt+left":
		return board.Written.WordBack, true
	case "alt+right":
		return board.Written.WordOn, true
	case "shift+home", "cmd+shift+left":
		return holdingTo(board.Written.LineStart), true
	case "shift+end", "cmd+shift+right":
		return holdingTo(board.Written.LineEnd), true
	case "alt+shift+left":
		return holdingTo(board.Written.WordBack), true
	case "alt+shift+right":
		return holdingTo(board.Written.WordOn), true
	case "delete-forward":
		return board.Written.Forward, true
	case "cmd+backspace":
		return backTo(board.Written.LineStart), true
	case "alt+backspace":
		return backTo(board.Written.WordBack), true
	}
	return nil, false
}

func holdingTo(walk func(board.Written) board.Written) func(board.Written) board.Written {
	return func(t board.Written) board.Written { return t.Selects(walk(t).At) }
}

func backTo(walk func(board.Written) board.Written) func(board.Written) board.Written {
	return func(t board.Written) board.Written { return t.BackTo(walk(t).At) }
}

func (w Window) holdingLines(by int) Window {
	writing := w.Form.Writing()
	to := w.walksLines(by).Form.Writing().At
	w.Form = w.Form.Writes(writing.Selects(to))
	return w.inSightOfTheCaret()
}
