package trigger

import "github.com/shankj3/ocelot/models/pb"

type ConditionalDirective struct {
	Conditions []Section
	Logical    Conditional
}

func (c *ConditionalDirective) IsFulfilled(td *pb.ChangesetData) bool {
	for _, condition := range c.Conditions {
		if condition.PassesMuster(td) {
			// if logical is Or, the first time this condition passes return true.
			if c.Logical == Or || c.Logical == CNone {
				return true
			}
			// if the logical is And, make sure every condition passes before calling this good
			continue
		} else if c.Logical == And {
			// if logical is And, every condition has to pass muster
			return false
		}
	}
	// if slice is exhausted, logical is And, then every condition passed
	if c.Logical == And {
		return true
	}
	return false
}
