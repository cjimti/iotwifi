package iotwifi

import (
	"os/exec"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

type SetupCfg struct {
	DnsmasqCfg       DnsmasqCfg       `json:"dnsmasq_cfg"`
	HostApdCfg       HostApdCfg       `json:"host_apd_cfg"`
	WpaSupplicantCfg WpaSupplicantCfg `json:"wpa_supplicant_cfg"`
}

type DnsmasqCfg struct {
	Address     string `json:"address"`      // --address=/#/192.168.27.1",
	DhcpRange   string `json:"dhcp_range"`   // "--dhcp-range=192.168.27.100,192.168.27.150,1h",
	VendorClass string `json:"vendor_class"` // "--dhcp-vendorclass=set:device,IoT",
}

type HostApdCfg struct {
	Ssid          string `json:"ssid"`           // ssid=iotwifi2
	WpaPassphrase string `json:"wpa_passphrase"` // wpa_passphrase=iotwifipass
	Channel       string `json:"channel"`        //  channel=6
	Ip            string `json:"ip"`             // 192.168.27.1
}

type WpaSupplicantCfg struct {
	CfgFile string `json:"cfg_file"` // /etc/wpa_supplicant/wpa_supplicant.conf
}

// Command for device network commands
type Command struct {
	Log      bunyan.Logger
	Runner   CmdRunner
	SetupCfg *SetupCfg
}

// RemoveApInterface
func (c *Command) RemoveApInterface() {
	cmd := exec.Command("iw", "dev", "uap0", "del")
	cmd.Start()
	cmd.Wait()
}

// ConfigureApInterface
func (c *Command) ConfigureApInterface() {
	cmd := exec.Command("ifconfig", "uap0", c.SetupCfg.HostApdCfg.Ip)
	cmd.Start()
	cmd.Wait()
}

// UpApInterface
func (c *Command) UpApInterface() {
	cmd := exec.Command("ifconfig", "uap0", "up")
	cmd.Start()
	cmd.Wait()
}

// AddInterface
func (c *Command) AddApInterface() {
	cmd := exec.Command("iw", "phy", "phy0", "interface", "add", "uap0", "type", "__ap")
	cmd.Start()
	cmd.Wait()
}

// CheckInterface
func (c *Command) CheckApInterface() {
	cmd := exec.Command("ifconfig", "uap0")
	go c.Runner.ProcessCmd("ifconfig_uap0", cmd)
}

// StartWpaSupplicant
func (c *Command) StartWpaSupplicant() {
	args := []string{
		"-d",
		"-Dnl80211",
		"-iwlan0",
		"-c" + c.SetupCfg.WpaSupplicantCfg.CfgFile,
	}

	cmd := exec.Command("wpa_supplicant", args...)
	go c.Runner.ProcessCmd("wpa_supplicant", cmd)
}

// StartDnsmasq
func (c *Command) StartDnsmasq() {
	// hostapd is enabled, fire up dnsmasq
	args := []string{
		"--no-hosts", // Don't read the hostnames in /etc/hosts.
		"--keep-in-foreground",
		"--log-queries",
		"--no-resolv",
		"--address=" + c.SetupCfg.DnsmasqCfg.Address,
		"--dhcp-range=" + c.SetupCfg.DnsmasqCfg.DhcpRange,
		"--dhcp-vendorclass=" + c.SetupCfg.DnsmasqCfg.VendorClass,
		"--dhcp-authoritative",
		"--log-facility=-",
	}

	cmd := exec.Command("dnsmasq", args...)
	go c.Runner.ProcessCmd("dnsmasq", cmd)
}

// StartHostapd
func (c *Command) StartHostapd() {

	c.Runner.Log.Info("Starting hostapd.")

	cmd := exec.Command("hostapd", "-d", "/dev/stdin")
	hostapdPipe, _ := cmd.StdinPipe()
	c.Runner.ProcessCmd("hostapd", cmd)

	cfg := `interface=uap0
ssid=` + c.SetupCfg.HostApdCfg.Ssid + `
hw_mode=g
channel=` + c.SetupCfg.HostApdCfg.Channel + `
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=` + c.SetupCfg.HostApdCfg.WpaPassphrase + `
wpa_key_mgmt=WPA-PSK
wpa_pairwise=TKIP
rsn_pairwise=CCMP`

	hostapdPipe.Write([]byte(cfg))
	hostapdPipe.Close()

}
