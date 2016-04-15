package cmdcapture

import (
	"fmt"
	"os/exec"
	"reflect"
	"testing"
)

func TestWriter(t *testing.T) {
	c := New()
	w := c.outw
	w.Write([]byte("buf1"))

	if len(c.blocks) != 0 {
		t.Fatal("unexpected flush")
	}
	if s := string(w.buf); s != "buf1" {
		t.Error("Bad buffer ", s)
	}

	w.Write([]byte(" end\nnext line"))

	if len(c.blocks) != 1 {
		t.Fatal("no flush")
	}

	if bt := c.blocks[0].Type; bt != Stdout {
		t.Error("Bad block type", bt)
	}

	if s := string(c.blocks[0].Data); s != "buf1 end\n" {
		t.Errorf("Bad data %q", s)
	}

	w.flush()
	if len(c.blocks) != 2 {
		t.Fatal("no flush")
	}

	if s := string(c.blocks[1].Data); s != "next line" {
		t.Errorf("Bad data %q", s)
	}

	w.flush()
	if len(c.blocks) != 2 {
		t.Fatal("unexpected flush")
	}
}

func TestLineBuf(t *testing.T) {
	c := New()
	o, e := c.OutWriter(), c.ErrWriter()

	fmt.Fprintf(o, "stdout1\nstdout2\nstdout partial")
	fmt.Fprintf(e, "stderr1\nstderr2\nstderr partial")
	fmt.Fprintf(o, " stdout complete\nstdout next partial")

	outexp := `stdout1
stdout2
stdout partial stdout complete
`
	errexp := `stderr1
stderr2
`

	// first call to merged should return the complete lines in the order they were completed
	mergeexp := `stdout1
stdout2
stderr1
stderr2
stdout partial stdout complete
`

	if result := string(c.Stdout()); result != outexp {
		t.Errorf("stout mismatch expected=%q actual=%q", outexp, result)
	}

	if result := string(c.Stderr()); result != errexp {
		t.Errorf("sterr mismatch expected=%q actual=%q", errexp, result)
	}

	if result := string(c.Combined()); result != mergeexp {
		t.Errorf("merged mismatch expected=%q actual=%q", mergeexp, result)
	}

	c.Flush()
	outexp += "stdout next partial"
	if result := string(c.Stdout()); result != outexp {
		t.Errorf("stout mismatch expected=%q actual=%q", outexp, result)
	}

	mergeexp += `stdout next partial
stderr partial
`
	if result := string(c.Combined()); result != mergeexp {
		t.Errorf("merged mismatch expected=%q actual=%q", mergeexp, result)
	}
}

func TestBlocks(t *testing.T) {
	c := New()
	o, e := c.OutWriter(), c.ErrWriter()

	fmt.Fprintf(o, "stdout1\nstdout2\nstdout partial")
	fmt.Fprintf(e, "stderr1\nstderr2\nstderr partial")
	fmt.Fprintf(o, " stdout complete\nstdout next partial")

	blocks := c.Blocks()
	expected := []Block{
		{Type: Stdout, Data: []byte("stdout1\nstdout2\n")},
		{Type: Stderr, Data: []byte("stderr1\nstderr2\n")},
		{Type: Stdout, Data: []byte("stdout partial stdout complete\n")},
	}

	if !reflect.DeepEqual(blocks, expected) {
		for _, b := range blocks {
			fmt.Printf("%q\n", string(b.Data))
		}
		t.Error("Mismatch")
	}
}

func TestNewWithCmd(t *testing.T) {
	cmd := new(exec.Cmd)
	c := NewWithCmd(cmd)
	if cmd.Stdout != c.outw {
		t.Error("outw not set")
	}
	if cmd.Stderr != c.errw {
		t.Error("errw not set")
	}
}
