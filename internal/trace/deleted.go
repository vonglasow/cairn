package trace

import "strings"

// WentWithItsTime is the pair a deletion writes last among what went with its task when the task took its time with it.
const WentWithItsTime = "time: taken"

const titlePair = "title: "

func pairsWent(c Carried) []string {
	if c.Did != Deleted || c.Was == "" {
		return nil
	}
	return strings.Split(c.Was, "; ")
}

func tookItsTime(c Carried) bool {
	pairs := pairsWent(c)
	return len(pairs) > 0 && pairs[len(pairs)-1] == WentWithItsTime
}

// WithoutTimeTaken is the acts with every act of a task that was deleted taking its time with it left out, so no sheet counts that time.
func WithoutTimeTaken(acts []Act) []Act {
	taken := map[int]bool{}
	for _, a := range acts {
		if c := a.Carried(); tookItsTime(c) {
			taken[c.Task] = true
		}
	}
	if len(taken) == 0 {
		return acts
	}
	kept := make([]Act, 0, len(acts))
	for _, a := range acts {
		if !taken[a.Carried().Task] {
			kept = append(kept, a)
		}
	}
	return kept
}

// TitleWent is the title the deletion of that task took with it; nothing when it was not deleted.
func TitleWent(acts []Act, task int) string {
	for _, a := range acts {
		c := a.Carried()
		if c.Task != task {
			continue
		}
		for _, pair := range pairsWent(c) {
			if title, ok := strings.CutPrefix(pair, titlePair); ok {
				return title
			}
		}
	}
	return ""
}
