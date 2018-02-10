package iotwifi

import (
	"bufio"
	"os/exec"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// CmdRunner
type CmdRunner struct {
	Log      bunyan.Logger
	Messages chan CmdOut
}

type CmdOut struct {
	Id      string
	Command string
	Message string
	Error   bool
}

// ProcessCmd
func (c *CmdRunner) ProcessCmd(id string, cmd exec.Cmd) {
	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	cmdStderrReader, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	stdOutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for stdOutScanner.Scan() {
			c.Messages <- CmdOut{
				Id:      id,
				Command: cmd.Path,
				Message: stdOutScanner.Text(),
				Error:   false,
			}
		}
	}()

	stdErrScanner := bufio.NewScanner(cmdStderrReader)
	go func() {
		for stdErrScanner.Scan() {
			c.Messages <- CmdOut{
				Id:      id,
				Command: cmd.Path,
				Message: stdErrScanner.Text(),
				Error:   true,
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
}
