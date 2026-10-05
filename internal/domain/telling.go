package domain

// TellOrNot tells a value, or that there is none when it is the zero value of
// its kind — an empty text, a length of nothing, a day never set. An absence
// is a fact that gets told, never a silence whoever listens has to notice.
func TellOrNot[T comparable](value T, told func(T), absent func()) {
	var nothing T
	if value == nothing {
		absent()
		return
	}
	told(value)
}

// TellAll tells every member of a set one at a time, or that it is empty, and
// then closes it: whoever listens cannot know a set is complete until it is
// said so.
func TellAll[T any](members []T, each func(T), none func(), closed func()) {
	if len(members) == 0 {
		none()
	}
	for _, one := range members {
		each(one)
	}
	closed()
}
