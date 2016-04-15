package cmdcapture_test

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"github.com/gwatts/gomisc/cmdcapture"
)

func ExampleCapture() {
	cmd := exec.Command("/bin/bash", "-c", "echo stdout1 && sleep 1 && echo stderr1 >&2 &&  echo stdout2 && echo stderr2 >&2")
	c := cmdcapture.NewWithCmd(cmd)

	if err := cmd.Run(); err != nil {
		log.Fatal("exec failed:", err)
	}

	// Print each line, prefixing lines form stdout with "-" and lines from
	// stderr with "!"
	for _, block := range c.Blocks() {
		prefix := "-"
		if block.Type == cmdcapture.Stderr {
			prefix = "!"
		}
		s := bufio.NewScanner(bytes.NewReader(block.Data))
		for s.Scan() {
			fmt.Println(prefix, s.Text())
		}
	}
}
