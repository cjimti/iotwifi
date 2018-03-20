package iotwifi

import (
	"os/exec"

	"github.com/bhoriuchi/go-bunyan/bunyan"
)

// Command for device network commands.
type Command struct {
	Log      bunyan.Logger
	Runner   CmdRunner
	SetupCfg *SetupCfg
}

// RemoveApInterface removes the AP interface.
func (c *Command) RemoveApInterface() {
	cmd := exec.Command("iw", "dev", "uap0", "del")
	cmd.Start()
	cmd.Wait()
}

// ConfigureApInterface configured the AP interface.
func (c *Command) ConfigureApInterface() {
	cmd := exec.Command("ifconfig", "uap0", c.SetupCfg.HostApdCfg.Ip)
	cmd.Start()
	cmd.Wait()
}

// UpApInterface ups the AP Interface.
func (c *Command) UpApInterface() {
	cmd := exec.Command("ifconfig", "uap0", "up")
	cmd.Start()
	cmd.Wait()
}

// AddApInterface adds the AP interface.
func (c *Command) AddApInterface() {
	cmd := exec.Command("iw", "phy", "phy0", "interface", "add", "uap0", "type", "__ap")
	cmd.Start()
	cmd.Wait()
}

// CheckInterface checks the AP interface.
func (c *Command) CheckApInterface() {
	cmd := exec.Command("ifconfig", "uap0")
	go c.Runner.ProcessCmd("ifconfig_uap0", cmd)
}

// StartWpaSupplicant starts wpa_supplicant.
func (c *Command) StartWpaSupplicant() {

	args := []string{
		"-d",
		"-Dnl80211",
		"-iwlan0",
		"-c/etc/wpa_supplicant/wpa_supplicant.conf",
	}

	cmd := exec.Command("wpa_supplicant", args...)
	go c.Runner.ProcessCmd("wpa_supplicant", cmd)
}

// StartDnsmasq starts dnsmasq.
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
