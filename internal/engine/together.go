package engine

// Several of a thing taken in one act. Deleting three tasks one call at a time
// left, when the third was refused, two gone and somebody to work out which;
// named together, they are taken in the one transaction an act runs in, every
// one of them or none.

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// alsoOf is the argument that names more of a thing, taken in the same act as
// the one the number names.
func alsoOf(of string) Argument {
	return Argument{Name: "also", Kind: List, Of: of,
		Means: "another one taken in the same act as the number — every one of them, or none"}
}

// together is the number named and every one named beside it, in the order
// they were given and each once.
func together(in Input) ([]int, error) {
	first, _ := in.Whole("number")
	numbers := []int{first}
	for _, raw := range in.Listed["also"] {
		more, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return nil, Misread(fmt.Errorf("%q: a number is expected beside the first", raw))
		}
		if !slices.Contains(numbers, more) {
			numbers = append(numbers, more)
		}
	}
	return numbers, nil
}
