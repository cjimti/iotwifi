// IoT Wifi packages is used to manage WiFi AP and Station (client) modes on
// a Raspberry Pi or other arm device. This code is intended to run in it's
// corresponding Alpine docker container.

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

// CmdOut structures command output
type CmdOut struct {
	Id      string
	Command string
	Message string
	Error   bool
}

// RunWifi starts AP and Station
func RunWifi(log bunyan.Logger, messages chan CmdOut) {

	log.Info("Loading IoT Wifi...")

	cmdRunner := CmdRunner{
		Log:      log,
		Messages: messages,
	}

	cmd := exec.Command("ifconfig", "uap0")
	go cmdRunner.ProcessCmd("myping", *cmd)

	staticFields := make(map[string]interface{})

	// command output loop
	//
	for {
		out := <-messages // Block until we receive a message on the channel

		if out.Command == "fun" {
			log.Info("GOT FUN!!!!")
		}

		staticFields["cmd_id"] = out.Id
		staticFields["cmd"] = out.Command
		staticFields["is_error"] = out.Error

		log.Info(staticFields, out.Message)
	}
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
