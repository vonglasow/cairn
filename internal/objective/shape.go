package objective

import "cairn/internal/domain"

// Shape is how an objective crosses to a machine.
type Shape struct {
	Number      int      `json:"number" jsonschema:"the objective's number, never handed out twice"`
	Title       string   `json:"title" jsonschema:"what the objective is called"`
	State       string   `json:"state" jsonschema:"where it stands: active, on-hold, reached or dropped"`
	Priority    string   `json:"priority" jsonschema:"low, normal, high or critical"`
	Description string   `json:"description" jsonschema:"what is written about it; empty when nothing is"`
	Labels      []string `json:"labels" jsonschema:"the labels it carries, lower case and sorted; empty or null when it carries none"`
	Deadline    string   `json:"deadline" jsonschema:"the day it is due, as 2026-12-01; empty when it has none"`
	Estimate    string   `json:"estimate" jsonschema:"how long it was expected to take, as 40h; empty when nobody said"`
}

// OfObjective turns an objective into the shape a machine reads.
func OfObjective(o Objective) Shape {
	c := o.Carried()
	labels := c.Labels
	if labels == nil {
		labels = []string{}
	}
	return Shape{
		Number: c.Number, Title: c.Title, State: string(c.State),
		Priority: string(c.Priority), Description: c.Description, Labels: labels,
		Deadline: domain.FormatDay(c.Deadline),
		Estimate: domain.FormatDuration(c.Estimate),
	}
}
