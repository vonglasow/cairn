package item

import "fmt"

// ManyOf is a count with the word it takes: one when it is 1, many otherwise.
func ManyOf(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
