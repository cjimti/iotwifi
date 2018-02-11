// IoT Wifi packages is used to manage WiFi AP and Station (client) modes on
// a Raspberry Pi or other arm device. This code is intended to run in it's
// corresponding Alpine docker container.

package iotwifi

import (
	"bufio"
	"os/exec"
	"strings"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// CmdRunner
type CmdRunner struct {
	Log      bunyan.Logger
	Messages chan CmdMessage
	Handlers map[string]func(message *CmdMessage)
}

// CmdMessage structures command output
type CmdMessage struct {
	Id      string
	Command string
	Message string
	Error   bool
}

// RunWifi starts AP and Station
func RunWifi(log bunyan.Logger, messages chan CmdMessage) {

	log.Info("Loading IoT Wifi...")

	cmdRunner := CmdRunner{
		Log:      log,
		Messages: messages,
		Handlers: make(map[string]func(message *CmdMessage), 0),
	}

	cmdRunner.HandleFunc("ifconfig_uap0", func(cmdMessage *CmdMessage) {
		log.Info("GOT------> %s", cmdMessage.Message)
		if strings.Contains(cmdMessage.Message, "uap0 does not exist") {
			log.Info("GOT no interface uap0")
		}
	})

	cmd := exec.Command("ifconfig", "uap0")
	go cmdRunner.ProcessCmd("ifconfig_uap0", *cmd)

	cmdRunner.HandleFunc("kill", func(cmdMessage *CmdMessage) {
		log.Info("got kill")
	})

	staticFields := make(map[string]interface{})

	// command output loop
	//
	for {
		out := <-messages // Block until we receive a message on the channel

		staticFields["cmd_id"] = out.Id
		staticFields["cmd"] = out.Command
		staticFields["is_error"] = out.Error

		log.Info(staticFields, out.Message)

		if handler, ok := cmdRunner.Handlers[out.Id]; ok {
			handler(&out)
		}
	}
}

// HandleFunc is a function that gets all channel messages for a command id
func (c *CmdRunner) HandleFunc(cmdId string, handler func(cmdMessage *CmdMessage)) {
	c.Handlers[cmdId] = handler
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
			c.Messages <- CmdMessage{
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
			c.Messages <- CmdMessage{
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
