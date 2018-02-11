// IoT Wifi Management

package main

import (
	"io"
	"net/http"
	"os"

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

	messages := make(chan iotwifi.CmdMessage, 1)

	go iotwifi.RunWifi(bunyanLogger, messages)

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		messages <- iotwifi.CmdMessage{Id: "kill"}
		io.WriteString(w, "OK\n")
	})

	http.ListenAndServe(":8080", nil)
}
