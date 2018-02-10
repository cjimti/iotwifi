// IoT Wifi Management

package main

import (
	"os"
	"os/exec"

	"github.com/bhoriuchi/go-bunyan/bunyan"
	"github.com/cjimti/iotwifi/iotwifi"
)

func main() {

	logConfig := bunyan.Config{
		Name:   "iotwifi",
		Stream: os.Stdout,
		Level:  bunyan.LogLevelDebug,
	}

	bunyanLogger, err := bunyan.CreateLogger(logConfig)
	if err != nil {
		panic(err)
	}
	bunyanLogger.Info("Loading IoT Wifi...")

	messages := make(chan iotwifi.CmdOut, 1)

	cmdRunner := iotwifi.CmdRunner{
		Log:      bunyanLogger,
		Messages: messages,
	}

	cmd := exec.Command("ifconfig", "uap0")
	go cmdRunner.ProcessCmd("myping", *cmd)

	staticFields := make(map[string]interface{})

	// command output loop
	//
	for {
		out := <-messages // Block until we receive a message on the channel

		staticFields["cmd_id"] = out.Id
		staticFields["cmd"] = out.Command
		staticFields["is_error"] = out.Error

		bunyanLogger.Info(staticFields, out.Message)
	}

}
