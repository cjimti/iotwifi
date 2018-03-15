# IOT Wifi (Pi AP + Client)

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
with my notes on setting up a Raspberry Pi 3 to run as a [Wifi Access Point][AP] (Hostspot)
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
# install script
$ curl -sSL https://get.docker.com | sh
```

![Install Docker](/doc_assets/install_docker.gif)

```bash
# add pi user to docker user group
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