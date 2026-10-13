package engine

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func alsoOf(of string) Argument {
	return Argument{Name: "also", Kind: List, Of: of,
		Means: "another one taken in the same act as the number — every one of them, or none"}
}

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
