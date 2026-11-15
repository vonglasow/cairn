package task

// EveryWithoutTime is the tasks with no time spent said.
func EveryWithoutTime(all []Task) []Task {
	if all == nil {
		return nil
	}
	out := make([]Task, 0, len(all))
	for _, t := range all {
		out = append(out, WithoutTime(t))
	}
	return out
}

// WithoutTime is the task with no time spent said.
func WithoutTime(t Task) Task {
	c := t.Carried()
	c.Spent = 0
	return Made(c)
}
