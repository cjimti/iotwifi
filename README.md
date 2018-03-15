# IOT Wifi (Raspberry Pi AP + Client)

![Raspberry Pi AP + Client](/doc_assets/pi.jpg)

TL;DR? If you are not interested in reading all this you can skip ahead to
[Getting Started](#getting-started).

IOT Wifi is a [Raspberry Pi] wifi management REST service written in [Go] and
intended to run in a [Docker] container on a Raspberry Pi.

IOT Wifi sets up network interfaces, runs [hostapd], [wpa_supplicant] and
[dnsmasq] to run simultaneously. This allows a user (or another service) to
connect to the Raspberry Pi via [hostapd]/[dnsmasq] and issue commands that
configure and connect [wpa_supplicant] to another AP. IOT Wifi then exposes
a small web server on the Pi and offers a JSON based REST API to configure Wifi.
This allows you to build a custom [Captive Portal] web page or even
programmatically connect from another device and use the exposed API to
configure the target device.

Using wifi to configure a wifi connection is a common requirement for [IOT].
As Raspberry Pis are becoming a popular choice as an [IOT] platform, helps
solve the common requirement to manage AP and Station modes.

## Background

Over a year ago I wrote a blog post called [RASPBERRY PI 3 - WIFI STATION+AP]
with my notes on setting up a **Raspberry Pi 3** to run as a [Wifi Access Point][AP] (Hostspot)
and a [Wifi Client (aka Wifi Station)][Station] simultaneously. This old blog post gets
a considerable amount of traffic, so it seems there is quite a bit of interest in
this. I have come to realize that some of the steps in my old post have changed
since newer versions of [Raspian] (n00bs build) have been released.

Since writing the post I have had a few personal and professional projects
requiring a Raspberry Pi to allow wifi setup **over wifi**. I decided to open
source this simple project to help others with similar requirements as well
as gain some feedback on where and how I can improve it. I would welcome
any contribution and credit any contributors.

## Getting Started

You will need a Raspberry Pi 3, running Raspian Stretch. You
can use the [Noobs] release to install the latest version of Raspian.

### Install Docker on Raspberry Pi

Ssh into the Pi or use the terminal application from the desktop on the Pi
to get a Bash shell.

```bash
# Docker install script
$ curl -sSL https://get.docker.com | sh
```

![Install Docker](/doc_assets/install_docker.gif)

```bash
# add pi user to Docker user group
$ sudo usermod -aG docker pi
```

![Usermod Docker](/doc_assets/usermod.gif)

Reboot the pi and test Docker.

```bash
$ sudo reboot
```

After reboot, ensure Docker is installed correctly by running a Hello World
Docker container.

```bash
# run the Docker Hello World container and remove the container
# when finished (the --rm flag)
$ docker run --rm hello-world
```

![Docker Hello World on Raspberry Pi](/doc_assets/docker-hello-world.gif)

### Pull the IOT Wifi Docker Image

You can optionally clone and build the entire project, however to get
started quickly I'll show you how to use a pre-build Docker Image. At
only 16MB this little image contains everything you need. The image
is based on [Alpine Linux] and contains [hostapd], [wpa_supplicant] and
[dnsmasq], along with a compiled wifi management utility written in go,
the souce is found in this repository: https://github.com/cjimti/iotwifi.

```bash
# Pull the IOT Wifi Docker Image
$ docker pull cjimti/iotwifi
```

![Docker IOT Wifi Image](/doc_assets/docker-pull-image.gif)

### IOT Wifi Configuration

You will need a configuration JSON file. You can download a default as
a template or just it unmodified for testing. You can mount the
configuration file into the container or specify a location with
an environment variable.

Using the default configuration file and location for testing:

```bash
# Download the default confguration file

$ wget https://raw.githubusercontent.com/cjimti/iotwifi/master/cfg/wificfg.json

```

![Download Configuration](/doc_assets/download-config.gif)

The default configuration looks like this:

```json
{
    "dnsmasq_cfg": {
      "address": "/#/192.168.27.1",
      "dhcp_range": "192.168.27.100,192.168.27.150,1h",
      "vendor_class": "set:device,IoT"
    },
    "host_apd_cfg": {
       "ip": "192.168.27.1",
       "ssid": "iot-wifi-cfg-3",
       "wpa_passphrase":"iotwifipass",
       "channel": "6"
    },
      "wpa_supplicant_cfg": {
        "cfg_file": "/etc/wpa_supplicant/wpa_supplicant.conf"
    }
}
```

You may want to change the **ssid** (AP/Hostspt Name) and the the **wpa_passphrase**
to something more appropriate to your needs. However the defaults are fine
for testing.

### Run The IOT Wifi Docker Container

The following `docker run` command will create a running Docker container from
the **[cjimti/iotwifi]** Docker image we pulled in the steps above. The container
needs to run in a **privileged mode** and have access to the **host network** (the
Raspberry Pi device) in order to configure and manage the network interfaces on
the the Raspberry Pi. We will also need to mount the configuration file.

We will run it in the foreground to observe the startup process.

```bash
$ docker run --rm --privileged --net host \
      -v $(pwd)/wificfg.json:/cfg/wificfg.json \
      cjimti/iotwifi
```

The IOT Wifi container outputs logs in the JSON format. While this makes
them a bit more difficult to read, we can feed them directly (or indirectly)
into tools like Elastic Search or other databases for further processing.

You should see some initial JSON objects with messages like `Starting IoT Wifi...`:

```json
{"hostname":"raspberrypi","level":30,"msg":"Starting IoT Wifi...","name":"iotwifi","pid":0,"time":"2018-03-15T20:19:50.374Z","v":0}
```

Keeping the current terminal open you can login on another terminal and
take a look the network interfaces on the Raspberry Pi.

```bash
# use ifconfig to view the network interfaces
$ ifconfig
```

You should see a new interface called **uap0**:

```plain
uap0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.27.1  netmask 255.255.255.0  broadcast 192.168.27.255
        inet6 fe80::6e13:d169:b00b:c946  prefixlen 64  scopeid 0x20<link>
        ether b8:27:eb:fe:c8:ab  txqueuelen 1000  (Ethernet)
        RX packets 111  bytes 8932 (8.7 KiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 182  bytes 24416 (23.8 KiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

The IOT Wifi container created.


[RASPBERRY PI 3 - WIFI STATION+AP]: http://imti.co/post/145442415333/raspberry-pi-3-wifi-station-ap
[Raspberry Pi]: https://www.raspberrypi.org/
[Raspian]: https://www.raspberrypi.org/downloads/raspbian/
[Noobs]: https://www.raspberrypi.org/downloads/noobs/
[hostapd]: https://w1.fi/hostapd/
[wpa_supplicant]: https://w1.fi/wpa_supplicant/
[dnsmasq]: http://www.thekelleys.org.uk/dnsmasq/doc.html
[Captive Portal]: https://en.wikipedia.org/wiki/Captive_portal
[AP]: https://en.wikipedia.org/wiki/Wireless_access_point
[Station]: https://en.wikipedia.org/wiki/Station_(networking)
[Go]: https://golang.org/
[IOT]: https://en.wikipedia.org/wiki/Internet_of_things
[Docker]: https://www.docker.com/
[Alpine Linux]: https://alpinelinux.org/
[cjimti/iotwifi]: https://hub.docker.com/r/cjimti/iotwifi/