package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// Called finds the operation two words name: the thing, then the verb.
func Called(thing, verb string) (Operation, bool) {
	for _, op := range All() {
		if op.Thing == thing && op.Verb == verb {
			return op, true
		}
	}
	return Operation{}, false
}

func (o Operation) argument(name string) (Argument, bool) {
	for _, arg := range o.Takes {
		if arg.Name == name {
			return arg, true
		}
	}
	return Argument{}, false
}

// Read turns the words a person typed after the two that name the operation into what it takes: the positional arguments in the order it lists them — a positional list taking every word left — then `--name value` or `--name=value` for the rest.
func (o Operation) Read(words []string) (Input, error) {
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	var positional []Argument
	for _, arg := range o.Takes {
		if arg.Positional {
			positional = append(positional, arg)
		}
	}
	at := 0
	for i := 0; i < len(words); i++ {
		word := words[i]
		if !strings.HasPrefix(word, "--") {
			if at >= len(positional) {
				return Input{}, Misread(fmt.Errorf("%q: %s takes no argument there", word, o.Line()))
			}
			if positional[at].Kind == List {
				in.Listed[positional[at].Name] = append(in.Listed[positional[at].Name], word)
				continue
			}
			in.Given[positional[at].Name] = word
			at++
			continue
		}
		name, value, said := strings.Cut(strings.TrimPrefix(word, "--"), "=")
		arg, _ := o.argument(name)
		if arg.Kind == Bool {
			if !said {
				value = "true"
			}
			if _, err := strconv.ParseBool(value); err != nil {
				return Input{}, Misread(fmt.Errorf("--%s=%s: true or false is expected", name, value))
			}
			in.Given[name] = value
			continue
		}
		if !said {
			if i+1 >= len(words) {
				return Input{}, Misread(fmt.Errorf("--%s: nothing was given for it", name))
			}
			i++
			value = words[i]
		}
		if arg.Kind == List {
			in.Listed[name] = append(in.Listed[name], value)
			continue
		}
		in.Given[name] = value
	}
	if err := o.Check(in); err != nil {
		return Input{}, err
	}
	return in, nil
}

// FromJSON turns the arguments an agent sent, as one JSON object, into what the operation takes.
func (o Operation) FromJSON(sent []byte) (Input, error) {
	in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
	if len(bytes.TrimSpace(sent)) == 0 {
		return in, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(sent))
	decoder.UseNumber()
	values := map[string]any{}
	if err := decoder.Decode(&values); err != nil {
		return Input{}, Misread(err)
	}
	for _, name := range slices.Sorted(maps.Keys(values)) {
		value := values[name]
		arg, _ := o.argument(name)
		if arg.Kind == List {
			list, err := stringsOf(value)
			if err != nil {
				return Input{}, Misread(fmt.Errorf("%q: %w", name, err))
			}
			in.Listed[name] = list
			continue
		}
		if value == nil {
			in.Given[name] = ""
			continue
		}
		if arg.Kind == Bool {
			yes, ok := value.(bool)
			if !ok {
				return Input{}, Misread(fmt.Errorf("%q: true or false is expected, got %#v", name, value))
			}
			in.Given[name] = strconv.FormatBool(yes)
			continue
		}
		in.Given[name] = fmt.Sprint(value)
	}
	return in, nil
}

func stringsOf(value any) ([]string, error) {
	if value == nil {
		return []string{}, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("a list is expected, got %#v", value)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("a string is expected among the list, got %#v", item)
		}
		out = append(out, s)
	}
	return out, nil
}
