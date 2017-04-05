package reorder

import (
	"fmt"
)

func Example() {
	// Create a shuffled slice of numbers
	values := []int{8, 2, 18, 0, 5, 7, 1, 16, 13, 4, 9, 12, 14, 10, 19, 11, 6, 3, 17, 15}
	fmt.Printf("Source values: %#v\n", values)

	w := WriterFunc(func(n int, buf []interface{}) error {
		fmt.Printf("Write n=%d  buf=%#v\n", n, buf)
		return nil
	})
	buf := NewBuffer(0, w)

	// Add the values to the buffer as they come.
	// we're using "n" for "item" here, but item could actually be anything.
	// The writer will be called as we add values and contiguous, ordered groups
	// become available.
	for _, n := range values {
		buf.Add(n, n)
	}

	// Output:
	// Source values: []int{8, 2, 18, 0, 5, 7, 1, 16, 13, 4, 9, 12, 14, 10, 19, 11, 6, 3, 17, 15}
	// Write n=0  buf=[]interface {}{0}
	// Write n=1  buf=[]interface {}{1, 2}
	// Write n=3  buf=[]interface {}{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	// Write n=15  buf=[]interface {}{15, 16, 17, 18, 19}
}
