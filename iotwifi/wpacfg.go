package iotwifi

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// WpaCfg for configuring wpa
type WpaCfg struct {
	Log    bunyan.Logger
	WpaCmd []string
	WpaCfg *SetupCfg
}

// WpaNetwork defines a wifi network to connect to.
type WpaNetwork struct {
	Bssid       string `json:"bssid"`
	Frequency   string `json:"frequency"`
	SignalLevel string `json:"signal_level"`
	Flags       string `json:"flags"`
	Ssid        string `json:"ssid"`
}

// WpaCredentials defines wifi network credentials.
type WpaCredentials struct {
	Ssid string `json:"ssid"`
	Psk  string `json:"psk"`
}

// WpaConnection defines a WPA connection.
type WpaConnection struct {
	Ssid    string `json:"ssid"`
	State   string `json:"state"`
	Ip      string `json:"ip"`
	Message string `json:"message"`
}

// NewWpaCfg produces WpaCfg configuration types.
func NewWpaCfg(log bunyan.Logger, cfgLocation string) *WpaCfg {

	setupCfg, err := loadCfg(cfgLocation)
	if err != nil {
		log.Error("Could not load config: %s", err.Error())
		panic(err)
	}

	return &WpaCfg{
		Log:    log,
		WpaCfg: setupCfg,
	}
}

// StartAP starts AP mode.
func (wpa *WpaCfg) StartAP() {
	wpa.Log.Info("Starting Hostapd.")

	command := &Command{
		Log:      wpa.Log,
		SetupCfg: wpa.WpaCfg,
	}

	command.RemoveApInterface()
	command.AddApInterface()
	command.UpApInterface()
	command.ConfigureApInterface()

	cmd := exec.Command("hostapd", "-d", "/dev/stdin")

	// pipes
	hostapdPipe, _ := cmd.StdinPipe()
	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	messages := make(chan string, 1)

	stdOutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for stdOutScanner.Scan() {
			wpa.Log.Info("HOSTAPD GOT: %s", stdOutScanner.Text())
			messages <- stdOutScanner.Text()
		}
	}()

	cfg := `interface=uap0
ssid=` + wpa.WpaCfg.HostApdCfg.Ssid + `
hw_mode=g
channel=` + wpa.WpaCfg.HostApdCfg.Channel + `
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=` + wpa.WpaCfg.HostApdCfg.WpaPassphrase + `
wpa_key_mgmt=WPA-PSK
wpa_pairwise=TKIP
rsn_pairwise=CCMP`

	wpa.Log.Info("Hostapd CFG: %s", cfg)
	hostapdPipe.Write([]byte(cfg))

	cmd.Start()
	hostapdPipe.Close()

	for {
		out := <-messages // Block until we receive a message on the channel
		if strings.Contains(out, "uap0: AP-DISABLED") {
			wpa.Log.Info("Hostapd DISABLED")
			//cmd.Process.Kill()
			//cmd.Wait()

			return

		}
		if strings.Contains(out, "uap0: AP-ENABLED") {
			wpa.Log.Info("Hostapd ENABLED")
			return
		}
	}
}

// ConfiguredNetworks returns a list of configured wifi networks.
func (wpa *WpaCfg) ConfiguredNetworks() string {
	netOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan").Output()
	if err != nil {
		wpa.Log.Fatal(err)
	}

	return string(netOut)
}

// ConnectNetwork connects to a wifi network
func (wpa *WpaCfg) ConnectNetwork(creds WpaCredentials) (WpaConnection, error) {
	connection := WpaConnection{}

	// 1. Add a network
	addNetOut, err := exec.Command("wpa_cli", "-i", "wlan0", "add_network").Output()
	if err != nil {
		wpa.Log.Fatal(err)
		return connection, err
	}
	net := strings.TrimSpace(string(addNetOut))
	wpa.Log.Info("WPA add network got: %s", net)

	// 2. Set the ssid for the new network
	addSsidOut, err := exec.Command("wpa_cli", "-i", "wlan0", "set_network", net, "ssid", "\""+creds.Ssid+"\"").Output()
	if err != nil {
		wpa.Log.Fatal(err)
		return connection, err
	}
	ssidStatus := strings.TrimSpace(string(addSsidOut))
	wpa.Log.Info("WPA add ssid got: %s", ssidStatus)

	// 3. Set the psk for the new network
	addPskOut, err := exec.Command("wpa_cli", "-i", "wlan0", "set_network", net, "psk", "\""+creds.Psk+"\"").Output()
	if err != nil {
		wpa.Log.Fatal(err.Error())
		return connection, err
	}
	pskStatus := strings.TrimSpace(string(addPskOut))
	wpa.Log.Info("WPA psk got: %s", pskStatus)

	// 4. Enable the new network
	enableOut, err := exec.Command("wpa_cli", "-i", "wlan0", "enable_network", net).Output()
	if err != nil {
		wpa.Log.Fatal(err.Error())
		return connection, err
	}
	enableStatus := strings.TrimSpace(string(enableOut))
	wpa.Log.Info("WPA enable got: %s", enableStatus)

	// regex for state
	rState := regexp.MustCompile("(?m)wpa_state=(.*)\n")

	// loop for status every second
	for i := 0; i < 5; i++ {
		wpa.Log.Info("WPA Checking wifi state")

		stateOut, err := exec.Command("wpa_cli", "-i", "wlan0", "status").Output()
		if err != nil {
			wpa.Log.Fatal("Got error checking state: %s", err.Error())
			return connection, err
		}
		ms := rState.FindSubmatch(stateOut)

		if len(ms) > 0 {
			state := string(ms[1])
			wpa.Log.Info("WPA Enable state: %s", state)
			// see https://developer.android.com/reference/android/net/wifi/SupplicantState.html
			if state == "COMPLETED" {
				// save the config
				saveOut, err := exec.Command("wpa_cli", "-i", "wlan0", "save_config").Output()
				if err != nil {
					wpa.Log.Fatal(err.Error())
					return connection, err
				}
				saveStatus := strings.TrimSpace(string(saveOut))
				wpa.Log.Info("WPA save got: %s", saveStatus)

				connection.Ssid = creds.Ssid
				connection.State = state

				return connection, nil
			}
		}

		time.Sleep(3 * time.Second)
	}

	connection.State = "FAIL"
	connection.Message = "Unable to connection to " + creds.Ssid
	return connection, nil
}

// Status returns the WPA wireless status.
func (wpa *WpaCfg) Status() (map[string]string, error) {
	cfgMap := make(map[string]string, 0)

	stateOut, err := exec.Command("wpa_cli", "-i", "wlan0", "status").Output()
	if err != nil {
		wpa.Log.Fatal("Got error checking state: %s", err.Error())
		return cfgMap, err
	}

	cfgMap = cfgMapper(stateOut)

	return cfgMap, nil
}

// cfgMapper takes a byte array and splits by \n and then by = and puts it all in a map.
func cfgMapper(data []byte) map[string]string {
	cfgMap := make(map[string]string, 0)

	lines := bytes.Split(data, []byte("\n"))

	for _, line := range lines {
		kv := bytes.Split(line, []byte("="))
		if len(kv) > 1 {
			cfgMap[string(kv[0])] = string(kv[1])
		}
	}

	return cfgMap
}

// ScanNetworks returns a map of WpaNetwork data structures.
func (wpa *WpaCfg) ScanNetworks() (map[string]WpaNetwork, error) {
	wpaNetworks := make(map[string]WpaNetwork, 0)

	scanOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan").Output()
	if err != nil {
		wpa.Log.Fatal(err)
		return wpaNetworks, err
	}
	scanOutClean := strings.TrimSpace(string(scanOut))

	// wait one second for results
	time.Sleep(1 * time.Second)

	if scanOutClean == "OK" {
		networkListOut, err := exec.Command("wpa_cli", "-i", "wlan0", "scan_results").Output()
		if err != nil {
			wpa.Log.Fatal(err)
			return wpaNetworks, err
		}

		networkListOutArr := strings.Split(string(networkListOut), "\n")
		for _, netRecord := range networkListOutArr[1:] {
			if strings.Contains(netRecord, "[P2P]") {
				continue
			}

			fields := strings.Fields(netRecord)

			if len(fields) > 4 {
				ssid := strings.Join(fields[4:], " ")
				wpaNetworks[ssid] = WpaNetwork{
					Bssid:       fields[0],
					Frequency:   fields[1],
					SignalLevel: fields[2],
					Flags:       fields[3],
					Ssid:        ssid,
				}
			}
		}

	}

	return wpaNetworks, nil
}
