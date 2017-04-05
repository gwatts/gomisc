package reorder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var growTests = []struct {
	input    []interface{}
	size     int
	expected []interface{}
}{
	{[]interface{}{1, 2, 3}, 3, []interface{}{1, 2, 3}},
	{[]interface{}{}, 3, []interface{}{errNilMarker, errNilMarker, errNilMarker}},
	{[]interface{}{1, 2, 3}[0:0], 3, []interface{}{errNilMarker, errNilMarker, errNilMarker}},
}

func TestGrow(t *testing.T) {
	assert := assert.New(t)
	for _, test := range growTests {
		result := grow(test.size, test.input)
		assert.Equal(test.expected, result, "size=%d input=%#v", test.size, test.input)
	}
}

func TestAddFast(t *testing.T) {
	assert := assert.New(t)
	var buf []interface{}
	var wn int
	w := WriterFunc(func(n int, b []interface{}) error { wn = n; buf = b; return nil })

	cb := NewBuffer(10, w)
	for i := 0; i < 3; i++ {
		cb.Add(i, i)
		assert.Equal([]interface{}{i}, buf, i)
		assert.Equal(i, wn)
	}

	var nilbuf []interface{}
	assert.Equal(nilbuf, cb.buf)
	assert.Equal(2, cb.n)
}

func TestFragmented(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var buf []interface{}
	var wn int
	callCount := 0

	w := WriterFunc(func(n int, b []interface{}) error {
		callCount++
		wn = n
		buf = b
		return nil
	})

	cb := NewBuffer(10, w)

	cb.Add(5, 5)
	require.Zero(callCount)
	cb.Add(4, 4)
	require.Zero(callCount)

	cb.Add(0, 0)
	require.Equal(1, callCount)
	assert.Equal([]interface{}{0}, buf, "should of flushed first entry only")
	assert.Equal(0, wn)
	//assert.Equal(4, len(cb.buf))

	cb.Add(2, 2)
	require.Equal(1, callCount)
	cb.Add(1, 1)
	require.Equal(2, callCount)
	assert.Equal([]interface{}{1, 2}, buf, "should of flushed first two entries only")
	assert.Equal(1, wn)

	// now have {gap, 4, 5}
	cb.Add(3, 3)
	require.Equal(3, callCount)
	assert.Equal([]interface{}{3, 4, 5}, buf, "should of flushed final entries")
	assert.Equal(3, wn)
	assert.Equal(0, len(cb.buf), "buffer should be zero length")

	// make sparse again
	cb.Add(7, 7)
	require.Equal(3, callCount)
	cb.Add(6, 6)
	require.Equal(4, callCount)
	assert.Equal([]interface{}{6, 7}, buf, "should of flushed final entries")
	assert.Equal(6, wn)
}

func TestFlushErrorFast(t *testing.T) {
	assert := assert.New(t)

	err := errors.New("fail")
	var buf []interface{}
	c := 0
	w := WriterFunc(func(n int, b []interface{}) error {
		c++
		if c == 1 {
			return err
		}
		buf = b
		return nil
	})

	cb := NewBuffer(0, w)
	assert.Equal(err, cb.Add(0, 0))
	assert.Nil(cb.Add(1, 1))
	assert.Equal([]interface{}{0, 1}, buf)

	// add a gap
	assert.Nil(cb.Add(3, 3))
	c = 0 // make f return an error on upcoming flush
	assert.Equal(err, cb.Add(2, 2))
	assert.Nil(cb.Add(4, 4))
	assert.Equal([]interface{}{2, 3, 4}, buf)
}

func TestMaxSize(t *testing.T) {
	assert := assert.New(t)
	w := WriterFunc(func(n int, b []interface{}) error { return nil })

	cb := NewBuffer(2, w)
	assert.Nil(cb.Add(1, 1))   // gap; size is now 2
	assert.Error(cb.Add(2, 2)) // attempt to make size=3 should fail
}

func TestIllegalN(t *testing.T) {
	assert := assert.New(t)
	w := WriterFunc(func(n int, b []interface{}) error { return nil })

	cb := NewBuffer(0, w)

	assert.Nil(cb.Add(1, 1))
	assert.Nil(cb.Add(0, 0)) // flushes 0, 1 - min n is now 2
	assert.Panics(func() { cb.Add(1, 1) })
}
