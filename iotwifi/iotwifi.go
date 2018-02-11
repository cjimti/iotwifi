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
	Handlers map[string]func(CmdMessage)
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
		Handlers: make(map[string]func(cmsg CmdMessage), 0),
	}

	// 1 - check for the uap0 interface
	//
	cmdRunner.HandleFunc("ifconfig_uap0", func(cmsg CmdMessage) {
		var cmd *exec.Cmd
		
		if strings.Contains(cmsg.Message, "Device not found") {
			// no uap so lets up it
			log.Info("uap0 not found... starting one up.")
			cmd = exec.Command("iw", "phy", "phy0", "interface", "add", "uap0", "type", "__ap");
			go cmdRunner.ProcessCmd("iw_up_uap0", cmd)
		}

		if strings.Contains(cmsg.Message, "uap0      Link encap") {
			log.Info("uap0 is available")
			cmd = exec.Command("ifconfig","uap0","up");
			cmdRunner.ProcessCmd("uap_0_up", cmd)

			cmd = exec.Command("ifconfig","uap0","192.168.27.1")
			cmdRunner.ProcessCmd("uap_0_configure", cmd)
		}
	})


	
	go cmdRunner.ProcessCmd("ifconfig_uap0", exec.Command("ifconfig", "uap0"))

	
	cmdRunner.HandleFunc("kill", func(cmsg CmdMessage) {
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
			handler(out)
		}
	}
}

// HandleFunc is a function that gets all channel messages for a command id
func (c *CmdRunner) HandleFunc(cmdId string, handler func(cmdMessage CmdMessage)) {
	c.Handlers[cmdId] = handler
}

// ProcessCmd
func (c *CmdRunner) ProcessCmd(id string, cmd *exec.Cmd) {
	c.Log.Debug("ProcessCmd got %s", id);
	
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
