// IoT Wifi Management

package main

import (
	"net/http"
	"os"
	"encoding/json"

	"github.com/bhoriuchi/go-bunyan/bunyan"
	"github.com/cjimti/iotwifi/iotwifi"
)

type ApiReturn struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Payload interface{} `json:"payload"`
}

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

	retError := func(w http.ResponseWriter, err error) {
		apiReturn := &ApiReturn{
			Status: "Faile",
			Message: err.Error(),
		}
		ret, _ := json.Marshal(apiReturn)
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)		
	}

	http.HandleFunc("/scan", func(w http.ResponseWriter, r *http.Request) {
		blog.Info("Got Scan")
		wpa := iotwifi.NewWpaCfg(blog)
		wpaNetworks, err := wpa.ScanNetworks()
		if err != nil {
			retError(w, err)
			return
		}

		apiReturn := &ApiReturn{
			Status: "OK",
			Message: "Networks",
			Payload: wpaNetworks,
		}
		
		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	})

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		messages <- iotwifi.CmdMessage{Id: "kill"}

		apiReturn := &ApiReturn{
			Status: "OK",
			Message: "Killing service.",
		}
		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	})

	blog.Info("HTTP Listening on 8080")
	http.ListenAndServe(":8080", nil)

}
