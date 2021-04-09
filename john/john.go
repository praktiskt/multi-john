package john

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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
		Results: make(chan []string, 100),
		Log:     logger.Sugar(),
	}
}

func (c *Cmd) args() []string {
	res := []string{c.File, potFlag}
	for k, v := range c.Flags {
		res = append(res, fmt.Sprintf("--%v=%v", k, v))
	}
	return res
}

func (c *Cmd) showArgs() []string {
	return []string{c.File, "--show", potFlag}
}

func (c *Cmd) Run() {
	go c.WatchResults()
	exec.Command(c.Bin, c.args()...).Run()
}

func (c *Cmd) WatchResults() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	c.KillChan = make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					c.Results <- c.ReadPotfile()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				c.Log.Errorf("error:", err)
			}
		}
	}()

	for {
		c.Log.Infof("waiting for file to be created...")
		if err := watcher.Add(potFile); err == nil {
			c.Log.Info("found potfile")
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	<-c.KillChan
}

func (c *Cmd) ReadPotfile() []string {
	b, err := ioutil.ReadFile(potFile)
	if err != nil {
		log.Print(err)
	}

	s := string(b)
	return strings.Split(s, "\n")
}
