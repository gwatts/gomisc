/*
Package cmdcapture collects stdout and stderr text emitted by an
exec.Cmd retaining the approximate order in which the output was written.

This allows for stdout and stderr text to be read separately, or for a
merged view to be retrieved,  or for each newline terminated block of
text to be accessed.

*/
package cmdcapture

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

const (
	invalid = iota
	// Stdout is set as a Block type for data written to stdout
	Stdout
	// Stderr is set as a Block type for data written to Stderr
	Stderr
)

// Block holds a block of text written to either stdout or stderr
//
// Blocks are terminated by a newline in the input stream (but may contain
// multiple lines of text) unless Flush is called to collect a partial block.
type Block struct {
	Type      int // either Stdout or Stderr
	Data      []byte
	IsPartial bool
}

// Capture provides writers for stdout and stderr to use with a cmd.Cmd.
type Capture struct {
	m      sync.Mutex
	blocks []Block
	outw   *writer
	errw   *writer
}

// New creates an initialized Capture type.
func New() *Capture {
	c := new(Capture)

	c.outw = &writer{wtype: Stdout, cap: c}
	c.errw = &writer{wtype: Stderr, cap: c}
	return c
}

// NewWithCmd reates a new Capture instance and configures the supplied Cmd
// to send its stdout and stderr to be captured.
func NewWithCmd(cmd *exec.Cmd) *Capture {
	c := New()
	cmd.Stdout = c.OutWriter()
	cmd.Stderr = c.ErrWriter()
	return c
}

// OutWriter provides a writer to supply to cmd.Cmd that will capture
// stdout text.
func (c *Capture) OutWriter() io.Writer {
	return c.outw
}

// ErrWriter provides a writer to supply to cmd.Cmd that will capture
// stderr text.
func (c *Capture) ErrWriter() io.Writer {
	return c.errw
}

// Stdout returns all of the text written to stdout so far, not including
// any partial lines unless Close or Flush has been called.
func (c *Capture) Stdout() []byte {
	return c.blockData(Stdout)
}

// Stderr returns all of the text written to stderr so far, not including
// any partial lines unless Close or Flush has been called.
func (c *Capture) Stderr() []byte {
	return c.blockData(Stderr)
}

// Combined returns all of the text written to both stdout and stderr so far,
// not including  any partial lines unless Close or Flush has been called.
//
// If Flush has been called then a newline will be added to the end of each
// partial block of text.
func (c *Capture) Combined() (result []byte) {
	c.m.Lock()
	defer c.m.Unlock()
	for _, b := range c.blocks {
		result = append(result, b.Data...)
		if result[len(result)-1] != '\n' {
			result = append(result, '\n')
		}
	}
	return result
}

// Blocks returns all of the text blocks written to either stdout or stderr
// so far.
func (c *Capture) Blocks() (blocks []Block) {
	c.m.Lock()
	defer c.m.Unlock()
	blocks = make([]Block, len(c.blocks))
	for i := range c.blocks {
		blocks[i] = c.blocks[i]
	}
	return blocks
}

// Flush causes any partial lines buffered by stdout or stderr to be returned
// by the next call to Stdout, Stderr, Merged or Blocks.
func (c *Capture) Flush() {
	c.outw.flush()
	c.errw.flush()
}

func (c *Capture) blockData(wtype int) (result []byte) {
	c.m.Lock()
	defer c.m.Unlock()
	for _, b := range c.blocks {
		if b.Type == wtype {
			result = append(result, b.Data...)
		}
	}
	return result
}

func (c *Capture) addBlock(b Block) {
	c.m.Lock()
	c.blocks = append(c.blocks, b)
	c.m.Unlock()
}

// writer implements a line buffered io.Writer
type writer struct {
	m     sync.Mutex
	wtype int
	cap   *Capture
	buf   []byte
}

// Write implements io.Writer to capture output to a memory buffer.
// It is goroutine safe.
func (w *writer) Write(p []byte) (n int, err error) {
	w.m.Lock()
	w.buf = append(w.buf, p...)
	if i := bytes.LastIndexByte(w.buf, '\n'); i > -1 {
		w.cap.addBlock(Block{Type: w.wtype, Data: w.buf[0 : i+1]})
		w.buf = w.buf[i+1:]
	}
	w.m.Unlock()
	return len(p), nil
}

func (w *writer) flush() {
	w.m.Lock()
	if len(w.buf) > 0 {
		w.cap.addBlock(Block{Type: w.wtype, Data: w.buf, IsPartial: true})
		w.buf = nil
	}
	w.m.Unlock()
}

var _ = fmt.Printf
