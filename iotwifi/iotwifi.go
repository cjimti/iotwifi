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
	
	cmdRunner.HandleFunc("kill", func(cmsg CmdMessage) {
		log.Error("GOT KILL")
		os.Exit(1)
	})

	cmdRunner.HandleFunc("wpa_supplicant", func(cmsg CmdMessage) {
		if strings.Contains(cmsg.Message, "P2P: Update channel list") {
			//ConnectWifi(cmdRunner)
		}
	})

	// listen to hostapd and start dnsmasq
	cmdRunner.HandleFunc("hostapd", func(cmsg CmdMessage) {
		if strings.Contains(cmsg.Message, "uap0: AP-DISABLED") {
			log.Error("CANNOT START AP")
			os.Exit(1)
		}
		
		if strings.Contains(cmsg.Message, "uap0: AP-ENABLED") {
			log.Info("Hostapd enabeled.");
			StartDnsmasq(cmdRunner)
			//StartWpaSupplicant(cmdRunner)
		}
	})

	// check for the uap0 interface
	//
	cmdRunner.HandleFunc("ifconfig_uap0", func(cmsg CmdMessage) {
		var cmd *exec.Cmd
		
		if strings.Contains(cmsg.Message, "Device not found") {
			// no uap so lets create it
			log.Info("uap0 not found... starting one up.")
			cmd = exec.Command("iw", "phy", "phy0", "interface", "add", "uap0", "type", "__ap");
			go cmdRunner.ProcessCmd("iw_up_uap0", cmd)
		}

		if strings.Contains(cmsg.Message, "uap0      Link encap") {
			
			log.Info("uap0 is available")

			// up uap0
			cmd = exec.Command("ifconfig","uap0","up");
			cmdRunner.ProcessCmd("uap_0_up", cmd)

			// configure uap0
			cmd = exec.Command("ifconfig","uap0","192.168.27.1")
			cmdRunner.ProcessCmd("uap_0_configure", cmd)

			StartHostapd(cmdRunner)

		}
	})
	
	go cmdRunner.ProcessCmd("ifconfig_uap0", exec.Command("ifconfig", "uap0"))

	// staticFields for logger
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
	cmd.Process.Wait()
	
	if err != nil {
		panic(err)
	}
}
