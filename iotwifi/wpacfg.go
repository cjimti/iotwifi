package iotwifi

import (
	"os/exec"
	"strings"
	"time"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// WpaCfg for configuring wpa
type WpaCfg struct {
	Log      bunyan.Logger
	WpaCmd   []string
}

type WpaNetwork struct {
	Bssid       string `json:"bssid"`
	Frequency   string `json:"frequency"`
	SignalLevel string `json:"signal_level"`
	Flags       string `json:"flags"`
	Ssid        string `json:"ssid"`
}

func NewWpaCfg(log bunyan.Logger) *WpaCfg {

	return &WpaCfg{
		Log: log,
	}
}

func (wpa *WpaCfg) ConfiguredNetworks() string {
	netOut, err := exec.Command("wpa_cli","-i","wlan0", "scan").Output()
	if err != nil {
		wpa.Log.Fatal(err)
	}

	return string(netOut)
}


func (wpa *WpaCfg) ConnectNetwork(ssid string, psk string) (status bool, err error) {

	n := wpa.ConfiguredNetworks()
	wpa.Log.Info("GOT: %s", n)
	// add a network
	//wpa_cli -i wlan0 add_networ

	/*
  ssid (network name, SSID)
  psk (WPA passphrase or pre-shared key)
  key_mgmt (key management protocol)
  identity (EAP identity)
  password (EAP password)
*/
	return true, nil
}


// ScanNetworks returns a map of WpaNetwork data structures
func (wpa *WpaCfg) ScanNetworks() (map[string]WpaNetwork, error) {
	wpaNetworks := make(map[string]WpaNetwork,0)
		
	scanOut, err := exec.Command("wpa_cli","-i","wlan0", "scan").Output()
	if err != nil {
		wpa.Log.Fatal(err)
		return wpaNetworks, err
	}
	scanOutClean := strings.TrimSpace(string(scanOut))

	// wait one second for results
	time.Sleep(1 * time.Second)

	
	if scanOutClean == "OK" {
		networkListOut, err := exec.Command("wpa_cli","-i","wlan0", "scan_results").Output()
		if err != nil {
			wpa.Log.Fatal(err)
			return wpaNetworks, err
		}

		networkListOutArr := strings.Split(string(networkListOut),"\n")
		for _, netRecord := range networkListOutArr[1:] {
			if strings.Contains(netRecord, "[P2P]") {
				continue
			}
			
			fields := strings.Fields(netRecord)
			
			if len(fields) > 4 {
				ssid := strings.Join(fields[4:],",")
				wpaNetworks[ssid] = WpaNetwork{
					Bssid: fields[0],
					Frequency: fields[1],
					SignalLevel: fields[2],
					Flags: fields[3],
					Ssid: ssid,
				}
			}
		}
		
	}

	return wpaNetworks, nil
}



