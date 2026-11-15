package engine

import "cairn/internal/store/files"

var windowName = Argument{Name: "name", Kind: Text, Required: true, OneOf: []string{"draft", "sizes"}, Means: "what the window keeps: the draft of a task being added, or the sizes it was left with"}

func (on workedOn) windowFile(name string) (files.Kept, bool) {
	at := Places{Store: on.Path}
	if name == "sizes" {
		return files.Kept(at.Sizes()), KeepsSizes(on.Home)
	}
	return files.Kept(at.Draft()), true
}

// WindowKeep keeps what the window leaves past closing; no surface lists it.
var WindowKeep = Operation{
	Thing: "window",
	Verb:  "keep",
	Short: "Keep what the window leaves past closing",
	Level: write,
	Needs: noStore,
	Takes: []Argument{windowName, {Name: "value", Kind: Text, Required: true, Means: "what is kept, as the window wrote it"}},
	Run: func(on workedOn, in Input) (Answer, error) {
		kept, keeps := on.windowFile(in.Given["name"])
		if !keeps {
			return Answer{}, nil
		}
		return Answer{}, kept.Keep([]byte(in.Given["value"]))
	},
}

// WindowKept says what the window left past closing; no surface lists it.
var WindowKept = Operation{
	Thing:   "window",
	Verb:    "kept",
	Short:   "Say what the window left past closing",
	Level:   Read,
	Needs:   noStore,
	Takes:   []Argument{windowName},
	Answers: "",
	Run: func(on workedOn, in Input) (Answer, error) {
		said := ""
		if kept, keeps := on.windowFile(in.Given["name"]); keeps {
			read, err := kept.Read()
			if err != nil {
				return Answer{}, err
			}
			said = string(read)
		}
		return Answer{Kept: &said}, nil
	},
}

// WindowDrop throws away what the window left past closing; no surface lists it.
var WindowDrop = Operation{
	Thing: "window",
	Verb:  "drop",
	Short: "Throw away what the window left past closing",
	Level: write,
	Needs: noStore,
	Takes: []Argument{windowName},
	Run: func(on workedOn, in Input) (Answer, error) {
		kept, _ := on.windowFile(in.Given["name"])
		return Answer{}, kept.Drop()
	},
}
