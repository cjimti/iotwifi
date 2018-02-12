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

	blog, err := bunyan.CreateLogger(logConfig)
	if err != nil {
		panic(err)
	}

	blog.Info("Starting IoT Wifi...")
	
	
	messages := make(chan iotwifi.CmdMessage, 1)

	go iotwifi.RunWifi(blog, messages)

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		messages <- iotwifi.CmdMessage{Id: "kill"}
		io.WriteString(w, "OK\n")
	})

	blog.Info("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)

}
