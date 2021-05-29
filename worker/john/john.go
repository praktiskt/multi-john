package john

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

var (
	potFile = "some" + ".pot"
	potFlag = "--pot=" + potFile
)

type Cmd struct {
	Bin      string
	File     string
	Flags    map[string]string
	KillChan chan bool
	Log      *zap.SugaredLogger
	Results  chan []string
}

func New(bin string, file string, flags map[string]string, logger *zap.Logger) Cmd {
	return Cmd{
		Bin:     bin,
		File:    file,
		Flags:   flags,
		Results: make(chan []string),
		Log:     logger.Sugar(),
	}
}

func (c *Cmd) args() []string {
	res := []string{c.File, potFlag}
	for k, v := range c.Flags {
		res = append(res, fmt.Sprintf("%v=%v", k, v))
		c.Log.Debug(res)
	}
	return res
}

func (c *Cmd) showArgs() []string {
	return []string{c.File, "--show", potFlag}
}

func (c *Cmd) Run() error {
	os.Create(potFile)
	c.WatchPotfile()
	cmd := exec.Command(c.Bin, c.args()...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	watch := func(stdx io.ReadCloser) {
		scanner := bufio.NewScanner(stdx)
		for scanner.Scan() {
			m := scanner.Text()
			c.Log.Info(m)
		}
	}
	go watch(stderr)
	go watch(stdout)
	c.Log.Debug("starting john")
	if err := cmd.Run(); err != nil {
		return err
	}
	c.Log.Debug("finished running john")
	return nil
}
