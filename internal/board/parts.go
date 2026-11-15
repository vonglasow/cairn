package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/task"
)

// Added is the number of the task an add answered with.
func Added(answer engine.Answer) int {
	if answer.Task == nil {
		return 0
	}
	return task.OfTask(*answer.Task).Number
}
