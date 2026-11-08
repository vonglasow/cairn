package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// ErrNotAField is what a field a row of a listing does not carry is refused with.
var ErrNotAField = errors.New("not a field a row carries")

func rowFields() []string {
	var names []string
	for _, field := range reflect.VisibleFields(reflect.TypeFor[Row]()) {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

var fieldsArgument = Argument{
	Name:    "fields",
	Kind:    List,
	OneOf:   rowFields(),
	Outside: ErrNotAField,
	Means:   fmt.Sprintf("the fields each task's row carries in the JSON answer, among %s; every one when it is not named", strings.Join(rowFields(), ", ")),
}

func fieldsOf(in Input) ([]string, error) {
	asked, named := in.Listed[fieldsArgument.Name]
	if !named {
		return nil, nil
	}
	if len(asked) == 0 {
		return nil, Misread(fmt.Errorf("%q: name at least one of %s", fieldsArgument.Name, strings.Join(fieldsArgument.OneOf, ", ")))
	}
	if err := accepted(in, fieldsArgument); err != nil {
		return nil, err
	}
	return asked, nil
}

// MarshalJSON writes every field of the row, or only the ones kept, in the order a row carries them.
func (r Row) MarshalJSON() ([]byte, error) {
	type whole Row
	written, err := json.Marshal(whole(r))
	if err != nil || len(r.kept) == 0 {
		return written, err
	}
	var carried map[string]json.RawMessage
	if err := json.Unmarshal(written, &carried); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteByte('{')
	for _, name := range rowFields() {
		if !slices.Contains(r.kept, name) {
			continue
		}
		if out.Len() > 1 {
			out.WriteByte(',')
		}
		key, _ := json.Marshal(name)
		out.Write(key)
		out.WriteByte(':')
		out.Write(carried[name])
	}
	out.WriteByte('}')
	return out.Bytes(), nil
}
