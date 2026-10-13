package trace

import "time"

// Shape is how an act crosses to a machine.
type Shape struct {
	Task    int    `json:"task"`
	Did     string `json:"did"`
	Field   string `json:"field,omitempty"`
	Was     string `json:"was,omitempty"`
	Became  string `json:"became,omitempty"`
	Through string `json:"through,omitempty"`
	At      string `json:"at"`
}

// OfAct turns an act into what crosses.
func OfAct(a Act) Shape {
	c := a.Carried()
	return Shape{
		Task: c.Task, Did: c.Did, Field: c.Field, Was: c.Was,
		Became: c.Became, Through: c.Through, At: c.At.UTC().Format(time.RFC3339),
	}
}
