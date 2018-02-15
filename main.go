// IoT Wifi Management

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/bhoriuchi/go-bunyan/bunyan"
	"github.com/cjimti/iotwifi/iotwifi"
	"github.com/gorilla/mux"
)

type ApiReturn struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
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

	cfgUrl := setEnvIfEmpty("IOTWIFI_CFG", "cfg/wificfg.json")
	port := setEnvIfEmpty("IOTWIFI_PORT", "8080")

	go iotwifi.RunWifi(blog, messages, cfgUrl)
	wpacfg := iotwifi.NewWpaCfg(blog, cfgUrl)

	apiPayloadReturn := func(w http.ResponseWriter, message string, payload interface{}) {
		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: message,
			Payload: payload,
		}
		ret, _ := json.Marshal(apiReturn)

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// marshallPost populates a struct with json in post body
	marshallPost := func(w http.ResponseWriter, r *http.Request, v interface{}) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			blog.Error(err)
			return
		}

		defer r.Body.Close()

		decoder := json.NewDecoder(strings.NewReader(string(bytes)))

		err = decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			blog.Error(err)
			return
		}
	}

	// common error return from api
	retError := func(w http.ResponseWriter, err error) {
		apiReturn := &ApiReturn{
			Status:  "FAIL",
			Message: err.Error(),
		}
		ret, _ := json.Marshal(apiReturn)

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// handle /status POSTs json in the form of iotwifi.WpaConnect
	statusHandler := func(w http.ResponseWriter, r *http.Request) {

		status, err := wpacfg.Status()
		if err != nil {
			blog.Error(err.Error())
			return
		}

		apiPayloadReturn(w, "status", status)
	}

	// handle /connect POSTs json in the form of iotwifi.WpaConnect
	connectHandler := func(w http.ResponseWriter, r *http.Request) {
		var creds iotwifi.WpaCredentials
		marshallPost(w, r, &creds)

		blog.Info("Connect Handler Got: ssid:|%s| psk:|%s|", creds.Ssid, creds.Psk)

		connection, err := wpacfg.ConnectNetwork(creds)
		if err != nil {
			blog.Error(err.Error())
			return
		}

		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: "Connection",
			Payload: connection,
		}

		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// scan for wifi networks
	scanHandler := func(w http.ResponseWriter, r *http.Request) {
		blog.Info("Got Scan")
		wpaNetworks, err := wpacfg.ScanNetworks()
		if err != nil {
			retError(w, err)
			return
		}

		apiReturn := &ApiReturn{
			Status:  "OK",
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
	}

	// kill the application
	killHandler := func(w http.ResponseWriter, r *http.Request) {
		messages <- iotwifi.CmdMessage{Id: "kill"}

		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: "Killing service.",
		}
		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// api headers for csx allowance
	allowHeaders := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With, Accept, Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")

			next.ServeHTTP(w, r)
		})
	}

	// common log middleware for api
	logHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			staticFields := make(map[string]interface{})
			staticFields["remote"] = r.RemoteAddr
			staticFields["method"] = r.Method
			staticFields["url"] = r.RequestURI

			blog.Info(staticFields, "HTTP")
			next.ServeHTTP(w, r)
		})
	}

	// setup router and middleware
	r := mux.NewRouter()
	r.Use(allowHeaders)
	r.Use(logHandler)

	// set app routes
	r.HandleFunc("/status", statusHandler)
	r.HandleFunc("/connect", connectHandler)
	r.HandleFunc("/scan", scanHandler)
	r.HandleFunc("/kill", killHandler)
	http.Handle("/", r)

	// serve http
	blog.Info("HTTP Listening on " + port)
	http.ListenAndServe(":" + port, nil)

}

// getEnv gets an environment variable or sets a default if
// one does not exist.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

// setEnvIfEmp<ty sets an environment variable to itself or
// fallback if empty.
func setEnvIfEmpty(env string, fallback string) (envVal string) {
	envVal = getEnv(env, fallback)
	os.Setenv(env, envVal)

	return envVal
}
