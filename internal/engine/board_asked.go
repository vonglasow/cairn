package engine

import (
	"strconv"
	"strings"
)

// Asked turns a form's fields into what the operation takes.
func Asked(fields []Field, op Operation) Input {
	kinds := map[string]Kind{}
	for _, argument := range op.Takes {
		kinds[argument.Name] = argument.Kind
	}
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	for _, field := range fields {
		if field.Arg == "" {
			continue
		}
		switch kinds[field.Arg] {
		case List:
			in.Listed[field.Arg] = Words(field.Value)
		case Number:
			if _, ok := NumberOf(field.Value); ok {
				in.Given[field.Arg] = strings.TrimSpace(field.Value)
			}
		default:
			in.Given[field.Arg] = field.Value
		}
	}
	return in
}

// NumberOf reads a field back as a task number, and says when there was nothing to read.
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
