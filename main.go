package main

import (
	"fmt"
	"time"
)

const (
	TimeStampFormat = "2006-01-02 15:04:05"
)

func main() {
	mainLoop(NewMainLoopConf())
}

type MainLoopConf struct {
	IsDebug      bool
	SleepSeconds int
}

func mainLoop(c MainLoopConf) {
	if !c.validate() {
		return
	}
	ln := 0
	for true {
		fmt.Printf("[%s] Loop %d\n", time.Now().Format(TimeStampFormat), ln)
		ln++
		time.Sleep(time.Duration(c.SleepSeconds) * time.Second)
	}
}

func NewMainLoopConf() MainLoopConf {
	return MainLoopConf{
		IsDebug:      true,
		SleepSeconds: 3,
	}
}
func (c MainLoopConf) validate() bool {
	if c.SleepSeconds < 0 {
		return false
	}
	return true
}
