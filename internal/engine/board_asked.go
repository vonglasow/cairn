package engine

// Turning what a form holds into what an operation takes.
//
// Neither half of that is a surface's business. A field says which argument it
// fills, an operation says what kind that argument is, and what comes out is an
// Input — the same on a terminal and in a window, because the difference
// between them is how a box is drawn.

import (
	"strconv"
	"strings"
)

// Asked turns a form's fields into what the operation takes. Every field is
// named, including the empty ones: a form is the whole of a task, so a field
// left blank is a field somebody emptied, not a field they said nothing about.
//
// Nothing here lists the fields or pairs them up. Each field says which
// argument it fills, and the operation says what kind that argument is — so a
// field added to a form, or an argument added to the engine, arrives without a
// line written here. That is the whole reason Field carries Arg.
func Asked(fields []Field, op Operation) Input {
	kinds := map[string]Kind{}
	for _, argument := range op.Takes {
		kinds[argument.Name] = argument.Kind
	}
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for _, field := range fields {
		// A field the engine has no argument for is the surface's own.
		if field.Arg == "" {
			continue
		}
		switch kinds[field.Arg] {
		case List:
			in.Listed[field.Arg] = Words(field.Value)
		case Number:
			// A number nobody can read is not a change anybody asked for. Empty
			// is not that: it reads as zero, which is how such a field is
			// cleared, and the engine reads it back the same way.
			if _, ok := NumberOf(field.Value); ok {
				in.Given[field.Arg] = strings.TrimSpace(field.Value)
			}
		default:
			in.Given[field.Arg] = field.Value
		}
	}
	return in
}

// NumberOf reads a field back as a task number, and says when there was nothing
// to read. Nothing is not zero: a parent of zero is a task hanging under
// nothing, which is what emptying the field means, so both answer the same
// value and the caller cannot tell them apart — nor does it need to.
func NumberOf(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, true
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return parsed, true
}
