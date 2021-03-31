package john

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

type Cmd struct {
	File  string
	Flags map[string]string
}

func (c *Cmd) args() []string {
	res := []string{c.File}
	for k, v := range c.Flags {
		res = append(res, fmt.Sprintf("--%v=%v", k, v))
	}
	return res
}

func (c *Cmd) showArgs() []string {
	flgs := c.args()
	return append(flgs, "--show")
}

func (c *Cmd) Run() {
	cmd := exec.Command("john", c.args()...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
}

func (c *Cmd) Results() []string {
	cmd := exec.Command("john", "dummy", "--show", "--format=raw-sha256")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(stdout)
	rgx := regexp.MustCompile(`\?:.*`)
	var result []string
	for scanner.Scan() {
		m := scanner.Text()
		for _, r := range rgx.FindAllString(m, -1) {
			log.Print(r[2:])
			result = append(result, r[2:])
		}
	}
	cmd.Wait()
	return result
}
