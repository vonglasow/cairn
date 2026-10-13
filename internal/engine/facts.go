package engine

import "cairn/internal/task"

func factsOf(t task.Task) task.Carried { return t.Carried() }
