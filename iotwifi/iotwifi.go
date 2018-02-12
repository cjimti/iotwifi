// IoT Wifi packages is used to manage WiFi AP and Station (client) modes on
// a Raspberry Pi or other arm device. This code is intended to run in it's
// corresponding Alpine docker container.

package iotwifi

import (
	"bufio"
	"os/exec"
	"strings"
	"os"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// CmdRunner
type CmdRunner struct {
	Log      bunyan.Logger
	Messages chan CmdMessage
	Handlers map[string]func(CmdMessage)
	Commands  map[string]*exec.Cmd	
}

// CmdMessage structures command output
type CmdMessage struct {
	Id      string
	Command string
	Message string
	Error   bool
	Cmd     *exec.Cmd
}

// RunWifi starts AP and Station
func RunWifi(log bunyan.Logger, messages chan CmdMessage) {

	log.Info("Loading IoT Wifi...")

	cmdRunner := CmdRunner{
		Log:      log,
		Messages: messages,
		Handlers: make(map[string]func(cmsg CmdMessage), 0),
		Commands: make(map[string]*exec.Cmd, 0),
	}

	command := &Command{
		Log: log,
		Runner: cmdRunner,
	}


	// listen to kill messages
	cmdRunner.HandleFunc("kill", func(cmsg CmdMessage) {
		log.Error("GOT KILL")
		os.Exit(1)
	})

	// listen to wpa_supplicant messages
	cmdRunner.HandleFunc("wpa_supplicant", func(cmsg CmdMessage) {
		if strings.Contains(cmsg.Message, "P2P: Update channel list") {
			// @TODO scan networks
		}
	})

	// listen to hostapd and start dnsmasq
	cmdRunner.HandleFunc("hostapd", func(cmsg CmdMessage) {

		if strings.Contains(cmsg.Message, "uap0: AP-DISABLED") {
			log.Error("CANNOT START AP")
			cmsg.Cmd.Process.Kill()
			cmsg.Cmd.Wait()
			os.Exit(3)
		}
		
		if strings.Contains(cmsg.Message, "uap0: AP-ENABLED") {
			log.Info("Hostapd enabeled.");
			command.StartDnsmasq()
		}
	})

	// check for the uap0 interface
	//
	cmdRunner.HandleFunc("ifconfig_uap0", func(cmsg CmdMessage) {
		
		if strings.Contains(cmsg.Message, "Device not found") {
			// no uap so lets create it
			log.Info("uap0 not found... starting one up.")
			cmsg.Cmd.Wait()
			
			// add interface
			command.AddApInterface()
			
			// re-check
			command.CheckApInterface()
			return
		}

		if strings.Contains(cmsg.Message, "uap0      Link encap") {
			
			log.Info("uap0 is available")
			cmsg.Cmd.Wait()

			// up uap0
			command.UpApInterface()
			
			// configure uap0
			command.ConfigureApInterface()

			// start hostapd
			command.StartHostapd()
		}
	})

	// remove AP interface (if there is one) and start fresh
	command.RemoveApInterface()

	// start wpa_supplicant (wifi client)
	command.StartWpaSupplicant()

	// start ap interface, chain of events for hostapd and dnsmasq starts here
	command.CheckApInterface()

	// staticFields for logger
	staticFields := make(map[string]interface{})

	// command output loop (channel messages)
	// loop, log and dispatch to handlers
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

	// add command to the commands map TODO close the readers
	c.Commands[id] = cmd
	
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
				Cmd:     cmd,
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
				Cmd:     cmd,
			}
		}
	}()
	
	err = cmd.Start()
	
	if err != nil {
		panic(err)
	}
}
