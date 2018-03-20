package iotwifi

// SetupCfg is the main configuration structure.
type SetupCfg struct {
	DnsmasqCfg       DnsmasqCfg       `json:"dnsmasq_cfg"`
	HostApdCfg       HostApdCfg       `json:"host_apd_cfg"`
	WpaSupplicantCfg WpaSupplicantCfg `json:"wpa_supplicant_cfg"`
}

// DnsmasqCfg configures dnsmasq and is used by SetupCfg.
type DnsmasqCfg struct {
	Address     string `json:"address"`      // --address=/#/192.168.27.1",
	DhcpRange   string `json:"dhcp_range"`   // "--dhcp-range=192.168.27.100,192.168.27.150,1h",
	VendorClass string `json:"vendor_class"` // "--dhcp-vendorclass=set:device,IoT",
}

// HostApdCfg configures hostapd and is used by SetupCfg.
type HostApdCfg struct {
	Ssid          string `json:"ssid"`           // ssid=iotwifi2
	WpaPassphrase string `json:"wpa_passphrase"` // wpa_passphrase=iotwifipass
	Channel       string `json:"channel"`        //  channel=6
	Ip            string `json:"ip"`             // 192.168.27.1
}

// WpaSupplicantCfg configures wpa_supplicant and is used by SetupCfg
type WpaSupplicantCfg struct {
	CfgFile string `json:"cfg_file"` // /etc/wpa_supplicant/wpa_supplicant.conf
}
